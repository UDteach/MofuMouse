//go:build windows

package winapp

import (
	"fmt"
	"testing"
	"time"

	"golang.org/x/sys/windows/registry"
)

func TestLaunchAtLoginCommandQuotesExecutablePath(t *testing.T) {
	got := launchAtLoginCommand(` C:\Program Files\MofuMouse\mofumouse-x64.exe `)
	want := `"C:\Program Files\MofuMouse\mofumouse-x64.exe"`
	if got != want {
		t.Fatalf("command = %q, want %q", got, want)
	}
}

func TestLaunchAtLoginUsesIsolatedRegistryKey(t *testing.T) {
	keyPath := fmt.Sprintf(`Software\MofuMouse\QA\AutoStart\%d`, time.Now().UnixNano())
	valueName := "MofuMouseTest"
	exePath := `C:\Program Files\MofuMouse\mofumouse-x64.exe`
	t.Cleanup(func() {
		_ = registry.DeleteKey(registry.CURRENT_USER, keyPath)
	})

	if err := setLaunchAtLoginInKey(registry.CURRENT_USER, keyPath, valueName, exePath, true); err != nil {
		t.Fatalf("enable launch at login: %v", err)
	}

	enabled, err := launchAtLoginEnabledInKey(registry.CURRENT_USER, keyPath, valueName, exePath)
	if err != nil {
		t.Fatalf("read launch at login: %v", err)
	}
	if !enabled {
		t.Fatal("launch at login should be enabled for matching command")
	}

	enabled, err = launchAtLoginEnabledInKey(registry.CURRENT_USER, keyPath, valueName, `C:\Other\MofuMouse.exe`)
	if err != nil {
		t.Fatalf("read launch at login with other executable: %v", err)
	}
	if enabled {
		t.Fatal("launch at login should be false for a different executable")
	}

	if err := setLaunchAtLoginInKey(registry.CURRENT_USER, keyPath, valueName, exePath, false); err != nil {
		t.Fatalf("disable launch at login: %v", err)
	}
	enabled, err = launchAtLoginEnabledInKey(registry.CURRENT_USER, keyPath, valueName, exePath)
	if err != nil {
		t.Fatalf("read disabled launch at login: %v", err)
	}
	if enabled {
		t.Fatal("launch at login should be disabled after deletion")
	}
}
