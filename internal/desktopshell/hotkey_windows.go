//go:build windows

package desktopshell

import (
	"fmt"
	"strconv"
	"strings"
)

const (
	modAlt     = 0x0001
	modControl = 0x0002
	modShift   = 0x0004
	modWin     = 0x0008

	vkBack   = 0x08
	vkTab    = 0x09
	vkReturn = 0x0D
	vkEscape = 0x1B
	vkSpace  = 0x20
	vkPrior  = 0x21
	vkNext   = 0x22
	vkEnd    = 0x23
	vkHome   = 0x24
	vkLeft   = 0x25
	vkUp     = 0x26
	vkRight  = 0x27
	vkDown   = 0x28
	vkInsert = 0x2D
	vkDelete = 0x2E
	vkF1     = 0x70
	vkF24    = 0x87
)

type hotkeyBinding struct {
	Text      string
	Modifiers uint32
	Key       uint32
}

func normalizeHotkeyValue(value string) (hotkeyBinding, error) {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return hotkeyBinding{}, nil
	}

	parts := strings.Split(trimmed, "+")
	modifiers := uint32(0)
	modifierLabels := make([]string, 0, 4)
	keyLabel := ""
	keyCode := uint32(0)

	for _, part := range parts {
		token := strings.TrimSpace(part)
		if token == "" {
			continue
		}

		upper := strings.ToUpper(token)
		switch upper {
		case "CTRL", "CONTROL":
			if modifiers&modControl == 0 {
				modifiers |= modControl
				modifierLabels = append(modifierLabels, "Ctrl")
			}
			continue
		case "ALT":
			if modifiers&modAlt == 0 {
				modifiers |= modAlt
				modifierLabels = append(modifierLabels, "Alt")
			}
			continue
		case "SHIFT":
			if modifiers&modShift == 0 {
				modifiers |= modShift
				modifierLabels = append(modifierLabels, "Shift")
			}
			continue
		case "WIN", "WINDOWS", "META", "CMD":
			if modifiers&modWin == 0 {
				modifiers |= modWin
				modifierLabels = append(modifierLabels, "Win")
			}
			continue
		}

		if keyCode != 0 {
			return hotkeyBinding{}, fmt.Errorf("hotkey must contain exactly one main key")
		}

		label, code, ok := resolveHotkeyKey(upper)
		if !ok {
			return hotkeyBinding{}, fmt.Errorf("unsupported hotkey key: %s", token)
		}
		keyLabel = label
		keyCode = code
	}

	if keyCode == 0 {
		return hotkeyBinding{}, fmt.Errorf("hotkey is missing a main key")
	}

	segments := append(modifierLabels, keyLabel)
	return hotkeyBinding{
		Text:      strings.Join(segments, "+"),
		Modifiers: modifiers,
		Key:       keyCode,
	}, nil
}

func resolveHotkeyKey(token string) (string, uint32, bool) {
	if len(token) == 1 {
		ch := token[0]
		if ch >= 'A' && ch <= 'Z' {
			return string(ch), uint32(ch), true
		}
		if ch >= '0' && ch <= '9' {
			return string(ch), uint32(ch), true
		}
	}

	if strings.HasPrefix(token, "F") {
		index, err := strconv.Atoi(strings.TrimPrefix(token, "F"))
		if err == nil && index >= 1 && index <= 24 {
			return fmt.Sprintf("F%d", index), vkF1 + uint32(index-1), true
		}
	}

	switch token {
	case "SPACE":
		return "Space", vkSpace, true
	case "TAB":
		return "Tab", vkTab, true
	case "ENTER":
		return "Enter", vkReturn, true
	case "ESC", "ESCAPE":
		return "Escape", vkEscape, true
	case "BACKSPACE":
		return "Backspace", vkBack, true
	case "LEFT", "ARROWLEFT":
		return "Left", vkLeft, true
	case "UP", "ARROWUP":
		return "Up", vkUp, true
	case "RIGHT", "ARROWRIGHT":
		return "Right", vkRight, true
	case "DOWN", "ARROWDOWN":
		return "Down", vkDown, true
	case "INSERT":
		return "Insert", vkInsert, true
	case "DELETE":
		return "Delete", vkDelete, true
	case "HOME":
		return "Home", vkHome, true
	case "END":
		return "End", vkEnd, true
	case "PAGEUP":
		return "PageUp", vkPrior, true
	case "PAGEDOWN":
		return "PageDown", vkNext, true
	}

	return "", 0, false
}
