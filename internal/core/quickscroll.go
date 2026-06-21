package core

import "strings"

type QuickScrollDirection string

const (
	QuickScrollUp   QuickScrollDirection = "up"
	QuickScrollDown QuickScrollDirection = "down"
)

func SanitizeQuickScrollLines(lines int) int {
	if lines <= 0 {
		return DefaultSettings().QuickScrollLines
	}
	return clampInt(lines, 1, 10)
}

func IsAllowedQuickScrollDirection(direction QuickScrollDirection) bool {
	switch QuickScrollDirection(strings.ToLower(strings.TrimSpace(string(direction)))) {
	case QuickScrollUp, QuickScrollDown:
		return true
	default:
		return false
	}
}

func QuickScrollLabel(direction QuickScrollDirection) string {
	switch direction {
	case QuickScrollUp:
		return "Scroll up"
	case QuickScrollDown:
		return "Scroll down"
	default:
		return ""
	}
}
