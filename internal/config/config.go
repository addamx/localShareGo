package config

import (
	"os"
	"path/filepath"
)

const (
	DefaultLanHost               = "0.0.0.0"
	DefaultPreferredPort         = 8765
	DefaultMaxTextBytes          = 65536
	DefaultClipboardPollInterval = 800
	DefaultTokenTTLMinutes       = 60
	DefaultDataFileName          = "localshare.json"
	DefaultWebRoute              = "/web"
)

type RuntimeConfig struct {
	LanHost               string `json:"lanHost"`
	PreferredPort         int    `json:"preferredPort"`
	MaxTextBytes          int    `json:"maxTextBytes"`
	ClipboardPollInterval int    `json:"clipboardPollIntervalMs"`
	TokenTTLMinutes       int    `json:"tokenTtlMinutes"`
	DatabaseFileName      string `json:"databaseFileName"`
	WebRoute              string `json:"webRoute"`
}

type AppPaths struct {
	AppDir       string `json:"appDir"`
	DataDir      string `json:"dataDir"`
	DatabasePath string `json:"databasePath"`
	LogsDir      string `json:"logsDir"`
}

func DefaultRuntimeConfig() RuntimeConfig {
	return RuntimeConfig{
		LanHost:               DefaultLanHost,
		PreferredPort:         DefaultPreferredPort,
		MaxTextBytes:          DefaultMaxTextBytes,
		ClipboardPollInterval: DefaultClipboardPollInterval,
		TokenTTLMinutes:       DefaultTokenTTLMinutes,
		DatabaseFileName:      DefaultDataFileName,
		WebRoute:              DefaultWebRoute,
	}
}

func ResolveAppPaths(config RuntimeConfig) AppPaths {
	root, err := os.UserConfigDir()
	if err != nil || root == "" {
		root = "."
	}

	appDir := filepath.Join(root, "LocalShareGo")
	dataDir := filepath.Join(appDir, "data")
	logsDir := filepath.Join(appDir, "logs")

	return AppPaths{
		AppDir:       appDir,
		DataDir:      dataDir,
		DatabasePath: filepath.Join(dataDir, config.DatabaseFileName),
		LogsDir:      logsDir,
	}
}

func EnsureAppDirs(paths AppPaths) error {
	for _, dir := range []string{paths.AppDir, paths.DataDir, paths.LogsDir} {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return err
		}
	}
	return nil
}
