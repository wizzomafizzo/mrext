package config

import (
	"os"
	"path/filepath"

	"gopkg.in/ini.v1"
)

// TODO: default values?

type altCoresConfig struct {
	LLAPI   []string `ini:"llapi,omitempty" delim:","`
	YC      []string `ini:"yc,omitempty" delim:","`
	DualRAM []string `ini:"dualram,omitempty" delim:","`
}

type launchSyncConfig struct{}

type playLogConfig struct {
	SaveEvery int `ini:"save_every,omitempty"`
}

type randomConfig struct{}

type searchConfig struct{}

type UserConfig struct {
	AltCores   altCoresConfig
	LaunchSync launchSyncConfig
	PlayLog    playLogConfig
	Random     randomConfig
	Search     searchConfig
}

func LoadUserConfig() (UserConfig, error) {
	// TODO: load up central default ini first
	var userConfig UserConfig

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
