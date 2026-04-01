package settings

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

type DesktopSettings struct {
	ShowAppHotkey string `json:"showAppHotkey"`
}

type Service struct {
	path string
	mu   sync.Mutex
}

func New(path string) *Service {
	return &Service{path: path}
}

func (s *Service) Load() (DesktopSettings, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	content, err := os.ReadFile(s.path)
	if err != nil {
		if os.IsNotExist(err) {
			return DesktopSettings{}, nil
		}
		return DesktopSettings{}, err
	}
	if len(content) == 0 {
		return DesktopSettings{}, nil
	}

	var state DesktopSettings
	if err := json.Unmarshal(content, &state); err != nil {
		return DesktopSettings{}, err
	}
	state.ShowAppHotkey = strings.TrimSpace(state.ShowAppHotkey)
	return state, nil
}

func (s *Service) Save(next DesktopSettings) (DesktopSettings, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	next.ShowAppHotkey = strings.TrimSpace(next.ShowAppHotkey)
	if err := os.MkdirAll(filepath.Dir(s.path), 0o755); err != nil {
		return DesktopSettings{}, err
	}

	content, err := json.MarshalIndent(next, "", "  ")
	if err != nil {
		return DesktopSettings{}, err
	}

	tempPath := s.path + ".tmp"
	if err := os.WriteFile(tempPath, content, 0o644); err != nil {
		return DesktopSettings{}, err
	}
	if err := os.Rename(tempPath, s.path); err != nil {
		return DesktopSettings{}, err
	}
	return next, nil
}
