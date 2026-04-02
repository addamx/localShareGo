//go:build !windows

package desktopshell

import (
	"context"

	"localShareGo/internal/config"
	"localShareGo/internal/settings"
)

type noopManager struct {
	settings *settings.Service
	pinned   bool
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
	if m.pinned {
		return nil
	}
	return nil
}

func (m *noopManager) IsPinned() bool {
	return m.pinned
}

func (m *noopManager) SetPinned(pinned bool) bool {
	m.pinned = pinned
	return m.pinned
}

func (m *noopManager) Settings() (settings.DesktopSettings, error) {
	return m.settings.Load()
}

func (m *noopManager) UpdateSettings(next settings.DesktopSettings) (settings.DesktopSettings, error) {
	return m.settings.Save(next)
}
