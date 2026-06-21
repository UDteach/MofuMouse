package core

import (
	"testing"
	"time"
)

func TestDefaultSettingsAreSafe(t *testing.T) {
	settings := DefaultSettings()

	if !settings.AssistantEnabled {
		t.Fatal("assistant should start enabled")
	}
	if settings.TargetAssist {
		t.Fatal("target assist should default to off")
	}
	if settings.TargetMove {
		t.Fatal("physical target movement should default to off")
	}
	if settings.TargetReturn {
		t.Fatal("target return should default to off")
	}
	if settings.TargetMoveDelayMs != 900 {
		t.Fatalf("target move delay = %d, want 900", settings.TargetMoveDelayMs)
	}
	if len(settings.TargetRules) != 0 {
		t.Fatalf("target rules len = %d, want 0", len(settings.TargetRules))
	}
	if settings.KeepAwakeMode != KeepAwakeOff {
		t.Fatalf("keep awake should default to off, got %q", settings.KeepAwakeMode)
	}
	if settings.JiggleMode != JiggleOff {
		t.Fatalf("jiggle should default to off, got %q", settings.JiggleMode)
	}
	if settings.EdgeWarpEnabled {
		t.Fatal("edge warp should default to off")
	}
	if settings.LaunchAtLogin {
		t.Fatal("launch at login should default to off")
	}
	if !settings.QuickKeysEnabled {
		t.Fatal("quick keys should default to on")
	}
	if len(settings.QuickKeys) == 0 {
		t.Fatal("default quick keys should not be empty")
	}
	if !settings.QuickScrollEnabled {
		t.Fatal("quick scroll should default to on")
	}
	if settings.QuickScrollLines != 3 {
		t.Fatalf("quick scroll lines = %d, want 3", settings.QuickScrollLines)
	}
	if settings.PetName != "Mofu" {
		t.Fatalf("unexpected default name: %q", settings.PetName)
	}
	if settings.SpriteSizePx != 32 {
		t.Fatalf("default sprite size = %d, want 32", settings.SpriteSizePx)
	}
	if settings.CompanionOffsetX != 0 || settings.CompanionOffsetY != 0 {
		t.Fatalf("default offset = %d,%d, want 0,0", settings.CompanionOffsetX, settings.CompanionOffsetY)
	}
	if settings.ShowPetName {
		t.Fatal("pet name should not be drawn by default")
	}
	if settings.CoatColor != CoatAgouti {
		t.Fatalf("default coat = %q, want agouti", settings.CoatColor)
	}
	if !settings.IdleActions {
		t.Fatal("idle actions should default to on")
	}
	if settings.JiggleIdleSec <= 0 {
		t.Fatal("jiggle idle seconds should be positive")
	}
}

func TestLegacyCompanionOffsetsAreCentered(t *testing.T) {
	state := NewState(Settings{
		AssistantEnabled: true,
		JiggleIdleSec:    60,
		IdleActionSec:    45,
		CompanionOffsetX: 28,
		CompanionOffsetY: 22,
		UpdateRepo:       "UDteach/MofuMouse",
	})

	if state.CompanionOffsetX != 0 || state.CompanionOffsetY != 0 {
		t.Fatalf("legacy offset = %d,%d, want 0,0", state.CompanionOffsetX, state.CompanionOffsetY)
	}
}

func TestStateToggles(t *testing.T) {
	state := NewState(DefaultSettings())

	if state.ToggleKeepAwakeOS() {
		t.Fatal("keep awake OS mode should stay off")
	}
	if state.KeepAwakeOS() {
		t.Fatal("keep awake OS mode should be off")
	}
	if state.ToggleKeepAwakeOS() {
		t.Fatal("keep awake OS mode should remain off")
	}

	if !state.ToggleJiggle1px() {
		t.Fatal("jiggle should toggle on")
	}
	if !state.Jiggle1px() {
		t.Fatal("jiggle should be on")
	}

	if !state.ToggleEdgeWarp() {
		t.Fatal("edge warp should toggle on")
	}
	if !state.EdgeWarpEnabled() {
		t.Fatal("edge warp should be on")
	}

	if state.ToggleIdleActions() {
		t.Fatal("idle actions should toggle off")
	}
	if state.IdleActionsEnabled() {
		t.Fatal("idle actions should be off")
	}

	if state.ToggleTargetAssist() {
		t.Fatal("target assist should toggle off")
	}
	if state.TargetAssistEnabled() {
		t.Fatal("target assist should be off")
	}

	if state.ToggleTargetMove() {
		t.Fatal("target move should stay off")
	}
	if state.TargetMoveEnabled() {
		t.Fatal("target move should be off")
	}

	if state.ToggleTargetReturn() {
		t.Fatal("target return should stay off")
	}
	if state.TargetReturnEnabled() {
		t.Fatal("target return should be off")
	}

	if !state.ToggleLaunchAtLogin() {
		t.Fatal("launch at login should toggle on")
	}
	if !state.LaunchAtLogin() {
		t.Fatal("launch at login should be on")
	}

	if state.ToggleQuickKeysEnabled() {
		t.Fatal("quick keys should toggle off")
	}
	if state.QuickKeysEnabled() {
		t.Fatal("quick keys should be off")
	}

	if state.ToggleQuickScrollEnabled() {
		t.Fatal("quick scroll should toggle off")
	}
	if state.QuickScrollEnabled() {
		t.Fatal("quick scroll should be off")
	}

}

