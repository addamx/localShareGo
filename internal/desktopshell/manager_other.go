//go:build !windows

package desktopshell

import (
	"context"

	"localShareGo/internal/config"
	"localShareGo/internal/settings"
)

type noopManager struct {
	settings *settings.Service
}

func newPlatformManager(paths config.AppPaths, _ []byte) (platformManager, error) {
	return &noopManager{settings: settings.New(paths.DesktopSettingsPath)}, nil
}

func (m *noopManager) Start(context.Context) error {
	return nil
}

func (m *noopManager) Stop() {}

func (m *noopManager) Show() error {
	return nil
}

func (m *noopManager) Hide() error {
	return nil
}

func (m *noopManager) Settings() (settings.DesktopSettings, error) {
	return m.settings.Load()
}

func (m *noopManager) UpdateSettings(next settings.DesktopSettings) (settings.DesktopSettings, error) {
	return m.settings.Save(next)
}
