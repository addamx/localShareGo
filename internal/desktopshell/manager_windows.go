//go:build windows

package desktopshell

import (
	"context"
	"fmt"
	"sync"
	"time"
	"unsafe"

	wruntime "github.com/wailsapp/wails/v2/pkg/runtime"
	"golang.org/x/sys/windows"

	"localShareGo/internal/config"
	"localShareGo/internal/settings"
)

const (
	gwlExStyle          = -20
	swHide              = 0
	swShow              = 5
	swRestore           = 9
	swpNoMove           = 0x0002
	swpNoSize           = 0x0001
	swpNoZOrder         = 0x0004
	swpFrameDraw        = 0x0020
	wsExAppWnd          = 0x00040000
	wsExToolWnd         = 0x00000080
	smXVirtualScreen    = 76
	smYVirtualScreen    = 77
	smCXVirtualScreen   = 78
	smCYVirtualScreen   = 79
	windowOffset        = 12
	defaultWindowWidth  = 512
	defaultWindowHeight = 384
)

var (
	user32                = windows.NewLazySystemDLL("user32.dll")
	procFindWindowW       = user32.NewProc("FindWindowW")
	procGetWindowLongPtrW = user32.NewProc("GetWindowLongPtrW")
	procGetWindowRect     = user32.NewProc("GetWindowRect")
	procGetSystemMetrics  = user32.NewProc("GetSystemMetrics")
	procSetWindowLongPtrW = user32.NewProc("SetWindowLongPtrW")
	procSetWindowPos      = user32.NewProc("SetWindowPos")
	procShowWindow        = user32.NewProc("ShowWindow")
	procSetForeground     = user32.NewProc("SetForegroundWindow")
)

type windowRect struct {
	Left   int32
	Top    int32
	Right  int32
	Bottom int32
}

type windowsManager struct {
	ctx      context.Context
	settings *settings.Service
	trayIcon []byte

	windowMu sync.Mutex
	window   windows.Handle
	loop     *trayLoop

	stateMu sync.RWMutex
	pinned  bool
}

func newPlatformManager(paths config.AppPaths, trayIcon []byte) (platformManager, error) {
	return &windowsManager{
		settings: settings.New(paths.DesktopSettingsPath),
		trayIcon: trayIcon,
	}, nil
}

func (m *windowsManager) Start(ctx context.Context) error {
	m.ctx = ctx

	currentSettings, err := m.settings.Load()
	if err != nil {
		return err
	}

	binding, err := normalizeHotkeyValue(currentSettings.ShowAppHotkey)
	if err != nil {
		binding = hotkeyBinding{}
	}

	loop, err := newTrayLoop(m.trayIcon, func() {
		_ = m.Show()
	}, func() {
		_ = m.showNearCursor()
	}, func() {
		if m.ctx != nil {
			wruntime.Quit(m.ctx)
		}
	})
	if err != nil {
		return err
	}
	if err := loop.Start(binding); err != nil {
		return err
	}
	m.loop = loop

	return m.applyTaskbarHiddenStyle()
}

func (m *windowsManager) Stop() {
	if m.loop != nil {
		m.loop.Stop()
	}
}

func (m *windowsManager) Show() error {
	return m.show(false)
}

func (m *windowsManager) showNearCursor() error {
	return m.show(true)
}

func (m *windowsManager) show(nearCursor bool) error {
	hwnd, err := m.ensureWindow()
	if err != nil {
		return err
	}
	if err := m.applyToolWindow(hwnd); err != nil {
		return err
	}
	if nearCursor {
		if err := m.positionWindowNearCursor(hwnd); err != nil {
			return err
		}
	}
	showWindow(hwnd, swRestore)
	showWindow(hwnd, swShow)
	setForegroundWindow(hwnd)
	return nil
}

func (m *windowsManager) Hide() error {
	if m.IsPinned() {
		return nil
	}

	hwnd, err := m.ensureWindow()
	if err != nil {
		return err
	}
	showWindow(hwnd, swHide)
	return nil
}

func (m *windowsManager) IsPinned() bool {
	m.stateMu.RLock()
	defer m.stateMu.RUnlock()
	return m.pinned
}

func (m *windowsManager) SetPinned(pinned bool) bool {
	m.stateMu.Lock()
	defer m.stateMu.Unlock()
	m.pinned = pinned
	return m.pinned
}

func (m *windowsManager) Settings() (settings.DesktopSettings, error) {
	return m.settings.Load()
}

