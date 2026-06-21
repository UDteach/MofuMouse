package core

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"
)

func DefaultConfigPath() (string, error) {
	dir, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "MofuMouse", "settings.json"), nil
}

func LoadSettingsFile(path string) (Settings, error) {
	settings := DefaultSettings()
	data, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return settings, nil
		}
		return settings, err
	}
	if err := json.Unmarshal(data, &settings); err != nil {
		return settings, err
	}
	if settings.JiggleIdleSec <= 0 {
		settings.JiggleIdleSec = DefaultSettings().JiggleIdleSec
	}
	if settings.IdleActionSec <= 0 {
		settings.IdleActionSec = DefaultSettings().IdleActionSec
	}
	if settings.SpriteSizePx <= 0 {
		settings.SpriteSizePx = DefaultSettings().SpriteSizePx
	}
	settings.TargetMoveDelayMs = sanitizeTargetMoveDelayMs(settings.TargetMoveDelayMs)
	settings.QuickKeys = SanitizeQuickKeys(settings.QuickKeys)
	settings.QuickScrollLines = SanitizeQuickScrollLines(settings.QuickScrollLines)
	settings.TargetRules = SanitizeAssistRules(settings.TargetRules)
	if settings.PetName == "" {
		settings.PetName = DefaultSettings().PetName
	}
	settings.CoatColor = sanitizeCoatColor(settings.CoatColor)
	settings.CompanionOffsetX, settings.CompanionOffsetY = sanitizeCompanionOffsets(settings.CompanionOffsetX, settings.CompanionOffsetY)
	settings.UpdateRepo = strings.TrimSpace(settings.UpdateRepo)
	if settings.UpdateRepo == "" {
		settings.UpdateRepo = DefaultSettings().UpdateRepo
	}
	return settings, nil
}

func SaveSettingsFile(path string, settings Settings) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	data, err := json.MarshalIndent(settings, "", "  ")
	if err != nil {
		return err
	}
	data = append(data, '\n')
	return os.WriteFile(path, data, 0o644)
}
