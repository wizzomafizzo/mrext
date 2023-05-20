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

type LastPlayedConfig struct {
	Name string `ini:"name,omitempty"`
}

type UserConfig struct {
	AltCores   AltCoresConfig   `ini:"altcores,omitempty"`
	LaunchSync LaunchSyncConfig `ini:"launchsync,omitempty"`
	PlayLog    PlayLogConfig    `ini:"playlog,omitempty"`
	Random     RandomConfig     `ini:"random,omitempty"`
	Search     SearchConfig     `ini:"search,omitempty"`
	LastPlayed LastPlayedConfig `ini:"lastplayed,omitempty"`
}

func LoadUserConfig(name string, defaultConfig *UserConfig) (*UserConfig, error) {
	// TODO: central default ini first

	iniPath := os.Getenv(UserConfigEnv)

	if iniPath == "" {
		absPath, err := os.Executable()
		if err != nil {
			return defaultConfig, err
		}

		iniPath = filepath.Join(filepath.Dir(absPath), name+".ini")
	}

	if _, err := os.Stat(iniPath); os.IsNotExist(err) {
		return defaultConfig, nil
	}

	cfg, err := ini.Load(iniPath)
	if err != nil {
		return defaultConfig, err
	}

	err = cfg.StrictMapTo(defaultConfig)
	if err != nil {
		return defaultConfig, err
	}

	return defaultConfig, nil
}
