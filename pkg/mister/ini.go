package mister

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/wizzomafizzo/mrext/pkg/utils"
	"gopkg.in/ini.v1"

	"github.com/wizzomafizzo/mrext/pkg/config"
)

// TODO: support getting/setting sections besides main

const ShadowDelimiter = ","

type MisterIni struct {
	Id          int       `json:"id"`
	DisplayName string    `json:"displayName"`
	Filename    string    `json:"filename"`
	Path        string    `json:"path"`
	File        *ini.File `json:"-"`
}

func GetAllMisterIni() ([]MisterIni, error) {
	var inis []MisterIni

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

	currentId := 1

	for _, filename := range iniFilenames {
		lower := strings.ToLower(filename)

		if lower == strings.ToLower(DefaultIniFilename) {
			inis = append(inis, MisterIni{
				Id:          currentId,
				DisplayName: "Main",
				Filename:    filename,
				Path:        filepath.Join(config.SdFolder, filename),
			})

			currentId++
		} else if strings.HasPrefix(lower, "mister_") {
			iniFile := MisterIni{
				Id:          currentId,
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

			currentId++
		}
	}

	return inis, nil
}

func GetActiveMisterIni() (MisterIni, error) {
	activeId, err := GetActiveIni()
	if err != nil {
		return MisterIni{}, err
	}

	if activeId == 0 {
		activeId = 1
	}

	inis, err := GetAllMisterIni()
	if err != nil {
		return MisterIni{}, err
	}

	if activeId < 1 || activeId > len(inis) {
		return MisterIni{}, fmt.Errorf("active ini id is out of range: %d (%d)", activeId, len(inis))
	}

	return inis[activeId-1], nil
}

func GetMisterIni(id int) (MisterIni, error) {
	inis, err := GetAllMisterIni()
	if err != nil {
		return MisterIni{}, err
	}

	if id < 1 || id > len(inis) {
		return MisterIni{}, fmt.Errorf("ini id is out of range: %d (%d)", id, len(inis))
	}

	return inis[id-1], nil
}

// GetAllWithDefaultMisterIni returns all ini files, setting up a default one if none exist.
func GetAllWithDefaultMisterIni() ([]MisterIni, error) {
	inis, err := GetAllMisterIni()
	if err != nil {
		return nil, err
	}

	if len(inis) == 0 {
		inis = append(inis, MisterIni{
			Id:          1,
			DisplayName: "Main",
			Filename:    DefaultIniFilename,
			Path:        filepath.Join(config.SdFolder, DefaultIniFilename),
		})
	}

	return inis, nil
}

func blankMisterIniFile() (*ini.File, error) {
	iniFile := ini.Empty()
	_, err := iniFile.NewSection(MainIniSection)
	return iniFile, err
}

func (mi *MisterIni) Load() error {
	ini.PrettyFormat = false
	ini.PrettyEqual = false

	if _, err := os.Stat(mi.Path); os.IsNotExist(err) {
		if mi.Filename == DefaultIniFilename {
			blank, err := blankMisterIniFile()
			if err != nil {
				return err
			}

			err = blank.SaveTo(mi.Path)
			if err != nil {
				return err
			}
		} else {
			return fmt.Errorf("ini file does not exist: %s", mi.Path)
		}
	}

	iniFile, err := ini.ShadowLoad(mi.Path)
	if err != nil {
		return err
	}

	if !iniFile.HasSection(MainIniSection) {
		_, err = iniFile.NewSection(MainIniSection)
		if err != nil {
			return err
		}
	}

	mi.File = iniFile

	return nil
}

func (mi *MisterIni) Save() error {
	if mi.File == nil {
		return fmt.Errorf("ini file is not loaded")
	}

	backupPath := fmt.Sprintf("%s.backup", mi.Path)

	backupData, err := os.ReadFile(mi.Path)
	if os.IsNotExist(err) {
		// skip backup if file doesn't exist
		return mi.File.SaveTo(mi.Path)
	} else if err != nil {
		return err
	}

	err = os.WriteFile(backupPath, backupData, 0644)
	if err != nil {
		return err
	}

	return mi.File.SaveTo(mi.Path)
}

func (mi *MisterIni) IsValidKey(key string) bool {
	return utils.Contains(ValidIniKeys, key)
}

func (mi *MisterIni) IsShadowedKey(key string) bool {
	return utils.Contains(ShadowedIniKeys, key)
}

func (mi *MisterIni) GetKey(key string) (string, error) {
	if mi.File == nil {
		return "", fmt.Errorf("ini file is not loaded")
	}

	section := mi.File.Section(MainIniSection)
	if section == nil {
		return "", nil
	}

	if strings.HasPrefix(key, "__") {
		return "", nil
	}

	if !mi.IsValidKey(key) {
		return "", fmt.Errorf("invalid ini key: %s", key)
	}

	if !section.HasKey(key) {
		return "", nil
	}

	if mi.IsShadowedKey(key) {
		vals := section.Key(key).StringsWithShadows(ShadowDelimiter)
		return strings.Join(vals, ShadowDelimiter), nil
	} else {
		return section.Key(key).Value(), nil
	}
}

// SetKey a key to an absolute value, or delete it if value is empty. Supports
// shadowed keys delimited with a comma.
func (mi *MisterIni) SetKey(key string, value string) error {
	if mi.File == nil {
		return fmt.Errorf("ini file is not loaded")
	}

	section := mi.File.Section(MainIniSection)
	if section == nil {
		return fmt.Errorf("ini file does not have a [MiSTer] section")
	}

	if strings.HasPrefix(key, "__") {
		return nil
	}

	if !mi.IsValidKey(key) {
		return fmt.Errorf("invalid ini key: %s", key)
	}

	if section.HasKey(key) && value == "" {
		section.DeleteKey(key)
		return nil
	} else if value == "" {
		return nil
	}

	if mi.IsShadowedKey(key) {
		if section.HasKey(key) {
			section.DeleteKey(key)
		}

		vals := strings.Split(value, ShadowDelimiter)

		if len(vals) == 0 {
			return nil
		}

		iniKey, err := section.NewKey(key, vals[0])
		if err != nil {
			return err
		}

		for _, val := range vals[1:] {
			err := iniKey.AddShadow(val)
			if err != nil {
				return err
			}
		}
	} else {
		if section.HasKey(key) {
			section.Key(key).SetValue(value)
		} else {
			_, err := section.NewKey(key, value)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

// AddKey sets a key to a value whether it exists or not and appends to any
// shadowed values.
func (mi *MisterIni) AddKey(key string, value string) error {
	currentValue, err := mi.GetKey(key)
	if err != nil {
		return err
	}

	if currentValue == "" {
		return mi.SetKey(key, value)
	}

	if mi.IsShadowedKey(key) {
		vals := strings.Split(currentValue, ShadowDelimiter)
		vals = append(vals, value)
		return mi.SetKey(key, strings.Join(vals, ShadowDelimiter))
	} else {
		return mi.SetKey(key, value)
	}
}

// RemoveKey removes a key from the ini file.
func (mi *MisterIni) RemoveKey(key string) error {
	return mi.SetKey(key, "")
}

func RecentsOptionEnabled() (bool, error) {
	iniFile, err := GetActiveMisterIni()
	if err != nil {
		return false, nil
	}

	err = iniFile.Load()
	if err != nil {
		return false, fmt.Errorf("error loading ini file: %s", err)
	}

	val, err := iniFile.GetKey(KeyRecents)
	if err != nil {
		return false, fmt.Errorf("error getting recents key: %s", err)
	}

	return val == "1", nil
}

func GetInisWithout(key string, value string) ([]MisterIni, error) {
	inis, err := GetAllMisterIni()
	if err != nil {
		return nil, err
	}

	var without []MisterIni
	for _, mi := range inis {
		err := mi.Load()
		if err != nil {
			return nil, err
		}

		val, err := mi.GetKey(key)
		if err != nil {
			return nil, err
		}

		if val == "" {
			without = append(without, mi)
			continue
		}

		if mi.IsShadowedKey(key) {
			vals := strings.Split(val, ShadowDelimiter)
			if !utils.Contains(vals, value) {
				without = append(without, mi)
			}
		} else {
			if val != value {
				without = append(without, mi)
			}
		}
	}

	return without, nil
}
