package update

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestPrepareHelperExecutableCopiesCurrentExe(t *testing.T) {
	temp := t.TempDir()
	current := filepath.Join(temp, "mofumouse-x64.exe")
	if err := os.WriteFile(current, []byte("current exe"), 0o755); err != nil {
		t.Fatal(err)
	}
	helperDir := filepath.Join(temp, "helpers")

	helper, err := PrepareHelperExecutable(current, helperDir)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.HasPrefix(filepath.Base(helper), "mofumouse-update-helper-") {
		t.Fatalf("helper name = %q", filepath.Base(helper))
	}
	data, err := os.ReadFile(helper)
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != "current exe" {
		t.Fatalf("helper content = %q", data)
	}
}

func TestPruneHelperExecutablesOnlyRemovesOldHelpers(t *testing.T) {
	temp := t.TempDir()
	oldHelper := filepath.Join(temp, "mofumouse-update-helper-old.exe")
	newHelper := filepath.Join(temp, "mofumouse-update-helper-new.exe")
	other := filepath.Join(temp, "mofumouse-x64.exe")
	for _, path := range []string{oldHelper, newHelper, other} {
		if err := os.WriteFile(path, []byte("x"), 0o755); err != nil {
			t.Fatal(err)
		}
	}
	now := time.Now()
	oldTime := now.Add(-8 * 24 * time.Hour)
	if err := os.Chtimes(oldHelper, oldTime, oldTime); err != nil {
		t.Fatal(err)
	}

	if err := PruneHelperExecutables(temp, now, 7*24*time.Hour); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(oldHelper); !os.IsNotExist(err) {
		t.Fatalf("old helper still exists or unexpected error: %v", err)
	}
	for _, path := range []string{newHelper, other} {
		if _, err := os.Stat(path); err != nil {
			t.Fatalf("%s should remain: %v", path, err)
		}
	}
}
