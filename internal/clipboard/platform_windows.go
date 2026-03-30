//go:build windows

package clipboard

import (
	"syscall"
	"time"
	"unicode/utf16"
	"unsafe"
)

const (
	cfUnicodeText = 13
	gmemMoveable  = 0x0002
)

var (
	user32                         = syscall.NewLazyDLL("user32.dll")
	kernel32                       = syscall.NewLazyDLL("kernel32.dll")
	procOpenClipboard              = user32.NewProc("OpenClipboard")
	procCloseClipboard             = user32.NewProc("CloseClipboard")
	procGetClipboardData           = user32.NewProc("GetClipboardData")
	procSetClipboardData           = user32.NewProc("SetClipboardData")
	procEmptyClipboard             = user32.NewProc("EmptyClipboard")
	procIsClipboardFormatAvailable = user32.NewProc("IsClipboardFormatAvailable")
	procGlobalAlloc                = kernel32.NewProc("GlobalAlloc")
	procGlobalFree                 = kernel32.NewProc("GlobalFree")
	procGlobalLock                 = kernel32.NewProc("GlobalLock")
	procGlobalUnlock               = kernel32.NewProc("GlobalUnlock")
)

func readClipboardText() (string, error) {
	if err := openClipboard(); err != nil {
		return "", err
	}
	defer procCloseClipboard.Call()

	available, _, _ := procIsClipboardFormatAvailable.Call(cfUnicodeText)
	if available == 0 {
		return "", nil
	}

	handle, _, err := procGetClipboardData.Call(cfUnicodeText)
	if handle == 0 {
		return "", err
	}

	ptr, _, err := procGlobalLock.Call(handle)
	if ptr == 0 {
		return "", err
	}
	defer procGlobalUnlock.Call(handle)

	return utf16PtrToString(ptr), nil
}

func writeClipboardText(text string) error {
	if err := openClipboard(); err != nil {
		return err
	}
	defer procCloseClipboard.Call()

	if _, _, err := procEmptyClipboard.Call(); err != syscall.Errno(0) {
		return err
	}

	data := utf16.Encode([]rune(text + "\x00"))
	size := uintptr(len(data) * 2)

	handle, _, err := procGlobalAlloc.Call(gmemMoveable, size)
	if handle == 0 {
		return err
	}

	ptr, _, err := procGlobalLock.Call(handle)
	if ptr == 0 {
		procGlobalFree.Call(handle)
		return err
	}

	buffer := unsafe.Slice((*uint16)(unsafe.Pointer(ptr)), len(data))
	copy(buffer, data)
	procGlobalUnlock.Call(handle)

	result, _, err := procSetClipboardData.Call(cfUnicodeText, handle)
	if result == 0 {
		procGlobalFree.Call(handle)
		return err
	}

	return nil
}

func openClipboard() error {
	var lastErr error
	for attempt := 0; attempt < 12; attempt++ {
		result, _, err := procOpenClipboard.Call(0)
		if result != 0 {
			return nil
		}
		lastErr = err
		time.Sleep(20 * time.Millisecond)
	}
	return lastErr
}

func utf16PtrToString(ptr uintptr) string {
	if ptr == 0 {
		return ""
	}
	var values []uint16
	for offset := uintptr(0); ; offset += 2 {
		value := *(*uint16)(unsafe.Pointer(ptr + offset))
		if value == 0 {
			break
		}
		values = append(values, value)
	}
	return string(utf16.Decode(values))
}
