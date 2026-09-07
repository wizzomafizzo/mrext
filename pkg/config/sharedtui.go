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

package config

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/ini.v1"
)

// SharedTUIConfigPath returns the shared [tui] file, honouring the override.
func SharedTUIConfigPath() string {
	if path := os.Getenv(SharedTUIConfigEnv); path != "" {
		return path
	}
	return SharedTUIConfigFile
}

// LoadSharedTUI reads the shared [tui] settings over the given defaults. Keys
// the file does not declare are left alone, so this layers cleanly under an
// app's own INI: app INI, then this, then the built-in defaults.
//
// A missing file is not an error. It is the normal case.
func LoadSharedTUI(defaults TUIConfig) (TUIConfig, error) {
	path := SharedTUIConfigPath()
	if _, err := os.Stat(path); err != nil {
		if os.IsNotExist(err) {
			return defaults, nil
		}
		return defaults, fmt.Errorf("stat shared interface configuration: %w", err)
	}
	file, err := ini.Load(filepath.Clean(path))
	if err != nil {
		return defaults, fmt.Errorf("load shared interface configuration: %w", err)
	}
	shared := defaults
	if err := file.Section("tui").MapTo(&shared); err != nil {
		return defaults, fmt.Errorf("map shared interface configuration: %w", err)
	}
	return shared, nil
}

// SaveSharedTUI writes the shared [tui] settings, creating the file and its
// directory when needed. Unknown sections and keys in an existing file are
// preserved, as everywhere else in mrext.
func SaveSharedTUI(settings TUIConfig) error {
	path := SharedTUIConfigPath()
	if err := os.MkdirAll(filepath.Dir(path), 0o750); err != nil {
		return fmt.Errorf("create shared configuration directory: %w", err)
	}

	file := ini.Empty()
	if _, err := os.Stat(path); err == nil {
		loaded, loadErr := ini.Load(filepath.Clean(path))
		if loadErr != nil {
			return fmt.Errorf("load shared interface configuration: %w", loadErr)
		}
		file = loaded
	}

	section, err := GetOrCreateSection(file, "tui")
	if err != nil {
		return err
	}
	SetKey(section, "theme", settings.Theme)
	SetKey(section, "mouse", FormatBool(settings.Mouse))
	SetKey(section, "crt_mode", FormatBool(settings.CRTMode))
	SetKey(section, "on_screen_keyboard", FormatBool(settings.OnScreenKeyboard))

	return WriteINIAtomically(path, file)
}

// RemoveLocalTUIKeys deletes the four managed [tui] keys from an app's own INI,
// leaving every other key, section and comment alone. An app stops overriding
// the shared file only when its own keys are gone.
func RemoveLocalTUIKeys(path string) error {
	if _, err := os.Stat(path); err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return fmt.Errorf("stat user configuration: %w", err)
	}
	return UpdateINI(path, func(file *ini.File) error {
		if !file.HasSection("tui") {
			// No [tui] section: already following the shared file.
			return nil
		}
		section := file.Section("tui")
		for _, key := range []string{"theme", "mouse", "crt_mode", "on_screen_keyboard"} {
			section.DeleteKey(key)
		}
		if len(section.Keys()) == 0 {
			file.DeleteSection("tui")
		}
		return nil
	})
}
