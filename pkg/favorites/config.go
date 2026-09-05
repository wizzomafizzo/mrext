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
	"errors"
	"fmt"
	"strings"

	"github.com/wizzomafizzo/mrext/pkg/config"
)

const DefaultConfigFile = `[favorites]
default_folder = _@Favorites
folder_name_contains = fav
create_default_folder = true
manage_arcade_core_links = true
hide_root_files = true
external_folder = /media/usb0
core_prefix =
; Add custom game roots by repeating this key. Built-in MiSTer roots remain enabled.
; games_folder = /media/network

[cores]
; Supported values: llapi, yc, or blank for standard cores.
all =

[tui]
theme = default
mouse = true
crt_mode = true
on_screen_keyboard = true
`

func DefaultUserConfig() *config.UserConfig {
	return &config.UserConfig{
		Favorites: config.FavoritesConfig{
			DefaultFolder:         "_@Favorites",
			ExternalFolder:        "/media/usb0",
			FolderNameContains:    []string{"fav"},
			CreateDefaultFolder:   true,
			ManageArcadeCoreLinks: true,
			HideRootFiles:         true,
		},
		TUI: config.TUIConfig{
			Theme:            "default",
			Mouse:            true,
			CRTMode:          true,
			OnScreenKeyboard: true,
		},
	}
}

func LoadConfig() (*config.UserConfig, error) {
	cfg, err := config.LoadUserConfig("favorites", DefaultUserConfig())
	return validateLoadedConfig(cfg, err)
}

func LoadConfigAt(iniPath, appPath string) (*config.UserConfig, error) {
	cfg, err := config.LoadUserConfigAt(iniPath, appPath, DefaultUserConfig())
	return validateLoadedConfig(cfg, err)
}

func validateLoadedConfig(cfg *config.UserConfig, err error) (*config.UserConfig, error) {
	if err != nil {
		return nil, fmt.Errorf("load user configuration: %w", err)
	}
	if err := ValidateConfig(cfg); err != nil {
		return nil, err
	}
	cfg.Systems.GamesFolder = append(cfg.Systems.GamesFolder, cfg.Favorites.GamesFolder...)
	return cfg, nil
}

func EnsureConfigFile(cfg *config.UserConfig) error {
	if cfg == nil {
		return errors.New("no Favorites configuration supplied")
	}
	if err := config.EnsureUserConfig(cfg.IniPath, DefaultConfigFile); err != nil {
		return fmt.Errorf("ensure Favorites configuration: %w", err)
	}
	return nil
}

func ValidateConfig(cfg *config.UserConfig) error {
	if cfg == nil {
		return errors.New("no Favorites configuration supplied")
	}
	if !strings.HasPrefix(cfg.Favorites.DefaultFolder, "_") {
		return errors.New("favorites.default_folder must start with an underscore")
	}
	if len(cfg.Favorites.FolderNameContains) == 0 {
		return errors.New("favorites.folder_name_contains must not be empty")
	}
	for _, fragment := range cfg.Favorites.FolderNameContains {
		if strings.TrimSpace(fragment) == "" {
			return errors.New("favorites.folder_name_contains contains an empty value")
		}
	}
	switch strings.ToLower(strings.TrimSpace(cfg.FavoritesCores.All)) {
	case "", "llapi", "yc":
	default:
		return fmt.Errorf("unsupported cores.all value: %s", cfg.FavoritesCores.All)
	}
	return nil
}
