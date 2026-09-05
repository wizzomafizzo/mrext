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

package favorites

import (
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"github.com/wizzomafizzo/mrext/pkg/config"
	"gopkg.in/ini.v1"
)

//nolint:govet // Field order groups settings by their INI sections.
type Settings struct {
	DefaultFolder         string
	ExternalFolder        string
	CorePrefix            string
	FolderNameContains    []string
	GamesFolders          []string
	AlternateCore         string
	Theme                 string
	CreateDefaultFolder   bool
	ManageArcadeCoreLinks bool
	HideRootFiles         bool
	Mouse                 bool
	CRTMode               bool
	OnScreenKeyboard      bool
}

func SettingsFromConfig(cfg *config.UserConfig) Settings {
	return Settings{
		DefaultFolder:         cfg.Favorites.DefaultFolder,
		ExternalFolder:        cfg.Favorites.ExternalFolder,
		CorePrefix:            cfg.Favorites.CorePrefix,
		FolderNameContains:    slices.Clone(cfg.Favorites.FolderNameContains),
		GamesFolders:          slices.Clone(cfg.Favorites.GamesFolder),
		AlternateCore:         strings.ToLower(strings.TrimSpace(cfg.FavoritesCores.All)),
		Theme:                 cfg.TUI.Theme,
		CreateDefaultFolder:   cfg.Favorites.CreateDefaultFolder,
		ManageArcadeCoreLinks: cfg.Favorites.ManageArcadeCoreLinks,
		HideRootFiles:         cfg.Favorites.HideRootFiles,
		Mouse:                 cfg.TUI.Mouse,
		CRTMode:               cfg.TUI.CRTMode,
		OnScreenKeyboard:      cfg.TUI.OnScreenKeyboard,
	}
}

func (s *Settings) ApplyTo(cfg *config.UserConfig) {
	cfg.Favorites.DefaultFolder = s.DefaultFolder
	cfg.Favorites.ExternalFolder = s.ExternalFolder
	cfg.Favorites.CorePrefix = s.CorePrefix
	cfg.Favorites.FolderNameContains = slices.Clone(s.FolderNameContains)
	cfg.Favorites.GamesFolder = slices.Clone(s.GamesFolders)
	cfg.FavoritesCores.All = s.AlternateCore
	cfg.TUI.Theme = s.Theme
	cfg.Favorites.CreateDefaultFolder = s.CreateDefaultFolder
	cfg.Favorites.ManageArcadeCoreLinks = s.ManageArcadeCoreLinks
	cfg.Favorites.HideRootFiles = s.HideRootFiles
	cfg.TUI.Mouse = s.Mouse
	cfg.TUI.CRTMode = s.CRTMode
	cfg.TUI.OnScreenKeyboard = s.OnScreenKeyboard
}

func (s *Settings) Validate() error {
	cfg := DefaultUserConfig()
	s.ApplyTo(cfg)
	return ValidateConfig(cfg)
}

func (s *Settings) Equal(other *Settings) bool {
	return s.DefaultFolder == other.DefaultFolder &&
		s.ExternalFolder == other.ExternalFolder &&
		s.CorePrefix == other.CorePrefix &&
		slices.Equal(s.FolderNameContains, other.FolderNameContains) &&
		slices.Equal(s.GamesFolders, other.GamesFolders) &&
		s.AlternateCore == other.AlternateCore &&
		s.Theme == other.Theme &&
		s.CreateDefaultFolder == other.CreateDefaultFolder &&
		s.ManageArcadeCoreLinks == other.ManageArcadeCoreLinks &&
		s.HideRootFiles == other.HideRootFiles &&
		s.Mouse == other.Mouse &&
		s.CRTMode == other.CRTMode &&
		s.OnScreenKeyboard == other.OnScreenKeyboard
}

