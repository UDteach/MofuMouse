//go:build windows

package winapp

import (
	"context"
	"time"

	"mofumouse/internal/core"
)

func RunEdgeWarp(ctx context.Context, state *core.State) {
	ticker := time.NewTicker(80 * time.Millisecond)
	defer ticker.Stop()

	var cooldownUntil time.Time
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if !state.EdgeWarpEnabled() || time.Now().Before(cooldownUntil) {
				continue
			}
			pos, ok := CursorPosition()
			if !ok {
				continue
			}
			x, y, w, h := virtualScreen()
			if w <= 4 || h <= 4 {
				continue
			}
			next := pos
			warped := false
			if pos.X <= x {
				next.X = x + w - 2
				warped = true
			} else if pos.X >= x+w-1 {
				next.X = x + 1
				warped = true
			}
			if pos.Y <= y {
				next.Y = y + h - 2
				warped = true
			} else if pos.Y >= y+h-1 {
				next.Y = y + 1
				warped = true
			}
			if warped {
				procSetCursorPos.Call(uintptr(int(next.X)), uintptr(int(next.Y)))
				cooldownUntil = time.Now().Add(350 * time.Millisecond)
			}
		}
	}
}

func virtualScreen() (x, y, w, h int32) {
	x = systemMetric(smXVirtualScreen)
	y = systemMetric(smYVirtualScreen)
	w = systemMetric(smCXVirtualScreen)
	h = systemMetric(smCYVirtualScreen)
	return x, y, w, h
}

func systemMetric(index uintptr) int32 {
	ret, _, _ := procGetSystemMetrics.Call(index)
	return int32(ret)
}
