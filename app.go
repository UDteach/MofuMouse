package main

import (
	"context"
	"os/exec"

	"mofumouse/internal/core"
)

type ControlCenterApp struct {
	ctx        context.Context
	configPath string
}

func NewControlCenterApp(configPath string) *ControlCenterApp {
	return &ControlCenterApp{configPath: configPath}
}

func (a *ControlCenterApp) startup(ctx context.Context) {
	a.ctx = ctx
}

func (a *ControlCenterApp) GetSettings() (core.Settings, error) {
	return core.LoadSettingsFile(a.configPath)
}

func (a *ControlCenterApp) SaveSettings(settings core.Settings) (core.Settings, error) {
	sanitized := core.NewState(settings).SettingsSnapshot()
	if err := core.SaveSettingsFile(a.configPath, sanitized); err != nil {
		return sanitized, err
	}
	return sanitized, nil
}

func (a *ControlCenterApp) ResetSettings() (core.Settings, error) {
	settings := core.DefaultSettings()
	if err := core.SaveSettingsFile(a.configPath, settings); err != nil {
		return settings, err
	}
	return settings, nil
}

func (a *ControlCenterApp) OpenConfigFile() error {
	return exec.Command("notepad.exe", a.configPath).Start()
}

func (a *ControlCenterApp) ConfigPath() string {
	return a.configPath
}