func (m *windowsManager) UpdateSettings(next settings.DesktopSettings) (settings.DesktopSettings, error) {
	binding, err := normalizeHotkeyValue(next.ShowAppHotkey)
	if err != nil {
		return settings.DesktopSettings{}, err
	}
	if m.loop != nil {
		if err := m.loop.UpdateHotkey(binding); err != nil {
			return settings.DesktopSettings{}, err
		}
	}
	next.ShowAppHotkey = binding.Text
	return m.settings.Save(next)
}

func (m *windowsManager) applyTaskbarHiddenStyle() error {
	hwnd, err := m.ensureWindow()
	if err != nil {
		return err
	}
	return m.applyToolWindow(hwnd)
}

func (m *windowsManager) ensureWindow() (windows.Handle, error) {
	m.windowMu.Lock()
	defer m.windowMu.Unlock()

	if m.window != 0 {
		return m.window, nil
	}

	className, err := windows.UTF16PtrFromString(WindowClassName)
	if err != nil {
		return 0, err
	}

	var hwnd windows.Handle
	for attempt := 0; attempt < 50; attempt++ {
		result, _, _ := procFindWindowW.Call(uintptr(unsafe.Pointer(className)), 0)
		hwnd = windows.Handle(result)
		if hwnd != 0 {
			m.window = hwnd
			return hwnd, nil
		}
		time.Sleep(40 * time.Millisecond)
	}

	return 0, fmt.Errorf("desktop window not found")
}

func (m *windowsManager) applyToolWindow(hwnd windows.Handle) error {
	style, _, callErr := procGetWindowLongPtrW.Call(uintptr(hwnd), windowLongIndex(gwlExStyle))
	if style == 0 && callErr != windows.ERROR_SUCCESS {
		return callErr
	}

	nextStyle := (style &^ wsExAppWnd) | wsExToolWnd
	if nextStyle == style {
		return nil
	}

	result, _, setErr := procSetWindowLongPtrW.Call(uintptr(hwnd), windowLongIndex(gwlExStyle), nextStyle)
	if result == 0 && setErr != windows.ERROR_SUCCESS {
		return setErr
	}

	posResult, _, posErr := procSetWindowPos.Call(
		uintptr(hwnd),
		0,
		0,
		0,
		0,
		0,
		uintptr(swpNoMove|swpNoSize|swpNoZOrder|swpFrameDraw),
	)
	if posResult == 0 && posErr != windows.ERROR_SUCCESS {
		return posErr
	}
	return nil
}

func (m *windowsManager) positionWindowNearCursor(hwnd windows.Handle) error {
	var point trayPoint
	if result, _, err := procGetCursorPos.Call(uintptr(unsafe.Pointer(&point))); result == 0 {
		return err
	}

	width, height, err := getWindowSize(hwnd)
	if err != nil {
		return err
	}

	virtualLeft := getSystemMetric(smXVirtualScreen)
	virtualTop := getSystemMetric(smYVirtualScreen)
	virtualWidth := getSystemMetric(smCXVirtualScreen)
	virtualHeight := getSystemMetric(smCYVirtualScreen)

	x := point.X + windowOffset
	y := point.Y + windowOffset

	maxX := virtualLeft + virtualWidth - width
	maxY := virtualTop + virtualHeight - height
	if x > maxX {
		x = maxX
	}
	if y > maxY {
		y = maxY
	}
	if x < virtualLeft {
		x = virtualLeft
	}
	if y < virtualTop {
		y = virtualTop
	}

	if result, _, err := procSetWindowPos.Call(
		uintptr(hwnd),
		0,
		uintptr(int32ToUint32(x)),
		uintptr(int32ToUint32(y)),
		0,
		0,
		uintptr(swpNoSize|swpNoZOrder),
	); result == 0 {
		return err
	}

	return nil
}

func getWindowSize(hwnd windows.Handle) (int32, int32, error) {
	rect := windowRect{}
	if result, _, err := procGetWindowRect.Call(uintptr(hwnd), uintptr(unsafe.Pointer(&rect))); result == 0 {
		return 0, 0, err
	}

	width := rect.Right - rect.Left
	height := rect.Bottom - rect.Top
	if width <= 0 {
		width = defaultWindowWidth
	}
	if height <= 0 {
		height = defaultWindowHeight
	}
	return width, height, nil
}

func getSystemMetric(metric int32) int32 {
	value, _, _ := procGetSystemMetrics.Call(uintptr(metric))
	return int32(value)
}

func int32ToUint32(value int32) uint32 {
	return uint32(value)
}

func showWindow(hwnd windows.Handle, command int32) {
	procShowWindow.Call(uintptr(hwnd), uintptr(command))
}

func setForegroundWindow(hwnd windows.Handle) {
	procSetForeground.Call(uintptr(hwnd))
}

func windowLongIndex(value int32) uintptr {
	return uintptr(uint32(value))
}
