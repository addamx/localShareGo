//go:build windows

package desktopshell

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"sync"
	"unsafe"

	"golang.org/x/sys/windows"
)

const (
	trayMenuShowID = 1001
	trayMenuQuitID = 1002
	trayMsgID      = 0x0401
	wmAppCommand   = 0x8001
	wmHotkey       = 0x0312
	modByCommand   = 0
	mfString       = 0
	tpmLeftAlign   = 0
	tpmBottomAlign = 0x20
	imageIcon      = 1
	lrLoadFromFile = 0x0010
	lrDefaultSize  = 0x0040
	nifMessage     = 0x00000001
	nifIcon        = 0x00000002
	nifTip         = 0x00000004
	nimAdd         = 0x00000000
	nimModify      = 0x00000001
	nimDelete      = 0x00000002
	idiApplication = 32512
	idcArrow       = 32512
	wmDestroy      = 0x0002
	wmClose        = 0x0010
	wmCommand      = 0x0111
	wmLButtonUp    = 0x0202
	wmRButtonUp    = 0x0205
	wsOverlapped   = 0x00000000
)

var (
	kernel32             = windows.NewLazySystemDLL("kernel32.dll")
	procGetModuleHandleW = kernel32.NewProc("GetModuleHandleW")
	procGetCurrentThread = kernel32.NewProc("GetCurrentThreadId")

	shell32             = windows.NewLazySystemDLL("shell32.dll")
	procShellNotifyIcon = shell32.NewProc("Shell_NotifyIconW")

	procAppendMenuW         = user32.NewProc("AppendMenuW")
	procCreatePopupMenu     = user32.NewProc("CreatePopupMenu")
	procCreateWindowExW     = user32.NewProc("CreateWindowExW")
	procDefWindowProcW      = user32.NewProc("DefWindowProcW")
	procDestroyMenu         = user32.NewProc("DestroyMenu")
	procDestroyWindow       = user32.NewProc("DestroyWindow")
	procDispatchMessageW    = user32.NewProc("DispatchMessageW")
	procGetCursorPos        = user32.NewProc("GetCursorPos")
	procGetMessageW         = user32.NewProc("GetMessageW")
	procLoadCursorW         = user32.NewProc("LoadCursorW")
	procLoadIconW           = user32.NewProc("LoadIconW")
	procLoadImageW          = user32.NewProc("LoadImageW")
	procPostMessageW        = user32.NewProc("PostMessageW")
	procPostQuitMessage     = user32.NewProc("PostQuitMessage")
	procRegisterClassExW    = user32.NewProc("RegisterClassExW")
	procRegisterHotKey      = user32.NewProc("RegisterHotKey")
	procSetForegroundWindow = user32.NewProc("SetForegroundWindow")
	procTrackPopupMenu      = user32.NewProc("TrackPopupMenu")
	procTranslateMessage    = user32.NewProc("TranslateMessage")
	procUnregisterClassW    = user32.NewProc("UnregisterClassW")
	procUnregisterHotKey    = user32.NewProc("UnregisterHotKey")
)

type trayLoop struct {
	iconBytes []byte
	onShow    func()
	onHotkey  func()
	onQuit    func()

	className *uint16
	window    windows.Handle
	menu      windows.Handle
	threadID  uint32

	commandMu   sync.Mutex
	commandFn   func() error
	commandDone chan error

	hotkey    hotkeyBinding
	startedCh chan error
	stopOnce  sync.Once
}

type trayPoint struct {
	X int32
	Y int32
}

type trayMessage struct {
	Hwnd    windows.Handle
	Message uint32
	WParam  uintptr
	LParam  uintptr
	Time    uint32
	Point   trayPoint
}

type trayWndClassEx struct {
	Size, Style                        uint32
	WndProc                            uintptr
	ClsExtra, WndExtra                 int32
	Instance, Icon, Cursor, Background windows.Handle
	MenuName, ClassName                *uint16
	IconSm                             windows.Handle
}

type trayNotifyIconData struct {
	Size                       uint32
	Wnd                        windows.Handle
	ID, Flags, CallbackMessage uint32
	Icon                       windows.Handle
	Tip                        [128]uint16
	State, StateMask           uint32
	Info                       [256]uint16
	Timeout, Version           uint32
	InfoTitle                  [64]uint16
	InfoFlags                  uint32
	GuidItem                   windows.GUID
	BalloonIcon                windows.Handle
}