func SaveSettings(path string, settings *Settings) error {
	if err := settings.Validate(); err != nil {
		return err
	}
	data, err := os.ReadFile(filepath.Clean(path))
	if err != nil {
		return fmt.Errorf("read Favorites configuration for update: %w", err)
	}
	file, err := ini.LoadSources(ini.LoadOptions{AllowShadows: true}, data)
	if err != nil {
		return fmt.Errorf("load Favorites configuration for update: %w", err)
	}
	favoritesSection, err := getOrCreateSection(file, "favorites")
	if err != nil {
		return err
	}
	setKey(favoritesSection, "default_folder", settings.DefaultFolder)
	setKey(favoritesSection, "folder_name_contains", strings.Join(settings.FolderNameContains, ","))
	setKey(favoritesSection, "create_default_folder", formatBool(settings.CreateDefaultFolder))
	setKey(favoritesSection, "manage_arcade_core_links", formatBool(settings.ManageArcadeCoreLinks))
	setKey(favoritesSection, "hide_root_files", formatBool(settings.HideRootFiles))
	setKey(favoritesSection, "external_folder", settings.ExternalFolder)
	setKey(favoritesSection, "core_prefix", settings.CorePrefix)
	if shadowErr := setShadowKeys(favoritesSection, "games_folder", settings.GamesFolders); shadowErr != nil {
		return shadowErr
	}

	coresSection, err := getOrCreateSection(file, "cores")
	if err != nil {
		return err
	}
	setKey(coresSection, "all", settings.AlternateCore)

	tuiSection, err := getOrCreateSection(file, "tui")
	if err != nil {
		return err
	}
	setKey(tuiSection, "theme", settings.Theme)
	setKey(tuiSection, "mouse", formatBool(settings.Mouse))
	setKey(tuiSection, "crt_mode", formatBool(settings.CRTMode))
	setKey(tuiSection, "on_screen_keyboard", formatBool(settings.OnScreenKeyboard))

	if err := retainINIComments(data, file); err != nil {
		return err
	}
	return writeINIAtomically(path, file)
}

// The INI parser replaces attached comments when loading shadow keys. Keep
// otherwise lost full-line notes as file comments, even when a key is removed.
func retainINIComments(original []byte, file *ini.File) error {
	var rendered strings.Builder
	if _, err := file.WriteTo(&rendered); err != nil {
		return fmt.Errorf("render Favorites configuration comments: %w", err)
	}
	present := make(map[string]bool)
	for _, line := range strings.Split(rendered.String(), "\n") {
		present[strings.TrimSpace(line)] = true
	}
	for _, line := range strings.Split(string(original), "\n") {
		line = strings.TrimSpace(line)
		if (strings.HasPrefix(line, ";") || strings.HasPrefix(line, "#")) && !present[line] {
			section := file.Section(ini.DefaultSection)
			section.Comment = strings.TrimSpace(section.Comment + "\n" + line)
			present[line] = true
		}
	}
	return nil
}

func getOrCreateSection(file *ini.File, name string) (*ini.Section, error) {
	section, err := file.GetSection(name)
	if err == nil {
		return section, nil
	}
	section, err = file.NewSection(name)
	if err != nil {
		return nil, fmt.Errorf("create %s configuration section: %w", name, err)
	}
	return section, nil
}

func setKey(section *ini.Section, name, value string) {
	section.Key(name).SetValue(value)
}

func setShadowKeys(section *ini.Section, name string, values []string) error {
	comment := ""
	if existing, err := section.GetKey(name); err == nil {
		comment = existing.Comment
	}
	section.DeleteKey(name)
	if len(values) == 0 {
		// Keep notes on cleared roots without leaving an active empty key.
		if comment != "" {
			section.Comment = strings.TrimSpace(section.Comment + "\n" + comment)
		}
		return nil
	}
	key, err := section.NewKey(name, values[0])
	if err != nil {
		return fmt.Errorf("create %s setting: %w", name, err)
	}
	key.Comment = comment
	for _, value := range values[1:] {
		if err := key.AddShadow(value); err != nil {
			return fmt.Errorf("add %s setting: %w", name, err)
		}
	}
	return nil
}

func formatBool(value bool) string {
	if value {
		return "true"
	}
	return "false"
}

func writeINIAtomically(path string, file *ini.File) error {
	directory := filepath.Dir(path)
	temporary, err := os.CreateTemp(directory, ".favorites-*.ini")
	if err != nil {
		return fmt.Errorf("create temporary Favorites configuration: %w", err)
	}
	temporaryPath := temporary.Name()
	removeTemporary := true
	defer func() {
		_ = temporary.Close()
		if removeTemporary {
			_ = os.Remove(temporaryPath)
		}
	}()
	if _, err := file.WriteTo(temporary); err != nil {
		return fmt.Errorf("write temporary Favorites configuration: %w", err)
	}
	if err := temporary.Sync(); err != nil {
		return fmt.Errorf("sync temporary Favorites configuration: %w", err)
	}
	if err := temporary.Chmod(0o644); err != nil {
		return fmt.Errorf("set Favorites configuration permissions: %w", err)
	}
	if err := temporary.Close(); err != nil {
		return fmt.Errorf("close temporary Favorites configuration: %w", err)
	}
	// #nosec G703 -- destination is the explicit Favorites configuration path.
	if err := os.Rename(temporaryPath, filepath.Clean(path)); err != nil {
		return fmt.Errorf("replace Favorites configuration: %w", err)
	}
	removeTemporary = false
	return nil
}

func ParseListSetting(value string) []string {
	parts := strings.Split(value, ",")
	values := make([]string, 0, len(parts))
	for _, part := range parts {
		if value := strings.TrimSpace(part); value != "" {
			values = append(values, value)
		}
	}
	return values
}
