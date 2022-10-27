package config

import (
	"os"
	"path/filepath"

	"gopkg.in/ini.v1"
)

// TODO: default values?

type AltCoresConfig struct {
	LLAPI   []string `ini:"llapi,omitempty" delim:","`
	YC      []string `ini:"yc,omitempty" delim:","`
	DualRAM []string `ini:"dualram,omitempty" delim:","`
}

type LaunchSyncConfig struct{}

type PlayLogConfig struct {
	SaveEvery   int    `ini:"save_every,omitempty"`
	OnCoreStart string `ini:"on_core_start,omitempty"`
	OnCoreStop  string `ini:"on_core_stop,omitempty"`
	OnGameStart string `ini:"on_game_start,omitempty"`
	OnGameStop  string `ini:"on_game_stop,omitempty"`
}

type RandomConfig struct{}

type SearchConfig struct {
	Filter []string `ini:"filter,omitempty" delim:","`
	Sort   string   `ini:"sort,omitempty"`
}

type UserConfig struct {
	AltCores   AltCoresConfig
	LaunchSync LaunchSyncConfig
	PlayLog    PlayLogConfig
	Random     RandomConfig
	Search     SearchConfig
}

func LoadUserConfig(defaultConfig UserConfig) (UserConfig, error) {
	userConfig := defaultConfig

	// TODO: central default ini first

	// TODO: check if this can be a relative path
	appFolder := filepath.Dir(os.Args[0])
	appFile := filepath.Base(os.Args[0])
	appName := appFile[:len(appFile)-len(filepath.Ext(appFile))]
	iniPath := filepath.Join(appFolder, appName+".ini")

	if _, err := os.Stat(iniPath); os.IsNotExist(err) {
		return userConfig, nil
	}

	err := ini.MapTo(&userConfig, iniPath)
	if err != nil {
		return userConfig, err
	}

	return userConfig, nil
}