func newTrayLoop(iconBytes []byte, onShow func(), onHotkey func(), onQuit func()) (*trayLoop, error) {
	className, err := windows.UTF16PtrFromString("LocalShareGoTrayWindow")
	if err != nil {
		return nil, err
	}
	return &trayLoop{
		iconBytes: iconBytes,
		onShow:    onShow,
		onHotkey:  onHotkey,
		onQuit:    onQuit,
		className: className,
		startedCh: make(chan error, 1),
	}, nil
}

func (t *trayLoop) Start(initialHotkey hotkeyBinding) error {
	go t.run(initialHotkey)
	return <-t.startedCh
}

func (t *trayLoop) Stop() {
	t.stopOnce.Do(func() {
		if t.window != 0 {
			procPostMessageW.Call(uintptr(t.window), uintptr(wmClose), 0, 0)
		}
	})
}

func (t *trayLoop) UpdateHotkey(binding hotkeyBinding) error {
	t.commandMu.Lock()
	t.commandFn = func() error {
		return t.applyHotkey(binding)
	}
	t.commandDone = make(chan error, 1)
	done := t.commandDone
	t.commandMu.Unlock()

	if t.window == 0 {
		return fmt.Errorf("tray loop is not ready")
	}
	procPostMessageW.Call(uintptr(t.window), uintptr(wmAppCommand), 0, 0)
	return <-done
}

func (t *trayLoop) run(initialHotkey hotkeyBinding) {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	threadID, _, _ := procGetCurrentThread.Call()
	t.threadID = uint32(threadID)

	if err := t.initWindow(); err != nil {
		t.startedCh <- err
		return
	}
	if err := t.initMenu(); err != nil {
		t.startedCh <- err
		return
	}
	if err := t.initNotifyIcon(); err != nil {
		t.startedCh <- err
		return
	}
	if err := t.applyHotkey(initialHotkey); err != nil {
		t.startedCh <- err
		return
	}

	t.startedCh <- nil

	var msg trayMessage
	for {
		result, _, err := procGetMessageW.Call(uintptr(unsafe.Pointer(&msg)), 0, 0, 0)
		switch int32(result) {
		case -1:
			return
		case 0:
			return
		default:
			_ = err
			if msg.Message == wmHotkey {
				go t.onHotkey()
				continue
			}
			procTranslateMessage.Call(uintptr(unsafe.Pointer(&msg)))
			procDispatchMessageW.Call(uintptr(unsafe.Pointer(&msg)))
		}
	}
}

func (t *trayLoop) initWindow() error {
	instance, _, err := procGetModuleHandleW.Call(0)
	if instance == 0 {
		return err
	}

	iconHandle, _, err := procLoadIconW.Call(0, uintptr(idiApplication))
	if iconHandle == 0 {
		return err
	}
	cursorHandle, _, err := procLoadCursorW.Call(0, uintptr(idcArrow))
	if cursorHandle == 0 {
		return err
	}

	windowClass := trayWndClassEx{
		Size:       uint32(unsafe.Sizeof(trayWndClassEx{})),
		WndProc:    windows.NewCallback(t.windowProc),
		Instance:   windows.Handle(instance),
		Icon:       windows.Handle(iconHandle),
		Cursor:     windows.Handle(cursorHandle),
		Background: windows.Handle(6),
		ClassName:  t.className,
		IconSm:     windows.Handle(iconHandle),
	}
	result, _, classErr := procRegisterClassExW.Call(uintptr(unsafe.Pointer(&windowClass)))
	if result == 0 {
		return classErr
	}

	windowHandle, _, createErr := procCreateWindowExW.Call(
		0,
		uintptr(unsafe.Pointer(t.className)),
		0,
		uintptr(wsOverlapped),
		0,
		0,
		0,
		0,
		0,
		0,
		instance,
		0,
	)
	if windowHandle == 0 {
		return createErr
	}

	t.window = windows.Handle(windowHandle)
	return nil
}

func (t *trayLoop) initMenu() error {
	menuHandle, _, err := procCreatePopupMenu.Call()
	if menuHandle == 0 {
		return err
	}
	t.menu = windows.Handle(menuHandle)

	showLabel, _ := windows.UTF16PtrFromString("Show App")
	quitLabel, _ := windows.UTF16PtrFromString("Exit")

	if result, _, appendErr := procAppendMenuW.Call(uintptr(t.menu), uintptr(mfString), uintptr(trayMenuShowID), uintptr(unsafe.Pointer(showLabel))); result == 0 {
		return appendErr
	}
	if result, _, appendErr := procAppendMenuW.Call(uintptr(t.menu), uintptr(mfString), uintptr(trayMenuQuitID), uintptr(unsafe.Pointer(quitLabel))); result == 0 {
		return appendErr
	}
	return nil
}

