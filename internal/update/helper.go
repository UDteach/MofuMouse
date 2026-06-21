package update

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

func PrepareHelperExecutable(currentExePath, helperDir string) (string, error) {
	currentExePath = strings.TrimSpace(currentExePath)
	if currentExePath == "" {
		return "", fmt.Errorf("current executable path is empty")
	}
	info, err := os.Stat(currentExePath)
	if err != nil {
		return "", err
	}
	if info.IsDir() {
		return "", fmt.Errorf("current executable path is a directory")
	}
	if helperDir == "" {
		helperDir, err = DefaultHelperDir()
		if err != nil {
			return "", err
		}
	}
	if err := os.MkdirAll(helperDir, 0o755); err != nil {
		return "", err
	}
	_ = PruneHelperExecutables(helperDir, time.Now(), 7*24*time.Hour)

	helperPath := filepath.Join(helperDir, "mofumouse-update-helper-"+time.Now().UTC().Format("20060102-150405.000000000")+".exe")
	if _, err := copyFile(currentExePath, helperPath); err != nil {
		_ = os.Remove(helperPath)
		return "", err
	}
	return helperPath, nil
}

func PruneHelperExecutables(helperDir string, now time.Time, maxAge time.Duration) error {
	if maxAge <= 0 {
		return nil
	}
	entries, err := os.ReadDir(helperDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	var firstErr error
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := strings.ToLower(entry.Name())
		if !strings.HasPrefix(name, "mofumouse-update-helper-") || !strings.HasSuffix(name, ".exe") {
			continue
		}
		info, err := entry.Info()
		if err != nil {
			if firstErr == nil {
				firstErr = err
			}
			continue
		}
		if now.Sub(info.ModTime()) <= maxAge {
			continue
		}
		if err := os.Remove(filepath.Join(helperDir, entry.Name())); err != nil && firstErr == nil {
			firstErr = err
		}
	}
	return firstErr
}
