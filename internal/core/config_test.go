package core

import (
	"path/filepath"
	"testing"
)

func TestSettingsRoundTrip(t *testing.T) {
	path := filepath.Join(t.TempDir(), "settings.json")
	settings := DefaultSettings()
	settings.PetName = "Moka"
	settings.AssistantEnabled = false
	settings.LaunchAtLogin = true
	settings.QuickKeysEnabled = true
	settings.QuickKeys = []QuickKeyID{QuickKeyCopy, QuickKeyFind}
	settings.QuickScrollEnabled = true
	settings.QuickScrollLines = 4
	settings.TargetAssist = false
	settings.TargetMove = true
	settings.TargetReturn = true
	settings.TargetMoveDelayMs = 1200
	settings.TargetRules = []AssistRule{
		{Enabled: true, App: "sample.exe", Window: "Dialog", Button: "Later", Action: AssistRulePrefer},
	}
	settings.EdgeWarpEnabled = true
	settings.KeepAwakeMode = KeepAwakeOS
	settings.SpriteSizePx = 28
	settings.ShowPetName = true
	settings.CoatColor = CoatGray
	settings.UpdateCheck = true

	if err := SaveSettingsFile(path, settings); err != nil {
		t.Fatalf("save settings: %v", err)
	}
	got, err := LoadSettingsFile(path)
	if err != nil {
		t.Fatalf("load settings: %v", err)
	}
	if got.AssistantEnabled {
		t.Fatal("assistant setting did not persist")
	}
	if !got.LaunchAtLogin {
		t.Fatal("launch at login setting did not persist")
	}
	if !got.QuickKeysEnabled {
		t.Fatal("quick key setting did not persist")
	}
	if len(got.QuickKeys) != 2 || got.QuickKeys[0] != QuickKeyCopy || got.QuickKeys[1] != QuickKeyFind {
		t.Fatalf("quick keys did not persist: %+v", got.QuickKeys)
	}
	if !got.QuickScrollEnabled {
		t.Fatal("quick scroll setting did not persist")
	}
	if got.QuickScrollLines != 4 {
		t.Fatalf("quick scroll lines = %d, want 4", got.QuickScrollLines)
	}
	if got.TargetAssist {
		t.Fatal("target assist setting did not persist")
	}
	if !got.TargetMove {
		t.Fatal("target move setting did not persist")
	}
	if !got.TargetReturn {
		t.Fatal("target return setting did not persist")
	}
	if got.TargetMoveDelayMs != 1200 {
		t.Fatalf("target move delay = %d, want 1200", got.TargetMoveDelayMs)
	}
	if len(got.TargetRules) != 1 || got.TargetRules[0].Button != "Later" || got.TargetRules[0].Action != AssistRulePrefer {
		t.Fatalf("target rules did not persist: %+v", got.TargetRules)
	}
	if !got.EdgeWarpEnabled {
		t.Fatal("edge warp setting did not persist")
	}
	if got.KeepAwakeMode != KeepAwakeOS {
		t.Fatalf("keep awake setting did not persist: %q", got.KeepAwakeMode)
	}
	if got.PetName != "Moka" {
		t.Fatalf("pet name setting did not persist: %q", got.PetName)
	}
	if got.SpriteSizePx != 28 {
		t.Fatalf("sprite size setting did not persist: %d", got.SpriteSizePx)
	}
	if !got.ShowPetName {
		t.Fatal("show pet name setting did not persist")
	}
	if got.CoatColor != CoatGray {
		t.Fatalf("coat color setting did not persist: %q", got.CoatColor)
	}
	if !got.UpdateCheck {
		t.Fatal("update check setting did not persist")
	}
}

func TestLoadSettingsSanitizesTargetMoveDelay(t *testing.T) {
	path := filepath.Join(t.TempDir(), "settings.json")
	settings := DefaultSettings()
	settings.TargetMoveDelayMs = 999999
	if err := SaveSettingsFile(path, settings); err != nil {
		t.Fatalf("save settings: %v", err)
	}
	got, err := LoadSettingsFile(path)
	if err != nil {
		t.Fatalf("load settings: %v", err)
	}
	if got.TargetMoveDelayMs != 5000 {
		t.Fatalf("target move delay = %d, want 5000", got.TargetMoveDelayMs)
	}
}

func TestLoadSettingsSanitizesCoatColor(t *testing.T) {
	path := filepath.Join(t.TempDir(), "settings.json")
	if err := SaveSettingsFile(path, Settings{PetName: "Mofu", CoatColor: "blue"}); err != nil {
		t.Fatalf("save settings: %v", err)
	}
	got, err := LoadSettingsFile(path)
	if err != nil {
		t.Fatalf("load settings: %v", err)
	}
	if got.CoatColor != CoatAgouti {
		t.Fatalf("invalid coat color = %q, want agouti", got.CoatColor)
	}
}

func TestLoadSettingsSanitizesTargetRules(t *testing.T) {
	path := filepath.Join(t.TempDir(), "settings.json")
	settings := DefaultSettings()
	settings.TargetRules = []AssistRule{
		{Enabled: true, Action: AssistRuleBlock},
		{Enabled: true, Button: "  Save  ", Action: "unknown"},
	}
	if err := SaveSettingsFile(path, settings); err != nil {
		t.Fatalf("save settings: %v", err)
	}
	got, err := LoadSettingsFile(path)
	if err != nil {
		t.Fatalf("load settings: %v", err)
	}
	if len(got.TargetRules) != 1 {
		t.Fatalf("target rules len = %d, want 1", len(got.TargetRules))
	}
	if got.TargetRules[0].Button != "Save" {
		t.Fatalf("target rule button = %q, want Save", got.TargetRules[0].Button)
	}
	if got.TargetRules[0].Action != AssistRulePrefer {
		t.Fatalf("target rule action = %q, want prefer", got.TargetRules[0].Action)
	}
}
