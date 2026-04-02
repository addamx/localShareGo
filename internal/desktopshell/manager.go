package desktopshell

import (
	"context"

	"localShareGo/internal/config"
	"localShareGo/internal/settings"
)

const WindowClassName = "LocalShareGoWindow"

type platformManager interface {
	Start(ctx context.Context) error
	Stop()
	Show() error
	Hide() error
	IsPinned() bool
	SetPinned(pinned bool) bool
	Settings() (settings.DesktopSettings, error)
	UpdateSettings(next settings.DesktopSettings) (settings.DesktopSettings, error)
}

type Manager struct {
	impl platformManager
}

func New(paths config.AppPaths, trayIcon []byte) (*Manager, error) {
	impl, err := newPlatformManager(paths, trayIcon)
	if err != nil {
		return nil, err
	}
	return &Manager{impl: impl}, nil
}

func (m *Manager) Start(ctx context.Context) error {
	return m.impl.Start(ctx)
}

func (m *Manager) Stop() {
	m.impl.Stop()
}

func (m *Manager) Show() error {
	return m.impl.Show()
}

func (m *Manager) Hide() error {
	return m.impl.Hide()
}

func (m *Manager) IsPinned() bool {
	return m.impl.IsPinned()
}

func (m *Manager) SetPinned(pinned bool) bool {
	return m.impl.SetPinned(pinned)
}

func (m *Manager) Settings() (settings.DesktopSettings, error) {
	return m.impl.Settings()
}

func (m *Manager) UpdateSettings(next settings.DesktopSettings) (settings.DesktopSettings, error) {
	return m.impl.UpdateSettings(next)
}
