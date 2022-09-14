package ini

import (
	"os"
	"path/filepath"

	"gopkg.in/ini.v1"
)

type UserConfigLaunchSync struct{}

type UserConfigRandom struct{}

type UserConfigSearch struct{}

type UserConfig struct {
	AltCores   map[string]string
	LaunchSync *UserConfigLaunchSync
	Random     *UserConfigRandom
	Search     *UserConfigSearch
}

func loadConfig(name string) (*ini.File, *UserConfig, error) {
	var userConfig UserConfig
	appFolder := filepath.Dir(os.Args[0])
	iniPath := filepath.Join(appFolder, name+".ini")

	cfg, err := ini.Load(iniPath)
	if err != nil {
		return nil, nil, err
	}

	section := cfg.Section("cores")
	for _, key := range section.Keys() {
		if key.Name() == "all" {

		}
	}

	return cfg, &userConfig, nil
}

func LoadLaunchSyncConfig() (*ini.File, *UserConfig, error) {
	cfg, userConfig, err := loadConfig("launchsync")
	if err != nil {
		return nil, nil, err
	}

	userConfig.LaunchSync = &UserConfigLaunchSync{}

	return cfg, userConfig, nil
}
