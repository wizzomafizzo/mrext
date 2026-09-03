// mrext
// Copyright (c) 2026 mrext contributors.
// SPDX-License-Identifier: GPL-3.0-or-later
//
// This file is part of mrext.
//
// mrext is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// mrext is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with mrext. If not, see <http://www.gnu.org/licenses/>.

package mister

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/wizzomafizzo/mrext/pkg/config"
	"github.com/wizzomafizzo/mrext/pkg/utils"
	"gopkg.in/ini.v1"
)

// TODO: support getting/setting sections besides main

const ShadowDelimiter = ","

type MisterIni struct {
	File        *ini.File `json:"-"`
	DisplayName string    `json:"displayName"`
	Filename    string    `json:"filename"`
	Path        string    `json:"path"`
	Id          int       `json:"id"` //nolint:revive // Legacy public field name.
}

func GetAllMisterIni() ([]MisterIni, error) {
	var inis []MisterIni

	files, err := os.ReadDir(config.SdFolder)
	if err != nil {
		return nil, fmt.Errorf("read MiSTer root: %w", err)
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

	currentID := 1

	for _, filename := range iniFilenames {
		lower := strings.ToLower(filename)

		if strings.EqualFold(lower, DefaultIniFilename) {
			inis = append(inis, MisterIni{
				Id:          currentID,
				DisplayName: "Main",
				Filename:    filename,
				Path:        filepath.Join(config.SdFolder, filename),
			})

			currentID++
		} else if strings.HasPrefix(lower, "mister_") {
			iniFile := MisterIni{
				Id:          currentID,
				DisplayName: "",
				Filename:    filename,
				Path:        filepath.Join(config.SdFolder, filename),
			}

			iniFile.DisplayName = filename[7:]
			iniFile.DisplayName = strings.TrimSuffix(iniFile.DisplayName, filepath.Ext(iniFile.DisplayName))

			switch iniFile.DisplayName {
			case "":
				iniFile.DisplayName = " -- "
			case "alt_1":
				iniFile.DisplayName = "Alt1"
			case "alt_2":
				iniFile.DisplayName = "Alt2"
			case "alt_3":
				iniFile.DisplayName = "Alt3"
			}

			if len(iniFile.DisplayName) > 4 {
				iniFile.DisplayName = iniFile.DisplayName[0:4]
			}

			if len(inis) < 4 {
				inis = append(inis, iniFile)
			}

			currentID++
		}
	}

	return inis, nil
}

func GetActiveMisterIni() (MisterIni, error) {
	activeID, err := GetActiveIni()
	if err != nil {
		return MisterIni{}, err
	}

	if activeID == 0 {
		activeID = 1
	}

	inis, err := GetAllMisterIni()
	if err != nil {
		return MisterIni{}, err
	}

	if activeID < 1 || activeID > len(inis) {
		return MisterIni{}, fmt.Errorf("active ini id is out of range: %d (%d)", activeID, len(inis))
	}

	return inis[activeID-1], nil
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
	if _, err := iniFile.NewSection(MainIniSection); err != nil {
		return nil, fmt.Errorf("create MiSTer INI section: %w", err)
	}
	return iniFile, nil
}

func (mi *MisterIni) Load() error {
	ini.PrettyFormat = false
	ini.PrettyEqual = false

	// #nosec G703 -- path belongs to a discovered MiSTer INI file.
	if _, err := os.Stat(mi.Path); os.IsNotExist(err) {
		if mi.Filename != DefaultIniFilename {
			return fmt.Errorf("ini file does not exist: %s", mi.Path)
		}

		blank, err := blankMisterIniFile()
		if err != nil {
			return err
		}
		if err = blank.SaveTo(mi.Path); err != nil {
			return fmt.Errorf("save blank MiSTer INI: %w", err)
		}
	}

	iniFile, err := ini.ShadowLoad(mi.Path)
	if err != nil {
		return fmt.Errorf("load MiSTer INI: %w", err)
	}

	if !iniFile.HasSection(MainIniSection) {
		_, err = iniFile.NewSection(MainIniSection)
		if err != nil {
			return fmt.Errorf("create missing MiSTer INI section: %w", err)
		}
	}

	mi.File = iniFile

	return nil
}

func (mi *MisterIni) Save() error {
	if mi.File == nil {
		return errors.New("ini file is not loaded")
	}

	backupPath := mi.Path + ".backup"

	backupData, err := os.ReadFile(mi.Path)
	if os.IsNotExist(err) {
		// skip backup if file doesn't exist
		if saveErr := mi.File.SaveTo(mi.Path); saveErr != nil {
			return fmt.Errorf("save MiSTer INI: %w", saveErr)
		}
		return nil
	}
	if err != nil {
		return fmt.Errorf("read MiSTer INI for backup: %w", err)
	}

	// #nosec G306,G703 -- discovered MiSTer INI backup must remain world-readable.
	if err := os.WriteFile(backupPath, backupData, 0o644); err != nil {
		return fmt.Errorf("write MiSTer INI backup: %w", err)
	}

	if err := mi.File.SaveTo(mi.Path); err != nil {
		return fmt.Errorf("save MiSTer INI: %w", err)
	}
	return nil
}

func (*MisterIni) IsValidKey(key string) bool {
	return utils.Contains(ValidIniKeys, key)
}

func (*MisterIni) IsShadowedKey(key string) bool {
	return utils.Contains(ShadowedIniKeys, key)
}

func (mi *MisterIni) GetKey(key string) (string, error) {
	if mi.File == nil {
		return "", errors.New("ini file is not loaded")
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
	}
	return section.Key(key).Value(), nil
}

// SetKey a key to an absolute value, or delete it if value is empty. Supports
// shadowed keys delimited with a comma.
func (mi *MisterIni) SetKey(key, value string) error {
	if mi.File == nil {
		return errors.New("ini file is not loaded")
	}

	section := mi.File.Section(MainIniSection)
	if section == nil {
		return errors.New("ini file does not have a [MiSTer] section")
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
	}
	if value == "" {
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
			return fmt.Errorf("create shadowed INI key: %w", err)
		}

		for _, val := range vals[1:] {
			err := iniKey.AddShadow(val)
			if err != nil {
				return fmt.Errorf("append shadowed INI value: %w", err)
			}
		}
	} else {
		if section.HasKey(key) {
			section.Key(key).SetValue(value)
		} else {
			_, err := section.NewKey(key, value)
			if err != nil {
				return fmt.Errorf("create INI key: %w", err)
			}
		}
	}

	return nil
}

// AddKey sets a key to a value whether it exists or not and appends to any
// shadowed values.
func (mi *MisterIni) AddKey(key, value string) error {
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
	}
	return mi.SetKey(key, value)
}

// RemoveKey removes a key from the ini file.
func (mi *MisterIni) RemoveKey(key string) error {
	return mi.SetKey(key, "")
}

func RecentsOptionEnabled() (bool, error) {
	iniFile, err := GetActiveMisterIni()
	if err != nil {
		return false, fmt.Errorf("error getting active ini: %w", err)
	}

	err = iniFile.Load()
	if err != nil {
		return false, fmt.Errorf("error loading ini file: %w", err)
	}

	val, err := iniFile.GetKey(KeyRecents)
	if err != nil {
		return false, fmt.Errorf("error getting recents key: %w", err)
	}

	return val == "1", nil
}
