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
	"fmt"

	"github.com/wizzomafizzo/mrext/pkg/config"
	"gopkg.in/ini.v1"
)

// Settings is the staged copy of the values the Settings page edits.
type Settings struct {
	Theme   string
	Mouse   bool
	CRTMode bool
}

func SettingsFromConfig(cfg *config.UserConfig) Settings {
	return Settings{
		Theme:   cfg.TUI.Theme,
		Mouse:   cfg.TUI.Mouse,
		CRTMode: cfg.TUI.CRTMode,
	}
}

func (s *Settings) ApplyTo(cfg *config.UserConfig) {
	cfg.TUI.Theme = s.Theme
	cfg.TUI.Mouse = s.Mouse
	cfg.TUI.CRTMode = s.CRTMode
}

func (s *Settings) Validate() error {
	cfg := DefaultUserConfig()
	s.ApplyTo(cfg)
	return ValidateConfig(cfg)
}

func (s *Settings) Equal(other *Settings) bool {
	return s.Theme == other.Theme && s.Mouse == other.Mouse && s.CRTMode == other.CRTMode
}

// SaveSettings writes only the managed [tui] keys, keeping every other
// section, key and comment in the file.
func SaveSettings(path string, settings *Settings) error {
	if err := settings.Validate(); err != nil {
		return err
	}
	err := config.UpdateINI(path, func(file *ini.File) error {
		section, sectionErr := config.GetOrCreateSection(file, "tui")
		if sectionErr != nil {
			return fmt.Errorf("get TUI section: %w", sectionErr)
		}
		config.SetKey(section, "theme", settings.Theme)
		config.SetKey(section, "mouse", config.FormatBool(settings.Mouse))
		config.SetKey(section, "crt_mode", config.FormatBool(settings.CRTMode))
		return nil
	})
	if err != nil {
		return fmt.Errorf("save Games Menu settings: %w", err)
	}
	return nil
}
