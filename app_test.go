package main

import (
	"path/filepath"
	"testing"

	"mofumouse/internal/core"
)

func TestControlCenterSavesSettings(t *testing.T) {
	path := filepath.Join(t.TempDir(), "settings.json")
	app := NewControlCenterApp(path)

	settings := core.DefaultSettings()
	settings.PetName = "Moka"
	settings.LaunchAtLogin = true
	settings.QuickKeysEnabled = true
	settings.QuickKeys = []core.QuickKeyID{core.QuickKeyCopy, core.QuickKeyFind}
	settings.QuickScrollEnabled = true
	settings.QuickScrollLines = 5
	settings.TargetAssist = false
	settings.TargetMove = true
	settings.TargetReturn = true
	settings.TargetMoveDelayMs = 1300
	settings.SpriteSizePx = 64

	saved, err := app.SaveSettings(settings)
	if err != nil {
		t.Fatalf("SaveSettings: %v", err)
	}
	if saved.PetName != "Moka" {
		t.Fatalf("saved name = %q, want Moka", saved.PetName)
	}

	loaded, err := app.GetSettings()
	if err != nil {
		t.Fatalf("GetSettings: %v", err)
	}
	if loaded.TargetAssist {
		t.Fatal("target assist setting did not persist")
	}
	if !loaded.LaunchAtLogin {
		t.Fatal("launch at login setting did not persist")
	}
	if !loaded.QuickKeysEnabled {
		t.Fatal("quick keys setting did not persist")
	}
	if len(loaded.QuickKeys) != 2 || loaded.QuickKeys[0] != core.QuickKeyCopy || loaded.QuickKeys[1] != core.QuickKeyFind {
		t.Fatalf("quick keys = %+v, want copy/find", loaded.QuickKeys)
	}
	if !loaded.QuickScrollEnabled {
		t.Fatal("quick scroll setting did not persist")
	}
	if loaded.QuickScrollLines != 5 {
		t.Fatalf("quick scroll lines = %d, want 5", loaded.QuickScrollLines)
	}
	if loaded.TargetMove {
		t.Fatal("target move should be forced off")
	}
	if loaded.TargetReturn {
		t.Fatal("target return should be forced off")
	}
	if loaded.TargetMoveDelayMs != 1300 {
		t.Fatalf("target move delay = %d, want 1300", loaded.TargetMoveDelayMs)
	}
	if loaded.SpriteSizePx != 64 {
		t.Fatalf("sprite size = %d, want 64", loaded.SpriteSizePx)
	}
}
