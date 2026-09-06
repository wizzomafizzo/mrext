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
	"path/filepath"
	"strings"
	"testing"
)

func TestSettingsPreserveExistingConfig(t *testing.T) {
	root := t.TempDir()
	filename := filepath.Join(root, IniFilename)
	cfg, err := LoadConfigAt(filename, filepath.Join(root, AppFilename))
	if err != nil {
		t.Fatal(err)
	}
	if configErr := EnsureConfigFile(cfg); configErr != nil {
		t.Fatal(configErr)
	}
	original := read(t, filename) + "games_folder = " + root + "\n\n[custom]\n; retained comment\nunknown = retained\n"
	put(t, filename, original)
	if configErr := EnsureConfigFile(cfg); configErr != nil {
		t.Fatal(configErr)
	}
	if read(t, filename) != original {
		t.Fatal("EnsureConfigFile overwrote configuration")
	}
	settings := SettingsFromConfig(cfg)
	settings.Mouse = false
	settings.Theme = "dracula"
	if saveErr := SaveSettings(filename, &settings); saveErr != nil {
		t.Fatal(saveErr)
	}
	data := read(t, filename)
	retained := []string{"; games_folder = /media/network", "[systems]", "[custom]", "; retained comment", "retained"}
	for _, text := range retained {
		if !strings.Contains(data, text) {
			t.Errorf("lost %q: %s", text, data)
		}
	}
	loaded, err := LoadConfigAt(filename, filepath.Join(root, AppFilename))
	if err != nil {
		t.Fatal(err)
	}
	if loaded.TUI.Mouse || loaded.TUI.Theme != "dracula" ||
		len(loaded.Systems.GamesFolder) != 1 || loaded.Systems.GamesFolder[0] != root {
		t.Fatalf("reload: %+v", loaded)
	}
	settings.Theme = "unknown theme"
	if err := SaveSettings(filename, &settings); err == nil {
		t.Fatal("invalid settings accepted")
	}
	if read(t, filename) != data {
		t.Fatal("invalid settings modified file")
	}
}
