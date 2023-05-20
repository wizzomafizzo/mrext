package mister

import (
	"fmt"
	"os"

	"gopkg.in/ini.v1"

	"github.com/wizzomafizzo/mrext/pkg/config"
)

func LoadMisterIni() (*ini.File, error) {
	if _, err := os.Stat(config.MisterIniFile); os.IsNotExist(err) {
		return nil, err
	}

	iniFile, err := ini.Load(config.MisterIniFile)
	if err != nil {
		return nil, err
	}

	if !iniFile.HasSection("MiSTer") {
		return nil, fmt.Errorf("mister.ini does not have a [MiSTer] section")
	}

	ini.PrettyFormat = false
	ini.PrettyEqual = false

	return iniFile, nil
}

func SaveMisterIni(iniFile *ini.File) error {
	return iniFile.SaveTo(config.MisterIniFile)
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
	file, err := LoadMisterIni()
	if err != nil {
		return false
	}

	option := GetMisterIniOption(file, "recents")
	if option == "" {
		return false
	}

	return option == "1"
}
