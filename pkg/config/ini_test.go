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

//nolint:gosec // Tests operate only on paths rooted in t.TempDir.
package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"gopkg.in/ini.v1"
)

func TestUpdateINIRetainsCommentsAndUnknownContent(t *testing.T) {
	path := filepath.Join(t.TempDir(), "app.ini")
	initial := "; file note\n[tui]\ntheme = default\n; root notes\ngames_folder = /mnt/old\n" +
		"; second root notes\ngames_folder = /mnt/second\n\n[custom]\n; keep this comment\nvalue = untouched\n"
	if err := os.WriteFile(path, []byte(initial), 0o600); err != nil {
		t.Fatal(err)
	}
	err := UpdateINI(path, func(file *ini.File) error {
		section, sectionErr := GetOrCreateSection(file, "tui")
		if sectionErr != nil {
			return sectionErr
		}
		SetKey(section, "theme", "nord")
		SetKey(section, "mouse", FormatBool(false))
		return SetShadowKeys(section, "games_folder", nil)
	})
	if err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	text := string(data)
	for _, expected := range []string{
		"; file note", "; root notes", "; second root notes", "theme = nord", "mouse = false",
		"[custom]", "; keep this comment", "value = untouched",
	} {
		if !strings.Contains(text, expected) {
			t.Fatalf("updated configuration lost %q:\n%s", expected, text)
		}
	}
	if strings.Contains(text, "games_folder") {
		t.Fatalf("cleared repeated key still present:\n%s", text)
	}
	entries, err := os.ReadDir(filepath.Dir(path))
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 1 {
		t.Fatalf("temporary files left behind: %v", entries)
	}
}

func TestUpdateINIRetainsRepeatedComments(t *testing.T) {
	path := filepath.Join(t.TempDir(), "app.ini")
	const note = "; shared root note"
	initial := "[systems]\n" + note + "\ngames_folder = /mnt/old\n" +
		note + "\ngames_folder = /mnt/second\n"
	if err := os.WriteFile(path, []byte(initial), 0o600); err != nil {
		t.Fatal(err)
	}
	// Repeated saves must neither lose identical notes nor add more copies.
	for range 2 {
		err := UpdateINI(path, func(file *ini.File) error {
			return SetShadowKeys(file.Section("systems"), "games_folder", []string{"/mnt/new", "/mnt/second"})
		})
		if err != nil {
			t.Fatal(err)
		}
		data, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		if count := strings.Count(string(data), note); count != 2 {
			t.Fatalf("retained %d identical comments, want 2:\n%s", count, data)
		}
	}
}

func TestUpdateINIRepeatedKeysRoundTrip(t *testing.T) {
	path := filepath.Join(t.TempDir(), "app.ini")
	if err := os.WriteFile(path, []byte("[systems]\n; roots\ngames_folder = /mnt/old\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	err := UpdateINI(path, func(file *ini.File) error {
		section, sectionErr := GetOrCreateSection(file, "systems")
		if sectionErr != nil {
			return sectionErr
		}
		return SetShadowKeys(section, "games_folder", []string{"/mnt/new", "/mnt/second"})
	})
	if err != nil {
		t.Fatal(err)
	}
	loaded, err := ini.ShadowLoad(path)
	if err != nil {
		t.Fatal(err)
	}
	values := loaded.Section("systems").Key("games_folder").ValueWithShadows()
	if len(values) != 2 || values[0] != "/mnt/new" || values[1] != "/mnt/second" {
		t.Fatalf("repeated values = %v", values)
	}
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(data), "; roots") {
		t.Fatalf("repeated key note lost:\n%s", data)
	}
}

func TestUpdateINIRequiresExistingFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "missing.ini")
	err := UpdateINI(path, func(*ini.File) error { return nil })
	if err == nil {
		t.Fatal("missing configuration accepted")
	}
	if _, statErr := os.Stat(path); !os.IsNotExist(statErr) {
		t.Fatalf("missing configuration was created: %v", statErr)
	}
}

func TestWriteINIAtomicallyUsesFilePermissions(t *testing.T) {
	path := filepath.Join(t.TempDir(), "app.ini")
	file := ini.Empty()
	file.Section("tui").Key("theme").SetValue("default")
	if err := WriteINIAtomically(path, file); err != nil {
		t.Fatal(err)
	}
	info, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}
	if info.Mode().Perm() != 0o644 {
		t.Fatalf("permissions = %v", info.Mode().Perm())
	}
}
