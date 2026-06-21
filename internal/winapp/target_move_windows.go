//go:build windows

package winapp

import (
	"context"
	"fmt"
	"os"
	"time"

	"mofumouse/internal/core"
)

const (
	targetMoveTick       = 90 * time.Millisecond
	targetReturnDelay    = 850 * time.Millisecond
	targetMoveCooldown   = 2 * time.Second
	targetCancelCooldown = 1200 * time.Millisecond
	targetMoveTolerance  = int32(4)
	targetMoveMinPixels  = int32(10)
)

func RunTargetMover(ctx context.Context, state *core.State) {
	ticker := time.NewTicker(targetMoveTick)
	defer ticker.Stop()

	var last point
	var stableSince time.Time
	var hasLast bool
	var cooldownUntil time.Time

	var returnPending bool
	var returnAt time.Time
	var origin point
	var target point

	for {
		select {
		case <-ctx.Done():
			return
		case now := <-ticker.C:
			pos, ok := CursorPosition()
			if !ok {
				continue
			}

			if returnPending {
				if !pointsNear(pos, target, targetMoveTolerance) {
					returnPending = false
					cooldownUntil = now.Add(targetCancelCooldown)
					hasLast = false
					continue
				}
				if now.Before(returnAt) {
					continue
				}
				state.SetIdleAction("return", 700*time.Millisecond)
				setCursorPosition(origin)
				returnPending = false
				cooldownUntil = now.Add(targetMoveCooldown)
				hasLast = false
				continue
			}

			if !state.AssistantEnabled() || !state.TargetAssistEnabled() || !state.TargetMoveEnabled() {
				debugTargetMove("disabled assistant=%v assist=%v move=%v", state.AssistantEnabled(), state.TargetAssistEnabled(), state.TargetMoveEnabled())
				hasLast = false
				continue
			}
			if now.Before(cooldownUntil) {
				debugTargetMove("cooldown until=%s", cooldownUntil.Format(time.RFC3339Nano))
				continue
			}

			if !hasLast || !pointsNear(pos, last, 1) {
				debugTargetMove("stable reset pos=%d,%d hasLast=%v last=%d,%d", pos.X, pos.Y, hasLast, last.X, last.Y)
				last = pos
				stableSince = now
				hasLast = true
				continue
			}
			last = pos

			delay := time.Duration(state.TargetMoveDelayMs) * time.Millisecond
			if now.Sub(stableSince) < delay || IdleDuration() < delay {
				debugTargetMove("waiting stable=%s idle=%s delay=%s", now.Sub(stableSince), IdleDuration(), delay)
				continue
			}

			targets := SafeForegroundAssistTargetsWithRules(state.TargetRules())
			if len(targets) == 0 {
				debugTargetMove("no safe targets rules=%d", len(state.TargetRules()))
				continue
			}
			next := targets[0].Point
			debugTargetMove("target label=%q point=%d,%d rules=%d", targets[0].Label, next.X, next.Y, len(state.TargetRules()))
			if !pointInsideVirtualScreen(next) || pointsNear(pos, next, targetMoveMinPixels) {
				debugTargetMove("target skipped inside=%v near=%v pos=%d,%d", pointInsideVirtualScreen(next), pointsNear(pos, next, targetMoveMinPixels), pos.X, pos.Y)
				cooldownUntil = now.Add(targetCancelCooldown)
				continue
			}

			origin = pos
			target = next
			if !setCursorPosition(target) {
				debugTargetMove("set cursor failed target=%d,%d", target.X, target.Y)
				cooldownUntil = now.Add(targetCancelCooldown)
				continue
			}
			debugTargetMove("moved origin=%d,%d target=%d,%d", origin.X, origin.Y, target.X, target.Y)
			hasLast = false
			if state.TargetReturnEnabled() {
				returnPending = true
				returnAt = now.Add(targetReturnDelay)
			} else {
				cooldownUntil = now.Add(targetMoveCooldown)
			}
		}
	}
}

func debugTargetMove(format string, args ...any) {
	path := os.Getenv("MOFUMOUSE_TARGET_MOVE_LOG")
	if path == "" {
		return
	}
	file, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644)
	if err != nil {
		return
	}
	defer file.Close()
	fmt.Fprintf(file, "%s ", time.Now().Format(time.RFC3339Nano))
	fmt.Fprintf(file, format, args...)
	fmt.Fprintln(file)
}

func setCursorPosition(p point) bool {
	ret, _, _ := procSetCursorPos.Call(uintptr(int(p.X)), uintptr(int(p.Y)))
	return ret != 0
}

func pointInsideVirtualScreen(p point) bool {
	x, y, w, h := virtualScreen()
	return pointInsideRect(p, x, y, w, h)
}

func pointInsideRect(p point, x, y, w, h int32) bool {
	if w <= 0 || h <= 0 {
		return false
	}
	return p.X >= x && p.Y >= y && p.X < x+w && p.Y < y+h
}

func pointsNear(a, b point, tolerance int32) bool {
	return abs32(a.X-b.X) <= tolerance && abs32(a.Y-b.Y) <= tolerance
}
