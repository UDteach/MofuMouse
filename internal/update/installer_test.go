package update

import (
	"archive/zip"
	"os"
	"path/filepath"
	"testing"
)

func TestInstallDownloadedAssetReplacesExecutableAndKeepsBackup(t *testing.T) {
	dir := t.TempDir()
	current := filepath.Join(dir, "mofumouse-x64.exe")
	asset := filepath.Join(dir, "downloaded.exe")
	backupDir := filepath.Join(dir, "rollback")
	writeFile(t, current, "old exe")
	writeFile(t, asset, "new exe")

	result, err := InstallDownloadedAsset(asset, current, backupDir)
	if err != nil {
		t.Fatalf("InstallDownloadedAsset: %v", err)
	}
	if result.InstalledPath != current {
		t.Fatalf("installed path = %q, want %q", result.InstalledPath, current)
	}
	if result.Bytes != int64(len("new exe")) {
		t.Fatalf("bytes = %d", result.Bytes)
	}
	if got := readFile(t, current); got != "new exe" {
		t.Fatalf("current exe = %q, want new exe", got)
	}
	if got := readFile(t, result.BackupPath); got != "old exe" {
		t.Fatalf("backup exe = %q, want old exe", got)
	}
	if _, err := os.Stat(current + ".new"); !os.IsNotExist(err) {
		t.Fatalf("staged executable should not remain: %v", err)
	}
}

func TestRollbackInstallRestoresBackup(t *testing.T) {
	dir := t.TempDir()
	current := filepath.Join(dir, "mofumouse-x64.exe")
	asset := filepath.Join(dir, "downloaded.exe")
	writeFile(t, current, "old exe")
	writeFile(t, asset, "new exe")

	result, err := InstallDownloadedAsset(asset, current, filepath.Join(dir, "rollback"))
	if err != nil {
		t.Fatalf("InstallDownloadedAsset: %v", err)
	}
	if err := RollbackInstall(result.InstalledPath, result.BackupPath); err != nil {
		t.Fatalf("RollbackInstall: %v", err)
	}
	if got := readFile(t, current); got != "old exe" {
		t.Fatalf("current exe after rollback = %q, want old exe", got)
	}
}

func TestInstallDownloadedZipAssetUsesMainExecutable(t *testing.T) {
	dir := t.TempDir()
	current := filepath.Join(dir, "mofumouse-x64.exe")
	zipPath := filepath.Join(dir, "mofumouse-x64.zip")
	writeFile(t, current, "old exe")
	writeZip(t, zipPath, map[string]string{
		"mofumouse-control-x64.exe": "control exe",
		"nested/mofumouse-x64.exe":  "new exe from zip",
	})

	result, err := InstallDownloadedAsset(zipPath, current, filepath.Join(dir, "rollback"))
	if err != nil {
		t.Fatalf("InstallDownloadedAsset zip: %v", err)
	}
	if got := readFile(t, current); got != "new exe from zip" {
		t.Fatalf("current exe = %q, want zip exe", got)
	}
	if got := readFile(t, result.BackupPath); got != "old exe" {
		t.Fatalf("backup exe = %q, want old exe", got)
	}
}

func TestInstallDownloadedAssetRejectsEmptyExecutable(t *testing.T) {
	dir := t.TempDir()
	current := filepath.Join(dir, "mofumouse-x64.exe")
	asset := filepath.Join(dir, "downloaded.exe")
	writeFile(t, current, "old exe")
	if err := os.WriteFile(asset, nil, 0o755); err != nil {
		t.Fatalf("write empty asset: %v", err)
	}

	if _, err := InstallDownloadedAsset(asset, current, filepath.Join(dir, "rollback")); err == nil {
		t.Fatal("empty executable should be rejected")
	}
	if got := readFile(t, current); got != "old exe" {
		t.Fatalf("current exe changed after rejected asset: %q", got)
	}
}

func writeFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(content), 0o755); err != nil {
		t.Fatalf("write %s: %v", path, err)
	}
}

func readFile(t *testing.T, path string) string {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	return string(data)
}

func writeZip(t *testing.T, path string, files map[string]string) {
	t.Helper()
	file, err := os.Create(path)
	if err != nil {
		t.Fatalf("create zip: %v", err)
	}
	writer := zip.NewWriter(file)
	for name, content := range files {
		entry, err := writer.Create(name)
		if err != nil {
			t.Fatalf("create zip entry: %v", err)
		}
		if _, err := entry.Write([]byte(content)); err != nil {
			t.Fatalf("write zip entry: %v", err)
		}
	}
	if err := writer.Close(); err != nil {
		t.Fatalf("close zip writer: %v", err)
	}
	if err := file.Close(); err != nil {
		t.Fatalf("close zip file: %v", err)
	}
}