func TestStateSnapshotIncludesRuntimeSettings(t *testing.T) {
	state := NewState(DefaultSettings())
	state.SetLaunchAtLogin(true)
	state.SetQuickKeysEnabled(true)
	state.SetQuickKeys([]QuickKeyID{QuickKeyNewTab, "delete", QuickKeyFind})
	state.SetQuickScrollEnabled(true)
	state.SetQuickScrollLines(999)
	state.SetJiggleIdleSec(999)
	state.SetIdleActionSec(999)
	state.SetTargetMoveDelayMs(999999)
	state.SetCompanionOffsets(999, -999)
	state.SetUpdateRepo("  owner/repo  ")

	snapshot := state.SettingsSnapshot()
	if !snapshot.LaunchAtLogin {
		t.Fatal("launch at login not included in snapshot")
	}
	if !snapshot.QuickKeysEnabled {
		t.Fatal("quick keys enabled not included in snapshot")
	}
	if len(snapshot.QuickKeys) != 2 || snapshot.QuickKeys[0] != QuickKeyNewTab || snapshot.QuickKeys[1] != QuickKeyFind {
		t.Fatalf("quick keys snapshot = %+v, want new_tab/find", snapshot.QuickKeys)
	}
	if !snapshot.QuickScrollEnabled {
		t.Fatal("quick scroll enabled not included in snapshot")
	}
	if snapshot.QuickScrollLines != 10 {
		t.Fatalf("quick scroll lines = %d, want 10", snapshot.QuickScrollLines)
	}
	if snapshot.JiggleIdleSec != 300 {
		t.Fatalf("jiggle idle = %d, want 300", snapshot.JiggleIdleSec)
	}
	if snapshot.IdleActionSec != 180 {
		t.Fatalf("idle action = %d, want 180", snapshot.IdleActionSec)
	}
	if snapshot.TargetMoveDelayMs != 5000 {
		t.Fatalf("target delay = %d, want 5000", snapshot.TargetMoveDelayMs)
	}
	if snapshot.CompanionOffsetX != 48 || snapshot.CompanionOffsetY != -48 {
		t.Fatalf("offset = %d,%d, want 48,-48", snapshot.CompanionOffsetX, snapshot.CompanionOffsetY)
	}
	if snapshot.UpdateRepo != "owner/repo" {
		t.Fatalf("update repo = %q, want owner/repo", snapshot.UpdateRepo)
	}
}

