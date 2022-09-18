package mister

import (
	"os"

	"gopkg.in/ini.v1"

	"github.com/wizzomafizzo/mrext/pkg/config"
)

func loadMisterIni() (*ini.File, error) {
	if _, err := os.Stat(config.MisterIniFile); os.IsNotExist(err) {
		return nil, err
	}

	return ini.Load(config.MisterIniFile)
}

func GetMisterIniOption(file *ini.File, name string) string {
	if file == nil {
		return ""
	}

	section := file.Section("MiSTer")
	if section == nil {
		return ""
	}

	key := section.Key(name)
	if key == nil {
		return ""
	}

	return key.String()
}

func RecentsOptionEnabled() bool {
	file, err := loadMisterIni()
	if err != nil {
		return false
	}

	option := GetMisterIniOption(file, "recents")
	if option == "" {
		return false
	}

	return option == "1"
}
