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
	"slices"
	"sort"
	"strings"
	"sync"

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

type iniLayout struct {
	names       []string
	mu          sync.Mutex
	initialized bool
}

var currentIniLayout iniLayout

func (layout *iniLayout) accept(names []string) bool {
	layout.mu.Lock()
	defer layout.mu.Unlock()
	if !layout.initialized {
		layout.names = slices.Clone(names)
		layout.initialized = true
	}
	return slices.Equal(layout.names, names)
}

func GetAllMisterIni() ([]MisterIni, error) {
	return discoverMisterInis(config.SdFolder, &currentIniLayout)
}

func getAllMisterIniAt(root string) ([]MisterIni, error) {
	return discoverMisterInis(root, nil)
}

func discoverMisterInis(root string, layout *iniLayout) ([]MisterIni, error) {
	// os.ReadDir sorts names; MiSTer's cfg.cpp cfg_get_name first selects
	// three candidates in readdir order, THEN sorts that subset ignoring case.
	// #nosec G304 -- configured MiSTer root, or an injected temporary fixture root.
	dir, err := os.Open(root)
	if err != nil {
		return nil, fmt.Errorf("open MiSTer root: %w", err)
	}
	defer func() { _ = dir.Close() }()
	names, err := dir.Readdirnames(-1)
	if err != nil {
		return nil, fmt.Errorf("read MiSTer root: %w", err)
	}
	slots := alternateIniNames(names)
	if layout != nil && !layout.accept(slots) {
		return nil, errors.New("INI slot layout changed; restart MiSTer and Remote before editing settings")
	}
	return iniEntries(root, names, slots), nil
}

// MiSTer's C-locale strcasecmp folds ASCII bytes, not Unicode case variants.
func iniASCIILower(name string) string {
	lower := []byte(name)
	for i, b := range lower {
		if b >= 'A' && b <= 'Z' {
			lower[i] = b + ('a' - 'A')
		}
	}
	return string(lower)
}

func alternateIniNames(names []string) []string {
	slots := make([]string, 0, 3)
	for _, name := range names {
		lower := iniASCIILower(name)
		if strings.HasPrefix(lower, "mister_") && strings.HasSuffix(lower, ".ini") {
			// Match MiSTer's 64-byte name buffer, including its terminator.
			if len(name) > 63 {
				name = name[:63]
			}
			slots = append(slots, name)
			if len(slots) == 3 {
				break
			}
		}
	}
	sort.SliceStable(slots, func(i, j int) bool { return iniASCIILower(slots[i]) < iniASCIILower(slots[j]) })
	return slots
}

func iniEntries(root string, names, slots []string) []MisterIni {
	main := DefaultIniFilename
	for _, name := range names {
		if strings.EqualFold(name, DefaultIniFilename) {
			main = name
			break
		}
	}
	inis := []MisterIni{{Id: 1, DisplayName: "Main", Filename: main, Path: filepath.Join(root, main)}}
	for i, filename := range slots {
		// The example consumes a MiSTer slot but must never become editable.
		// Do not compact IDs when omitting it or an unusable filename.
		if strings.EqualFold(filename, ExampleIniFilename) || !strings.EqualFold(filepath.Ext(filename), ".ini") {
			continue
		}
		label := strings.TrimSuffix(filename[7:], filepath.Ext(filename))
		switch strings.ToLower(label) {
		case "alt_1":
			label = "Alt1"
		case "alt_2":
			label = "Alt2"
		case "alt_3":
			label = "Alt3"
		case "":
			label = " -- "
		}
		inis = append(inis, MisterIni{
			Id: i + 2, DisplayName: label, Filename: filename, Path: filepath.Join(root, filename),
		})
	}
	return inis
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

	return iniByID(inis, activeID)
}

func GetMisterIni(id int) (MisterIni, error) {
	inis, err := GetAllMisterIni()
	if err != nil {
		return MisterIni{}, err
	}

	return iniByID(inis, id)
}

func iniByID(inis []MisterIni, id int) (MisterIni, error) {
	for i := range inis {
		if inis[i].Id == id {
			return inis[i], nil
		}
	}
	return MisterIni{}, fmt.Errorf("INI slot %d is unavailable for editing", id)
}

// GetAllWithDefaultMisterIni includes Main even when its file does not exist.
func GetAllWithDefaultMisterIni() ([]MisterIni, error) {
	return GetAllMisterIni()
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
		// Reading settings must not create configuration. Save is the explicit
		// write boundary, including on installations containing only the example.
		mi.File = blank
		return nil
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
