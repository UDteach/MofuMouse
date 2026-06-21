package core

import "testing"

func TestQuickScrollLinesAreClamped(t *testing.T) {
	tests := []struct {
		name  string
		lines int
		want  int
	}{
		{name: "default", lines: 0, want: DefaultSettings().QuickScrollLines},
		{name: "minimum", lines: -5, want: DefaultSettings().QuickScrollLines},
		{name: "valid", lines: 4, want: 4},
		{name: "maximum", lines: 999, want: 10},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := SanitizeQuickScrollLines(tt.lines); got != tt.want {
				t.Fatalf("lines = %d, want %d", got, tt.want)
			}
		})
	}
}

func TestQuickScrollDirectionAllowlist(t *testing.T) {
	for _, direction := range []QuickScrollDirection{QuickScrollUp, QuickScrollDown} {
		if !IsAllowedQuickScrollDirection(direction) {
			t.Fatalf("direction %q should be allowed", direction)
		}
	}
	if IsAllowedQuickScrollDirection("sideways") {
		t.Fatal("sideways quick scroll should not be allowed")
	}
}
