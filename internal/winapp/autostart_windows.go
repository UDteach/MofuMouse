//go:build windows

package winapp

import (
	"errors"
	"fmt"
	"strings"

	"golang.org/x/sys/windows/registry"
)

const (
	launchAtLoginRunKey    = `Software\Microsoft\Windows\CurrentVersion\Run`
	launchAtLoginValueName = "MofuMouse"
)

func SetLaunchAtLogin(enabled bool, executablePath string) error {
	return setLaunchAtLoginInKey(registry.CURRENT_USER, launchAtLoginRunKey, launchAtLoginValueName, executablePath, enabled)
}

func LaunchAtLoginEnabled(executablePath string) (bool, error) {
	return launchAtLoginEnabledInKey(registry.CURRENT_USER, launchAtLoginRunKey, launchAtLoginValueName, executablePath)
}

func launchAtLoginCommand(executablePath string) string {
	path := strings.TrimSpace(executablePath)
	path = strings.Trim(path, `"`)
	return `"` + strings.ReplaceAll(path, `"`, `\"`) + `"`
}

func setLaunchAtLoginInKey(root registry.Key, keyPath, valueName, executablePath string, enabled bool) error {
	if strings.TrimSpace(valueName) == "" {
		return errors.New("launch-at-login value name is empty")
	}

	if enabled {
		if strings.TrimSpace(executablePath) == "" {
			return errors.New("launch-at-login executable path is empty")
		}
		key, _, err := registry.CreateKey(root, keyPath, registry.SET_VALUE|registry.QUERY_VALUE)
		if err != nil {
			return fmt.Errorf("open launch-at-login key: %w", err)
		}
		defer key.Close()
		if err := key.SetStringValue(valueName, launchAtLoginCommand(executablePath)); err != nil {
			return fmt.Errorf("set launch-at-login value: %w", err)
		}
		return nil
	}

	key, err := registry.OpenKey(root, keyPath, registry.SET_VALUE|registry.QUERY_VALUE)
	if err != nil {
		if errors.Is(err, registry.ErrNotExist) {
			return nil
		}
		return fmt.Errorf("open launch-at-login key: %w", err)
	}
	defer key.Close()

	if err := key.DeleteValue(valueName); err != nil && !errors.Is(err, registry.ErrNotExist) {
		return fmt.Errorf("delete launch-at-login value: %w", err)
	}
	return nil
}

func launchAtLoginEnabledInKey(root registry.Key, keyPath, valueName, executablePath string) (bool, error) {
	key, err := registry.OpenKey(root, keyPath, registry.QUERY_VALUE)
	if err != nil {
		if errors.Is(err, registry.ErrNotExist) {
			return false, nil
		}
		return false, fmt.Errorf("open launch-at-login key: %w", err)
	}
	defer key.Close()

	value, _, err := key.GetStringValue(valueName)
	if err != nil {
		if errors.Is(err, registry.ErrNotExist) {
			return false, nil
		}
		return false, fmt.Errorf("read launch-at-login value: %w", err)
	}
	return strings.EqualFold(strings.TrimSpace(value), launchAtLoginCommand(executablePath)), nil
}
