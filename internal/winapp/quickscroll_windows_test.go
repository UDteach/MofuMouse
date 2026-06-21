//go:build windows

package winapp

import (
	"testing"

	"mofumouse/internal/core"
)

func TestQuickScrollWheelDelta(t *testing.T) {
	tests := []struct {
		name      string
		direction core.QuickScrollDirection
		lines     int
		want      int32
	}{
		{name: "up", direction: core.QuickScrollUp, lines: 2, want: 240},
		{name: "down", direction: core.QuickScrollDown, lines: 2, want: -240},
		{name: "default lines", direction: core.QuickScrollUp, lines: 0, want: int32(core.DefaultSettings().QuickScrollLines * wheelDelta)},
		{name: "clamped", direction: core.QuickScrollDown, lines: 999, want: -1200},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, ok := quickScrollWheelDelta(tt.direction, tt.lines)
			if !ok {
				t.Fatal("quick scroll direction should be allowed")
			}
			if got != tt.want {
				t.Fatalf("delta = %d, want %d", got, tt.want)
			}
		})
	}
}

func TestQuickScrollRejectsUnknownDirection(t *testing.T) {
	if _, ok := quickScrollWheelDelta("sideways", 3); ok {
		t.Fatal("sideways quick scroll should not be allowed")
	}
}
