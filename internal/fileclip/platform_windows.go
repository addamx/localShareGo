//go:build windows

package fileclip

import (
	"fmt"
	"path/filepath"
	"syscall"
	"time"
	"unicode/utf16"
	"unsafe"
)

const (
	cfHdrop      = 15
	gmemMoveable = 0x0002
)

type dropFiles struct {
	PFiles uint32
	X      int32
	Y      int32
	FNC    int32
	FWide  int32
}

var (
	user32                         = syscall.NewLazyDLL("user32.dll")
	shell32                        = syscall.NewLazyDLL("shell32.dll")
	kernel32                       = syscall.NewLazyDLL("kernel32.dll")
	procOpenClipboard              = user32.NewProc("OpenClipboard")
	procCloseClipboard             = user32.NewProc("CloseClipboard")
	procGetClipboardData           = user32.NewProc("GetClipboardData")
	procSetClipboardData           = user32.NewProc("SetClipboardData")
	procEmptyClipboard             = user32.NewProc("EmptyClipboard")
	procIsClipboardFormatAvailable = user32.NewProc("IsClipboardFormatAvailable")
	procDragQueryFileW             = shell32.NewProc("DragQueryFileW")
	procGlobalAlloc                = kernel32.NewProc("GlobalAlloc")
	procGlobalFree                 = kernel32.NewProc("GlobalFree")
	procGlobalLock                 = kernel32.NewProc("GlobalLock")
	procGlobalUnlock               = kernel32.NewProc("GlobalUnlock")
)

func ReadClipboardFile() (ClipboardFile, bool, error) {
	if err := openClipboard(); err != nil {
		return ClipboardFile{}, false, err
	}
	defer procCloseClipboard.Call()

	available, _, _ := procIsClipboardFormatAvailable.Call(cfHdrop)
	if available == 0 {
		return ClipboardFile{}, false, nil
	}

	handle, _, err := procGetClipboardData.Call(cfHdrop)
	if handle == 0 {
		return ClipboardFile{}, false, err
	}

	count, _, err := procDragQueryFileW.Call(handle, 0xFFFFFFFF, 0, 0)
	if count == 0 {
		return ClipboardFile{}, false, err
	}

	length, _, err := procDragQueryFileW.Call(handle, 0, 0, 0)
	if length == 0 {
		return ClipboardFile{}, false, err
	}

	buffer := make([]uint16, length+1)
	procDragQueryFileW.Call(handle, 0, uintptr(unsafe.Pointer(&buffer[0])), uintptr(len(buffer)))
	path := syscall.UTF16ToString(buffer)
	if path == "" {
		return ClipboardFile{}, false, nil
	}

	meta, err := InspectPath(path)
	if err != nil {
		return ClipboardFile{}, false, err
	}

	return ClipboardFile{
		Path:     path,
		Metadata: meta,
	}, true, nil
}

func WriteClipboardFile(path string) error {
	if _, err := InspectPath(path); err != nil {
		return err
	}

	if err := openClipboard(); err != nil {
		return err
	}
	defer procCloseClipboard.Call()

	if result, _, err := procEmptyClipboard.Call(); result == 0 {
		return err
	}

	encoded := utf16.Encode([]rune(filepath.Clean(path) + "\x00\x00"))
	size := uintptr(unsafe.Sizeof(dropFiles{})) + uintptr(len(encoded))*2

	handle, _, err := procGlobalAlloc.Call(gmemMoveable, size)
	if handle == 0 {
		return err
	}

	ptr, _, err := procGlobalLock.Call(handle)
	if ptr == 0 {
		procGlobalFree.Call(handle)
		return err
	}

	drop := (*dropFiles)(unsafe.Pointer(ptr))
	drop.PFiles = uint32(unsafe.Sizeof(dropFiles{}))
	drop.FWide = 1

	dataPtr := uintptr(ptr) + unsafe.Sizeof(dropFiles{})
	buffer := unsafe.Slice((*uint16)(unsafe.Pointer(dataPtr)), len(encoded))
	copy(buffer, encoded)
	procGlobalUnlock.Call(handle)

	result, _, err := procSetClipboardData.Call(cfHdrop, handle)
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
	if lastErr != nil {
		return fmt.Errorf("%v", lastErr)
	}
	return fmt.Errorf("open clipboard failed")
}
