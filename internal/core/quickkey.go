package core

import "strings"

type QuickKeyID string

const (
	QuickKeyFind      QuickKeyID = "find"
	QuickKeyCopy      QuickKeyID = "copy"
	QuickKeySelectAll QuickKeyID = "select_all"
	QuickKeyNewTab    QuickKeyID = "new_tab"
)

var defaultQuickKeys = []QuickKeyID{
	QuickKeyFind,
	QuickKeyCopy,
	QuickKeySelectAll,
	QuickKeyNewTab,
}

func DefaultQuickKeys() []QuickKeyID {
	return append([]QuickKeyID(nil), defaultQuickKeys...)
}

func SanitizeQuickKeys(keys []QuickKeyID) []QuickKeyID {
	seen := map[QuickKeyID]bool{}
	out := make([]QuickKeyID, 0, len(keys))
	for _, key := range keys {
		key = QuickKeyID(strings.ToLower(strings.TrimSpace(string(key))))
		if !IsAllowedQuickKey(key) || seen[key] {
			continue
		}
		seen[key] = true
		out = append(out, key)
	}
	return out
}

func IsAllowedQuickKey(key QuickKeyID) bool {
	switch key {
	case QuickKeyFind, QuickKeyCopy, QuickKeySelectAll, QuickKeyNewTab:
		return true
	default:
		return false
	}
}

func QuickKeyLabel(key QuickKeyID) string {
	switch key {
	case QuickKeyFind:
		return "検索"
	case QuickKeyCopy:
		return "コピー"
	case QuickKeySelectAll:
		return "全選択"
	case QuickKeyNewTab:
		return "新しいタブ"
	default:
		return ""
	}
}

func QuickKeyShortcut(key QuickKeyID) string {
	switch key {
	case QuickKeyFind:
		return "Ctrl+F"
	case QuickKeyCopy:
		return "Ctrl+C"
	case QuickKeySelectAll:
		return "Ctrl+A"
	case QuickKeyNewTab:
		return "Ctrl+T"
	default:
		return ""
	}
}
