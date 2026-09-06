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

package gamesmenu

import (
	"errors"
	"fmt"

	"github.com/wizzomafizzo/mrext/pkg/config"
	"github.com/wizzomafizzo/mrext/pkg/tui"
)

// DefaultConfigFile is written beside the binary on the first interactive run
// when no gamesmenu.ini exists. The Python script had no configuration file,
// so every key here is optional and only affects the TUI or extra game roots.
const DefaultConfigFile = `[tui]
theme = default
mouse = true
crt_mode = true

[systems]
; Add custom game roots by repeating this key. Built-in MiSTer roots remain enabled.
; games_folder = /media/network
`

func DefaultUserConfig() *config.UserConfig {
	return &config.UserConfig{
		TUI: config.TUIConfig{
			Theme:   "default",
			Mouse:   true,
			CRTMode: true,
		},
	}
}

func LoadConfig() (*config.UserConfig, error) {
	cfg, err := config.LoadUserConfig("gamesmenu", DefaultUserConfig())
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
	return cfg, nil
}

// EnsureConfigFile creates the commented default configuration when missing
// and never touches an existing file.
func EnsureConfigFile(cfg *config.UserConfig) error {
	if cfg == nil {
		return errors.New("no Games Menu configuration supplied")
	}
	if err := config.EnsureUserConfig(cfg.IniPath, DefaultConfigFile); err != nil {
		return fmt.Errorf("ensure Games Menu configuration: %w", err)
	}
	return nil
}

func ValidateConfig(cfg *config.UserConfig) error {
	if cfg == nil {
		return errors.New("no Games Menu configuration supplied")
	}
	if _, ok := tui.AvailableThemes[cfg.TUI.Theme]; !ok {
		return fmt.Errorf("unknown tui.theme value: %s", cfg.TUI.Theme)
	}
	return nil
}
