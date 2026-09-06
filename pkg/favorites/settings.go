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
	err := config.UpdateINI(path, func(file *ini.File) error {
		favoritesSection, sectionErr := config.GetOrCreateSection(file, "favorites")
		if sectionErr != nil {
			return fmt.Errorf("get favorites section: %w", sectionErr)
		}
		config.SetKey(favoritesSection, "default_folder", settings.DefaultFolder)
		config.SetKey(favoritesSection, "folder_name_contains", strings.Join(settings.FolderNameContains, ","))
		config.SetKey(favoritesSection, "create_default_folder", config.FormatBool(settings.CreateDefaultFolder))
		config.SetKey(favoritesSection, "manage_arcade_core_links", config.FormatBool(settings.ManageArcadeCoreLinks))
		config.SetKey(favoritesSection, "hide_root_files", config.FormatBool(settings.HideRootFiles))
		config.SetKey(favoritesSection, "external_folder", settings.ExternalFolder)
		config.SetKey(favoritesSection, "core_prefix", settings.CorePrefix)
		if shadowErr := config.SetShadowKeys(
			favoritesSection, "games_folder", settings.GamesFolders,
		); shadowErr != nil {
			return fmt.Errorf("set games folders: %w", shadowErr)
		}

		coresSection, sectionErr := config.GetOrCreateSection(file, "cores")
		if sectionErr != nil {
			return fmt.Errorf("get cores section: %w", sectionErr)
		}
		config.SetKey(coresSection, "all", settings.AlternateCore)

		tuiSection, sectionErr := config.GetOrCreateSection(file, "tui")
		if sectionErr != nil {
			return fmt.Errorf("get TUI section: %w", sectionErr)
		}
		config.SetKey(tuiSection, "theme", settings.Theme)
		config.SetKey(tuiSection, "mouse", config.FormatBool(settings.Mouse))
		config.SetKey(tuiSection, "crt_mode", config.FormatBool(settings.CRTMode))
		config.SetKey(tuiSection, "on_screen_keyboard", config.FormatBool(settings.OnScreenKeyboard))
		return nil
	})
	if err != nil {
		return fmt.Errorf("save Favorites settings: %w", err)
	}
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
