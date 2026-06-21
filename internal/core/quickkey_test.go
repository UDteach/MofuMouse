package core

import "testing"

func TestSanitizeQuickKeys(t *testing.T) {
	got := SanitizeQuickKeys([]QuickKeyID{
		QuickKeyCopy,
		" delete ",
		QuickKeyCopy,
		"FIND",
		QuickKeyNewTab,
	})
	want := []QuickKeyID{QuickKeyCopy, QuickKeyFind, QuickKeyNewTab}
	if len(got) != len(want) {
		t.Fatalf("quick keys len = %d, want %d: %+v", len(got), len(want), got)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("quick key[%d] = %q, want %q", i, got[i], want[i])
		}
	}
}

func TestQuickKeyLabels(t *testing.T) {
	for _, key := range DefaultQuickKeys() {
		if QuickKeyLabel(key) == "" {
			t.Fatalf("missing label for %q", key)
		}
		if QuickKeyShortcut(key) == "" {
			t.Fatalf("missing shortcut for %q", key)
		}
	}
}