func (t *trayLoop) initNotifyIcon() error {
	nid := trayNotifyIconData{
		Size:            uint32(unsafe.Sizeof(trayNotifyIconData{})),
		Wnd:             t.window,
		ID:              1,
		Flags:           nifMessage,
		CallbackMessage: trayMsgID,
	}

	if len(t.iconBytes) > 0 {
		iconPath, err := writeTrayIcon(t.iconBytes)
		if err != nil {
			return err
		}
		iconHandle, _, loadErr := procLoadImageW.Call(
			0,
			uintptr(unsafe.Pointer(iconPath)),
			uintptr(imageIcon),
			0,
			0,
			uintptr(lrLoadFromFile|lrDefaultSize),
		)
		if iconHandle == 0 {
			return loadErr
		}
		nid.Icon = windows.Handle(iconHandle)
		nid.Flags |= nifIcon
	}

	tip, _ := windows.UTF16FromString("LocalShareGo")
	copy(nid.Tip[:], tip)
	nid.Flags |= nifTip

	result, _, err := procShellNotifyIcon.Call(uintptr(nimAdd), uintptr(unsafe.Pointer(&nid)))
	if result == 0 {
		return err
	}
	return nil
}

func (t *trayLoop) applyHotkey(binding hotkeyBinding) error {
	if t.hotkey.Key != 0 {
		procUnregisterHotKey.Call(0, 1)
		t.hotkey = hotkeyBinding{}
	}

	if binding.Key == 0 {
		return nil
	}

	result, _, err := procRegisterHotKey.Call(0, 1, uintptr(binding.Modifiers), uintptr(binding.Key))
	if result == 0 {
		return fmt.Errorf("register hotkey failed: %w", err)
	}

	t.hotkey = binding
	return nil
}

func (t *trayLoop) windowProc(hwnd uintptr, message uint32, wParam, lParam uintptr) uintptr {
	switch message {
	case trayMsgID:
		switch uint32(lParam) {
		case wmLButtonUp:
			go t.onShow()
		case wmRButtonUp:
			t.showMenu()
		}
		return 0
	case wmHotkey:
		go t.onHotkey()
		return 0
	case wmCommand:
		switch uint32(wParam) {
		case trayMenuShowID:
			go t.onShow()
		case trayMenuQuitID:
			go t.onQuit()
		}
		return 0
	case wmAppCommand:
		t.handleCommand()
		return 0
	case wmClose:
		procDestroyWindow.Call(hwnd)
		return 0
	case wmDestroy:
		t.cleanup()
		procPostQuitMessage.Call(0)
		return 0
	default:
		result, _, _ := procDefWindowProcW.Call(hwnd, uintptr(message), wParam, lParam)
		return result
	}
}

func (t *trayLoop) handleCommand() {
	t.commandMu.Lock()
	commandFn := t.commandFn
	commandDone := t.commandDone
	t.commandFn = nil
	t.commandDone = nil
	t.commandMu.Unlock()

	if commandFn == nil || commandDone == nil {
		return
	}
	commandDone <- commandFn()
}

func (t *trayLoop) showMenu() {
	var point trayPoint
	if result, _, _ := procGetCursorPos.Call(uintptr(unsafe.Pointer(&point))); result == 0 {
		return
	}
	procSetForegroundWindow.Call(uintptr(t.window))
	procTrackPopupMenu.Call(
		uintptr(t.menu),
		uintptr(tpmLeftAlign|tpmBottomAlign),
		uintptr(point.X),
		uintptr(point.Y),
		0,
		uintptr(t.window),
		0,
	)
}

func (t *trayLoop) cleanup() {
	if t.hotkey.Key != 0 {
		procUnregisterHotKey.Call(0, 1)
	}

	nid := trayNotifyIconData{
		Size: uint32(unsafe.Sizeof(trayNotifyIconData{})),
		Wnd:  t.window,
		ID:   1,
	}
	procShellNotifyIcon.Call(uintptr(nimDelete), uintptr(unsafe.Pointer(&nid)))

	if t.menu != 0 {
		procDestroyMenu.Call(uintptr(t.menu))
	}

	module, _, _ := procGetModuleHandleW.Call(0)
	procUnregisterClassW.Call(uintptr(unsafe.Pointer(t.className)), module)
}

func writeTrayIcon(iconBytes []byte) (*uint16, error) {
	sum := md5.Sum(iconBytes)
	path := filepath.Join(os.TempDir(), "localsharego-tray-"+hex.EncodeToString(sum[:])+".ico")
	if _, err := os.Stat(path); os.IsNotExist(err) {
		if err := os.WriteFile(path, iconBytes, 0o644); err != nil {
			return nil, err
		}
	}
	return windows.UTF16PtrFromString(path)
}
