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
	"strings"

	"gopkg.in/ini.v1"
)

// UpdateINI loads the INI file at path with shadow keys enabled, lets update
// change only the keys it manages, re-attaches full-line comments the parser
// drops, and replaces the file atomically. Unknown sections and keys survive.
// A missing file is an error; callers create defaults first.
func UpdateINI(path string, update func(*ini.File) error) error {
	data, err := os.ReadFile(filepath.Clean(path))
	if err != nil {
		return fmt.Errorf("read configuration for update: %w", err)
	}
	file, err := ini.LoadSources(ini.LoadOptions{AllowShadows: true}, data)
	if err != nil {
		return fmt.Errorf("load configuration for update: %w", err)
	}
	if err := update(file); err != nil {
		return err
	}
	if err := RetainINIComments(data, file); err != nil {
		return err
	}
	return WriteINIAtomically(path, file)
}

// RetainINIComments keeps full-line notes that the INI parser replaced when
// loading shadow keys as file comments, even when a key is removed.
func RetainINIComments(original []byte, file *ini.File) error {
	var rendered strings.Builder
	if _, err := file.WriteTo(&rendered); err != nil {
		return fmt.Errorf("render configuration comments: %w", err)
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

// GetOrCreateSection returns the named section, creating it when missing.
func GetOrCreateSection(file *ini.File, name string) (*ini.Section, error) {
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

// SetKey sets a single-valued key, creating it when missing.
func SetKey(section *ini.Section, name, value string) {
	section.Key(name).SetValue(value)
}

// SetShadowKeys replaces a repeated key with values, keeping any comment
// attached to the first existing occurrence. An empty list removes the key and
// moves its comment to the section so the note is not lost.
func SetShadowKeys(section *ini.Section, name string, values []string) error {
	comment := ""
	if existing, err := section.GetKey(name); err == nil {
		comment = existing.Comment
	}
	section.DeleteKey(name)
	if len(values) == 0 {
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

// FormatBool renders a boolean the way mrext INI files spell it.
func FormatBool(value bool) string {
	if value {
		return "true"
	}
	return "false"
}

// WriteINIAtomically writes file to a temporary sibling and renames it over
// path, so a failure never leaves a truncated configuration behind.
func WriteINIAtomically(path string, file *ini.File) error {
	path = filepath.Clean(path)
	directory := filepath.Dir(path)
	base := strings.TrimSuffix(filepath.Base(path), filepath.Ext(path))
	temporary, err := os.CreateTemp(directory, "."+base+"-*.ini")
	if err != nil {
		return fmt.Errorf("create temporary configuration: %w", err)
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
		return fmt.Errorf("write temporary configuration: %w", err)
	}
	if err := temporary.Sync(); err != nil {
		return fmt.Errorf("sync temporary configuration: %w", err)
	}
	if err := temporary.Chmod(0o644); err != nil {
		return fmt.Errorf("set configuration permissions: %w", err)
	}
	if err := temporary.Close(); err != nil {
		return fmt.Errorf("close temporary configuration: %w", err)
	}
	// #nosec G703 -- destination is the explicit configuration path chosen by the caller.
	if err := os.Rename(temporaryPath, path); err != nil {
		return fmt.Errorf("replace configuration: %w", err)
	}
	removeTemporary = false
	return nil
}
