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

package bgm

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// Document is a minimal model of Python's configparser file format. It keeps
// sections and keys in file order so unknown entries survive a rewrite, and
// renders exactly like ConfigParser.write() so existing bgm.ini files keep
// their shape.
type Document struct {
	sections []*iniSection
}

type iniSection struct {
	name    string
	entries []*iniEntry
}

type iniEntry struct {
	key   string
	parts []string
}

var (
	iniSectionPattern = regexp.MustCompile(`^\[(.+)\]`)
	iniOptionPattern  = regexp.MustCompile(`^(.*?)\s*[=:]\s*(.*)$`)
)

// ParseINI reads configparser syntax: case-sensitive sections, lowercased
// keys, "=" or ":" delimiters, full-line "#" or ";" comments and indented
// continuation lines. Malformed lines are skipped where Python would raise.
func ParseINI(data string) *Document {
	doc := &Document{}
	var section *iniSection
	var option *iniEntry
	indentLevel := 0
	for _, rawLine := range strings.Split(data, "\n") {
		line := strings.TrimRight(rawLine, "\r")
		value := strings.TrimSpace(line)
		if strings.HasPrefix(value, "#") || strings.HasPrefix(value, ";") {
			continue
		}
		if value == "" {
			if option != nil {
				option.parts = append(option.parts, "")
			}
			continue
		}
		indent := len(line) - len(strings.TrimLeft(line, " \t\f\v"))
		if section != nil && option != nil && indent > indentLevel {
			option.parts = append(option.parts, value)
			continue
		}
		indentLevel = indent
		if match := iniSectionPattern.FindStringSubmatch(value); match != nil {
			section = doc.section(match[1], true)
			option = nil
			continue
		}
		if section == nil {
			continue
		}
		match := iniOptionPattern.FindStringSubmatch(value)
		if len(match) < 3 || match[1] == "" {
			continue
		}
		option = section.entry(strings.ToLower(strings.TrimRight(match[1], " \t")), true)
		option.parts = []string{strings.TrimSpace(match[2])}
	}
	return doc
}

// ReadINI parses the file at path. A missing file yields an empty document,
// matching ConfigParser.read().
func ReadINI(path string) (*Document, error) {
	data, err := os.ReadFile(filepath.Clean(path))
	if err != nil {
		if os.IsNotExist(err) {
			return &Document{}, nil
		}
		return nil, fmt.Errorf("read BGM configuration: %w", err)
	}
	return ParseINI(string(data)), nil
}

// Get returns the value of key in section with "%%" unescaped, as
// ConfigParser.get() does with its default interpolation.
func (d *Document) Get(section, key string) (string, bool) {
	sec := d.section(section, false)
	if sec == nil {
		return "", false
	}
	entry := sec.entry(strings.ToLower(key), false)
	if entry == nil {
		return "", false
	}
	return strings.ReplaceAll(entry.value(), "%%", "%"), true
}

// Set stores value under section/key, creating either when missing and
// escaping "%" so Python can still read the file.
func (d *Document) Set(section, key, value string) {
	entry := d.section(section, true).entry(strings.ToLower(key), true)
	entry.parts = []string{strings.ReplaceAll(value, "%", "%%")}
}

// HasSection reports whether the section exists.
func (d *Document) HasSection(section string) bool {
	return d.section(section, false) != nil
}

// String renders the document the way ConfigParser.write() does.
func (d *Document) String() string {
	var out strings.Builder
	for _, section := range d.sections {
		_, _ = fmt.Fprintf(&out, "[%s]\n", section.name)
		for _, entry := range section.entries {
			_, _ = fmt.Fprintf(&out, "%s = %s\n", entry.key, strings.ReplaceAll(entry.value(), "\n", "\n\t"))
		}
		_, _ = out.WriteString("\n")
	}
	return out.String()
}

// WriteFile atomically replaces path with the rendered document.
func (d *Document) WriteFile(path string) error {
	directory := filepath.Dir(path)
	temporary, err := os.CreateTemp(directory, ".bgm-*.ini")
	if err != nil {
		return fmt.Errorf("create temporary BGM configuration: %w", err)
	}
	temporaryPath := temporary.Name()
	removeTemporary := true
	defer func() {
		_ = temporary.Close()
		if removeTemporary {
			_ = os.Remove(temporaryPath)
		}
	}()
	if _, err := temporary.WriteString(d.String()); err != nil {
		return fmt.Errorf("write temporary BGM configuration: %w", err)
	}
	if err := temporary.Sync(); err != nil {
		return fmt.Errorf("sync temporary BGM configuration: %w", err)
	}
	if err := temporary.Chmod(0o644); err != nil {
		return fmt.Errorf("set BGM configuration permissions: %w", err)
	}
	if err := temporary.Close(); err != nil {
		return fmt.Errorf("close temporary BGM configuration: %w", err)
	}
	// #nosec G703 -- destination is the explicit BGM configuration path.
	if err := os.Rename(temporaryPath, filepath.Clean(path)); err != nil {
		return fmt.Errorf("replace BGM configuration: %w", err)
	}
	removeTemporary = false
	return nil
}

func (d *Document) section(name string, create bool) *iniSection {
	for _, section := range d.sections {
		if section.name == name {
			return section
		}
	}
	if !create {
		return nil
	}
	section := &iniSection{name: name}
	d.sections = append(d.sections, section)
	return section
}

func (s *iniSection) entry(key string, create bool) *iniEntry {
	for _, entry := range s.entries {
		if entry.key == key {
			return entry
		}
	}
	if !create {
		return nil
	}
	entry := &iniEntry{key: key}
	s.entries = append(s.entries, entry)
	return entry
}

// value joins continuation lines like configparser's _join_multiline_values.
func (e *iniEntry) value() string {
	return strings.TrimRight(strings.Join(e.parts, "\n"), " \t\n\r\f\v")
}
