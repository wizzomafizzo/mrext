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
	"testing"
)

func TestNamesPythonRules(t *testing.T) {
	filename := filepath.Join(t.TempDir(), "names.txt")
	names, err := LoadNames(filename)
	if err != nil || names.Replace("NES") != "NES" {
		t.Fatalf("missing: %v, %v", names, err)
	}
	put(t, filename, "ignored\n NES : First/Name\nnes: second\nempty: \nchars: <>:\"\\|?*\n")
	names, err = LoadNames(filename)
	if err != nil {
		t.Fatal(err)
	}
	checks := []struct{ input, want string }{
		{"nEs", "First & Name"},
		{"empty", ""},
		{"chars", "        "},
		{"unknown", "unknown"},
		{" NES ", " NES "},
	}
	for _, check := range checks {
		if got := names.Replace(check.input); got != check.want {
			t.Errorf("Replace(%q)=%q want %q", check.input, got, check.want)
		}
	}
	if got := names.MenuPath("NES", "empty", "unknown"); got != "_First & Name/_/_unknown" {
		t.Fatalf("menu path: %q", got)
	}
}

func TestExtensionRulesUseCatalogMGLSlots(t *testing.T) {
	for _, folder := range menuSystems() {
		if folder.Key == "Arcade" {
			t.Fatal("Arcade has no MGL slot")
		}
		if matchSystem(extensionRules(folder.Systems), "custom.mgl") != nil {
			t.Fatalf("MGL accepted for %s", folder.Key)
		}
		if folder.Key == "Pico8" && matchSystem(extensionRules(folder.Systems), "game.P8.PNG") == nil {
			t.Fatal("multi-dot catalog suffix not accepted")
		}
	}
}