func TestApplySettingsUpdatesRuntimeState(t *testing.T) {
	state := NewState(DefaultSettings())
	state.SetIdleAction("sniff", time.Second)

	state.ApplySettings(Settings{
		PetName:            "  Piko  ",
		AssistantEnabled:   false,
		LaunchAtLogin:      true,
		QuickKeysEnabled:   false,
		QuickKeys:          []QuickKeyID{QuickKeyCopy, "danger", QuickKeyFind},
		QuickScrollEnabled: false,
		QuickScrollLines:   999,
		TargetAssist:       false,
		TargetMove:         true,
		TargetReturn:       true,
		TargetMoveDelayMs:  999999,
		TargetRules:        []AssistRule{{Enabled: true, Button: "Save", Action: AssistRulePrefer}},
		KeepAwakeMode:      KeepAwakeOS,
		JiggleMode:         Jiggle1px,
		JiggleIdleSec:      999,
		SpriteSizePx:       4,
		ShowPetName:        true,
		CoatColor:          CoatGray,
		CompanionOffsetX:   999,
		CompanionOffsetY:   -999,
		EdgeWarpEnabled:    true,
		IdleActions:        false,
		IdleActionSec:      999,
		UpdateCheck:        true,
		UpdateRepo:         "  owner/repo  ",
	})

	snapshot := state.SettingsSnapshot()
	if snapshot.PetName != "Piko" || snapshot.AssistantEnabled || !snapshot.LaunchAtLogin {
		t.Fatalf("basic settings not applied: %+v", snapshot)
	}
	if snapshot.QuickKeysEnabled || len(snapshot.QuickKeys) != 2 || snapshot.QuickKeys[0] != QuickKeyCopy || snapshot.QuickKeys[1] != QuickKeyFind {
		t.Fatalf("quick keys not sanitized/applied: %+v", snapshot.QuickKeys)
	}
	if snapshot.QuickScrollEnabled || snapshot.QuickScrollLines != 10 {
		t.Fatalf("quick scroll = enabled:%v lines:%d, want false/10", snapshot.QuickScrollEnabled, snapshot.QuickScrollLines)
	}
	if snapshot.TargetAssist || snapshot.TargetMove || snapshot.TargetReturn || snapshot.TargetMoveDelayMs != 5000 || len(snapshot.TargetRules) != 0 {
		t.Fatalf("target settings should stay disabled: %+v", snapshot)
	}
	if snapshot.KeepAwakeMode != KeepAwakeOff || snapshot.JiggleMode != Jiggle1px || snapshot.JiggleIdleSec != 300 {
		t.Fatalf("awake/jiggle not applied: %+v", snapshot)
	}
	if snapshot.SpriteSizePx != 16 || !snapshot.ShowPetName || snapshot.CoatColor != CoatGray {
		t.Fatalf("appearance not applied: %+v", snapshot)
	}
	if snapshot.CompanionOffsetX != 48 || snapshot.CompanionOffsetY != -48 {
		t.Fatalf("offset = %d,%d, want 48,-48", snapshot.CompanionOffsetX, snapshot.CompanionOffsetY)
	}
	if !snapshot.EdgeWarpEnabled || snapshot.IdleActions || snapshot.IdleActionSec != 180 {
		t.Fatalf("movement settings not applied: %+v", snapshot)
	}
	if !snapshot.UpdateCheck || snapshot.UpdateRepo != "owner/repo" {
		t.Fatalf("update settings not applied: %+v", snapshot)
	}
	if got := state.CurrentIdleAction(time.Now()); got != "sniff" {
		t.Fatalf("transient idle action was reset: %q", got)
	}
}

func TestStateTargetRulesSnapshot(t *testing.T) {
	state := NewState(DefaultSettings())
	state.SetTargetRules([]AssistRule{
		{Enabled: true, App: "sample.exe", Button: "Later", Action: AssistRulePrefer},
		{Enabled: true, Action: AssistRuleBlock},
	})

	rules := state.TargetRules()
	if len(rules) != 0 {
		t.Fatalf("target rules len = %d, want 0", len(rules))
	}

	snapshot := state.SettingsSnapshot()
	if len(snapshot.TargetRules) != 0 {
		t.Fatalf("snapshot rules len = %d, want 0", len(snapshot.TargetRules))
	}
}

func TestSpriteSizeIsClamped(t *testing.T) {
	state := NewState(DefaultSettings())

	state.SetSpriteSizePx(4)
	if got := state.SpriteSizePx(); got != 16 {
		t.Fatalf("small sprite size = %d, want 16", got)
	}

	state.SetSpriteSizePx(400)
	if got := state.SpriteSizePx(); got != 192 {
		t.Fatalf("large sprite size = %d, want 192", got)
	}
}

func TestCoatColorIsSanitized(t *testing.T) {
	state := NewState(DefaultSettings())

	state.SetCoatColor(CoatGray)
	if got := state.CoatColor(); got != CoatGray {
		t.Fatalf("coat = %q, want gray", got)
	}

	state.SetCoatColor("unknown")
	if got := state.CoatColor(); got != CoatAgouti {
		t.Fatalf("invalid coat = %q, want agouti", got)
	}
}

func TestStatusText(t *testing.T) {
	state := NewState(DefaultSettings())
	if got := state.StatusText(); got != "MofuMouse: Mofu / 見守り中" {
		t.Fatalf("unexpected status: %q", got)
	}

	state.SetKeepAwakeOS(true)
	state.SetJiggle1px(true)
	if got := state.StatusText(); got != "MofuMouse: Mofu / 1px移動中" {
		t.Fatalf("unexpected status: %q", got)
	}

	state.SetKeepAwakeOS(false)
	state.SetJiggle1px(false)
	state.SetAssistantEnabled(false)
	if got := state.StatusText(); got != "MofuMouse: Mofu / お休み中" {
		t.Fatalf("unexpected status: %q", got)
	}
}

func TestNameIsSanitized(t *testing.T) {
	state := NewState(DefaultSettings())
	state.SetName("  very-long-mofumouse-name  ")
	if got := []rune(state.Name()); len(got) != 16 {
		t.Fatalf("name length = %d, want 16: %q", len(got), string(got))
	}

	state.SetName("   ")
	if got := state.Name(); got != "Mofu" {
		t.Fatalf("blank name = %q, want default", got)
	}
}
