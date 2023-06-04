package mister

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

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
	// TODO: this needs to check the *active* ini, and i guess we need to assume it could be off
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

type IniFile struct {
	DisplayName string `json:"displayName"`
	Filename    string `json:"filename"`
	Path        string `json:"path"`
}

func ListMisterInis() ([]IniFile, error) {
	var inis []IniFile

	files, err := os.ReadDir(config.SdFolder)
	if err != nil {
		return nil, err
	}

	var iniFilenames []string

	for _, file := range files {
		if file.IsDir() {
			continue
		}

		if filepath.Ext(strings.ToLower(file.Name())) == ".ini" {
			iniFilenames = append(iniFilenames, file.Name())
		}
	}

	for _, filename := range iniFilenames {
		lower := strings.ToLower(filename)

		if lower == "mister.ini" {
			inis = append(inis, IniFile{
				DisplayName: "Main",
				Filename:    filename,
				Path:        filepath.Join(config.SdFolder, filename),
			})
		} else if strings.HasPrefix(lower, "mister_") {
			iniFile := IniFile{
				DisplayName: "",
				Filename:    filename,
				Path:        filepath.Join(config.SdFolder, filename),
			}

			iniFile.DisplayName = filename[7:]
			iniFile.DisplayName = strings.TrimSuffix(iniFile.DisplayName, filepath.Ext(iniFile.DisplayName))

			if iniFile.DisplayName == "" {
				iniFile.DisplayName = " -- "
			} else if iniFile.DisplayName == "alt_1" {
				iniFile.DisplayName = "Alt1"
			} else if iniFile.DisplayName == "alt_2" {
				iniFile.DisplayName = "Alt2"
			} else if iniFile.DisplayName == "alt_3" {
				iniFile.DisplayName = "Alt3"
			}

			if len(iniFile.DisplayName) > 4 {
				iniFile.DisplayName = iniFile.DisplayName[0:4]
			}

			if len(inis) < 4 {
				inis = append(inis, iniFile)
			}
		}
	}

	return inis, nil
}
