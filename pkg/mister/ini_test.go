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

//nolint:gosec // Tests operate only within temporary roots.
package mister

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestExampleOnlyRootCreatesMainOnlyOnSave(t *testing.T) {
	root := t.TempDir()
	example := filepath.Join(root, ExampleIniFilename)
	original := "; example must stay untouched\n[MiSTer]\nvscale_mode=4\n"
	if err := os.WriteFile(example, []byte(original), 0o600); err != nil {
		t.Fatal(err)
	}
	inis, err := getAllMisterIniAt(root)
	if err != nil || len(inis) != 1 {
		t.Fatalf("inis=%v err=%v", inis, err)
	}
	main := &inis[0]
	if main.Id != 1 || main.DisplayName != "Main" || main.Filename != DefaultIniFilename {
		t.Fatalf("Main not exposed: %+v", main)
	}
	if loadErr := main.Load(); loadErr != nil {
		t.Fatal(loadErr)
	}
	if _, statErr := os.Stat(main.Path); !os.IsNotExist(statErr) {
		t.Fatalf("Load created Main: %v", statErr)
	}
	if setErr := main.SetKey("vscale_mode", "2"); setErr != nil {
		t.Fatal(setErr)
	}
	if saveErr := main.Save(); saveErr != nil {
		t.Fatal(saveErr)
	}
	data, err := os.ReadFile(main.Path)
	if err != nil || !strings.Contains(string(data), "[MiSTer]") || !strings.Contains(string(data), "vscale_mode=2") {
		t.Fatalf("invalid saved Main: %s %v", data, err)
	}
	data, err = os.ReadFile(example)
	if err != nil || string(data) != original {
		t.Fatalf("example changed: %s %v", data, err)
	}
}

func TestMissingMainKeepsAlternateFilesSeparate(t *testing.T) {
	root := t.TempDir()
	original := "; preserve this file\n[MiSTer]\nvideo_mode=8\n"
	for _, name := range []string{"MISTER_EXAMPLE.INI", "MiSTer_alt_1.ini", "MiSTer_alt_2.ini"} {
		if err := os.WriteFile(filepath.Join(root, name), []byte(original), 0o600); err != nil {
			t.Fatal(err)
		}
	}
	inis, err := getAllMisterIniAt(root)
	if err != nil || len(inis) != 3 {
		t.Fatalf("inis=%v err=%v", inis, err)
	}
	for i := range inis {
		if inis[i].Id != i+1 {
			t.Fatalf("noncontiguous IDs: %v", inis)
		}
	}
	if inis[0].Filename != DefaultIniFilename || inis[1].Filename != "MiSTer_alt_1.ini" {
		t.Fatalf("Main or alternate redirected: %v", inis)
	}
	if loadErr := inis[0].Load(); loadErr != nil {
		t.Fatal(loadErr)
	}
	if saveErr := inis[0].Save(); saveErr != nil {
		t.Fatal(saveErr)
	}
	for _, name := range []string{"MISTER_EXAMPLE.INI", "MiSTer_alt_1.ini", "MiSTer_alt_2.ini"} {
		data, readErr := os.ReadFile(filepath.Join(root, name))
		if readErr != nil || string(data) != original {
			t.Fatalf("changed %s: %s %v", name, data, readErr)
		}
	}
	listed, err := getAllMisterIniAt(root)
	if err != nil || len(listed) != 3 || listed[0].Path != inis[0].Path {
		t.Fatalf("saving Main changed selection identity: %v %v", listed, err)
	}
}
