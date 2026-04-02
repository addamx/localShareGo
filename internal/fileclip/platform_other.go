//go:build !windows

package fileclip

import "fmt"

func ReadClipboardFile() (ClipboardFile, bool, error) {
	return ClipboardFile{}, false, nil
}

func WriteClipboardFile(string) error {
	return fmt.Errorf("file clipboard integration is only implemented for Windows in this build")
}
