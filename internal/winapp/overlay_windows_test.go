//go:build windows

package winapp

import (
	"testing"
	"time"

	"mofumouse/internal/core"
)

func TestCurrentSpritePrioritizesReturnAction(t *testing.T) {
	state := core.NewState(core.DefaultSettings())
	state.SetIdleAction("return", time.Second)

	returnSprite := &dibSprite{}
	overlay := &Overlay{
		state: state,
		sprites: map[string]map[int32]*dibSprite{
			spriteKey("return", core.CoatAgouti): {32: returnSprite},
			spriteKey("walk", core.CoatAgouti):   {32: &dibSprite{}},
			spriteKey("idle", core.CoatAgouti):   {32: &dibSprite{}},
			spriteKey("sniff", core.CoatAgouti):  {32: &dibSprite{}},
		},
		lastMove:    time.Now(),
		assistValid: true,
	}

	sprite, action := overlay.currentSprite(time.Now(), 32)
	if action != "return" {
		t.Fatalf("action = %q, want return", action)
	}
	if sprite == nil {
		t.Fatal("return action produced no sprite")
	}
}

func TestLoadOverlaySpritesIncludesCoatSniffRuntime(t *testing.T) {
	sprites := loadOverlaySprites()
	for _, coat := range []core.CoatColor{core.CoatGray, core.CoatDark, core.CoatCream, core.CoatWhite, core.CoatPied} {
		tiers := sprites[spriteKey("sniff", coat)]
		if tiers == nil {
			t.Fatalf("%s sniff sprite tiers were not loaded", coat)
		}
		for _, tier := range []int32{32, 48, 64, 96} {
			sprite := tiers[tier]
			if sprite == nil {
				t.Fatalf("%s sniff tier %d was not loaded", coat, tier)
			}
			if sprite.width != tier || sprite.height != tier {
				t.Fatalf("%s sniff tier %d dimensions = %dx%d, want %dx%d", coat, tier, sprite.width, sprite.height, tier, tier)
			}
		}
	}
}

func TestLoadOverlaySpritesIncludesWalk2Runtime(t *testing.T) {
	sprites := loadOverlaySprites()
	for _, coat := range []core.CoatColor{core.CoatAgouti, core.CoatGray, core.CoatDark, core.CoatCream, core.CoatWhite, core.CoatPied} {
		tiers := sprites[spriteKey("walk2", coat)]
		if tiers == nil {
			t.Fatalf("%s walk2 sprite tiers were not loaded", coat)
		}
		for _, tier := range []int32{32, 48, 64, 96} {
			sprite := tiers[tier]
			if sprite == nil {
				t.Fatalf("%s walk2 tier %d was not loaded", coat, tier)
			}
			if sprite.width != tier || sprite.height != tier {
				t.Fatalf("%s walk2 tier %d dimensions = %dx%d, want %dx%d", coat, tier, sprite.width, sprite.height, tier, tier)
			}
		}
	}
}

func TestMovingSpriteCanUseWalk2Frame(t *testing.T) {
	state := core.NewState(core.DefaultSettings())
	walk2Sprite := &dibSprite{}
	overlay := &Overlay{
		state:    state,
		lastMove: time.UnixMilli(0),
		sprites: map[string]map[int32]*dibSprite{
			spriteKey("walk2", core.CoatAgouti): {32: walk2Sprite},
			spriteKey("idle", core.CoatAgouti):  {32: &dibSprite{}},
		},
	}

	sprite, action := overlay.currentSprite(time.UnixMilli(95), 32)
	if action != "move" {
		t.Fatalf("action = %q, want move", action)
	}
	if sprite != walk2Sprite {
		t.Fatal("moving sequence did not use walk2 pose")
	}
}

func TestAssistTargetMissGraceKeepsOnlySameForeground(t *testing.T) {
	now := time.Now()
	overlay := &Overlay{
		assistValid:      true,
		assistForeground: 100,
		lastAssistSeen:   now.Add(-2 * time.Second),
	}

	if !overlay.keepAssistTargetOnMiss(now, 100) {
		t.Fatal("same foreground inside grace should keep assist target")
	}
	if overlay.keepAssistTargetOnMiss(now, 200) {
		t.Fatal("different foreground should not keep stale assist target")
	}

	overlay.lastAssistSeen = now.Add(-(assistTargetGrace + time.Millisecond))
	if overlay.keepAssistTargetOnMiss(now, 100) {
		t.Fatal("expired grace should not keep stale assist target")
	}
}

func TestCompanionOffsetYPositiveMovesUp(t *testing.T) {
	screenX, screenY, screenW, screenH := virtualScreen()
	if screenW < 240 || screenH < 240 {
		t.Skipf("virtual screen too small for unclamped offset test: %dx%d", screenW, screenH)
	}
	cursor := point{X: screenX + screenW/2, Y: screenY + screenH/2}
	baseState := core.NewState(core.DefaultSettings())
	upState := core.NewState(core.DefaultSettings())
	upState.SetCompanionOffsets(0, 12)

	baseOverlay := &Overlay{state: baseState, companionSide: companionSideRight}
	upOverlay := &Overlay{state: upState, companionSide: companionSideRight}

	_, baseY := baseOverlay.targetWindowPosition(cursor, 96, 96, 32)
	_, upY := upOverlay.targetWindowPosition(cursor, 96, 96, 32)
	if got := baseY - upY; got != 12 {
		t.Fatalf("positive Y offset delta = %.0f, want 12 upward pixels", got)
	}
}
