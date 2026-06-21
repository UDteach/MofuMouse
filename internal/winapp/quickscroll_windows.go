//go:build windows

package winapp

import "mofumouse/internal/core"

const (
	mouseEventWheel = 0x0800
	wheelDelta      = 120
)

func RunQuickScroll(direction core.QuickScrollDirection, lines int) bool {
	delta, ok := quickScrollWheelDelta(direction, lines)
	if !ok {
		return false
	}
	procMouseEvent.Call(mouseEventWheel, 0, 0, uintptr(uint32(delta)), 0)
	return true
}

func quickScrollWheelDelta(direction core.QuickScrollDirection, lines int) (int32, bool) {
	if !core.IsAllowedQuickScrollDirection(direction) {
		return 0, false
	}
	units := int32(core.SanitizeQuickScrollLines(lines) * wheelDelta)
	if direction == core.QuickScrollDown {
		return -units, true
	}
	return units, true
}
