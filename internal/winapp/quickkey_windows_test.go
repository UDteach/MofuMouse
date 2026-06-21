//go:build windows

package winapp

import (
	"testing"

	"mofumouse/internal/core"
)

func TestQuickKeyChordAllowlist(t *testing.T) {
	tests := []core.QuickKeyID{
		core.QuickKeyFind,
		core.QuickKeyCopy,
		core.QuickKeySelectAll,
		core.QuickKeyNewTab,
	}
	for _, key := range tests {
		if _, ok := quickKeyChord(key); !ok {
			t.Fatalf("quick key %q should be allowed", key)
		}
	}
	if _, ok := quickKeyChord("delete"); ok {
		t.Fatal("delete quick key should not be allowed")
	}
}
