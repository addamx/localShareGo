//go:build !windows

package clipboard

import "fmt"

func readClipboardText() (string, error) {
	return "", fmt.Errorf("clipboard integration is only implemented for Windows in this build")
}

func writeClipboardText(string) error {
	return fmt.Errorf("clipboard integration is only implemented for Windows in this build")
}
