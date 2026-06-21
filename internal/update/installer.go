package update

import (
	"archive/zip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type InstallResult struct {
	InstalledPath string
	BackupPath    string
	SourcePath    string
	Bytes         int64
	RolledBack    bool
}

func InstallDownloadedAsset(assetPath, currentExePath, backupDir string) (InstallResult, error) {
	assetPath = strings.TrimSpace(assetPath)
	currentExePath = strings.TrimSpace(currentExePath)
	if assetPath == "" {
		return InstallResult{}, fmt.Errorf("asset path is empty")
	}
	if currentExePath == "" {
		return InstallResult{}, fmt.Errorf("current executable path is empty")
	}
	currentInfo, err := os.Stat(currentExePath)
	if err != nil {
		return InstallResult{}, err
	}
	if currentInfo.IsDir() {
		return InstallResult{}, fmt.Errorf("current executable path is a directory")
	}
	if backupDir == "" {
		backupDir = filepath.Join(filepath.Dir(currentExePath), ".mofumouse-rollback")
	}
	if err := os.MkdirAll(backupDir, 0o755); err != nil {
		return InstallResult{}, err
	}

	sourcePath, cleanup, err := installSourceFromAsset(assetPath, filepath.Base(currentExePath))
	if err != nil {
		return InstallResult{}, err
	}
	defer cleanup()

	backupPath := uniqueBackupPath(backupDir, filepath.Base(currentExePath))
	if _, err := copyFile(currentExePath, backupPath); err != nil {
		return InstallResult{}, fmt.Errorf("backup current executable: %w", err)
	}

	tempInstall := currentExePath + ".new"
	_ = os.Remove(tempInstall)
	bytes, err := copyFile(sourcePath, tempInstall)
	if err != nil {
		_ = os.Remove(tempInstall)
		return InstallResult{InstalledPath: currentExePath, BackupPath: backupPath, SourcePath: sourcePath}, fmt.Errorf("stage update executable: %w", err)
	}
	result := InstallResult{
		InstalledPath: currentExePath,
		BackupPath:    backupPath,
		SourcePath:    sourcePath,
		Bytes:         bytes,
	}
	if err := replaceWithRollback(tempInstall, currentExePath, backupPath); err != nil {
		result.RolledBack = true
		return result, err
	}
	return result, nil
}

func RollbackInstall(installedPath, backupPath string) error {
	if strings.TrimSpace(installedPath) == "" {
		return fmt.Errorf("installed path is empty")
	}
	if strings.TrimSpace(backupPath) == "" {
		return fmt.Errorf("backup path is empty")
	}
	tempPath := installedPath + ".rollback"
	_ = os.Remove(tempPath)
	if _, err := copyFile(backupPath, tempPath); err != nil {
		_ = os.Remove(tempPath)
		return err
	}
	return replaceWithRollback(tempPath, installedPath, backupPath)
}

func installSourceFromAsset(assetPath, currentBase string) (string, func(), error) {
	ext := strings.ToLower(filepath.Ext(assetPath))
	switch ext {
	case ".exe":
		info, err := os.Stat(assetPath)
		if err != nil {
			return "", func() {}, err
		}
		if info.IsDir() || info.Size() == 0 {
			return "", func() {}, fmt.Errorf("update executable is empty or invalid")
		}
		return assetPath, func() {}, nil
	case ".zip":
		tempDir, err := os.MkdirTemp("", "mofumouse-update-*")
		if err != nil {
			return "", func() {}, err
		}
		cleanup := func() { _ = os.RemoveAll(tempDir) }
		exePath, err := extractExecutableFromZip(assetPath, currentBase, tempDir)
		if err != nil {
			cleanup()
			return "", func() {}, err
		}
		return exePath, cleanup, nil
	default:
		return "", func() {}, fmt.Errorf("unsupported update asset extension: %s", ext)
	}
}

func extractExecutableFromZip(zipPath, currentBase, destDir string) (string, error) {
	reader, err := zip.OpenReader(zipPath)
	if err != nil {
		return "", err
	}
	defer reader.Close()

	var fallback *zip.File
	for _, file := range reader.File {
		if file.FileInfo().IsDir() {
			continue
		}
		base := filepath.Base(file.Name)
		if !strings.EqualFold(filepath.Ext(base), ".exe") {
			continue
		}
		if strings.Contains(strings.ToLower(base), "control") {
			continue
		}
		if strings.EqualFold(base, currentBase) {
			return extractZipFile(file, filepath.Join(destDir, base))
		}
		if fallback == nil && strings.HasPrefix(strings.ToLower(base), "mofumouse") {
			fallback = file
		}
	}
	if fallback != nil {
		return extractZipFile(fallback, filepath.Join(destDir, filepath.Base(fallback.Name)))
	}
	return "", fmt.Errorf("zip asset does not contain a main MofuMouse executable")
}

func extractZipFile(file *zip.File, destPath string) (string, error) {
	source, err := file.Open()
	if err != nil {
		return "", err
	}
	defer source.Close()

	dest, err := os.OpenFile(destPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0o755)
	if err != nil {
		return "", err
	}
	bytes, copyErr := io.Copy(dest, source)
	closeErr := dest.Close()
	if copyErr != nil {
		_ = os.Remove(destPath)
		return "", copyErr
	}
	if closeErr != nil {
		_ = os.Remove(destPath)
		return "", closeErr
	}
	if bytes == 0 {
		_ = os.Remove(destPath)
		return "", fmt.Errorf("zip executable is empty")
	}
	return destPath, nil
}

func replaceWithRollback(stagedPath, installedPath, backupPath string) error {
	if err := os.Remove(installedPath); err != nil {
		_ = os.Remove(stagedPath)
		return fmt.Errorf("remove current executable: %w", err)
	}
	if err := os.Rename(stagedPath, installedPath); err != nil {
		_, restoreErr := copyFile(backupPath, installedPath)
		if restoreErr != nil {
			return fmt.Errorf("replace executable: %w; rollback failed: %v", err, restoreErr)
		}
		return fmt.Errorf("replace executable: %w", err)
	}
	return nil
}

func uniqueBackupPath(backupDir, currentBase string) string {
	stamp := time.Now().UTC().Format("20060102-150405.000000000")
	return filepath.Join(backupDir, currentBase+"."+stamp+".bak")
}

func copyFile(sourcePath, destPath string) (int64, error) {
	source, err := os.Open(sourcePath)
	if err != nil {
		return 0, err
	}
	defer source.Close()

	dest, err := os.OpenFile(destPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0o755)
	if err != nil {
		return 0, err
	}
	bytes, copyErr := io.Copy(dest, source)
	closeErr := dest.Close()
	if copyErr != nil {
		return bytes, copyErr
	}
	if closeErr != nil {
		return bytes, closeErr
	}
	if bytes == 0 {
		return bytes, fmt.Errorf("copied file is empty")
	}
	return bytes, nil
}
