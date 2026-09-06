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

package main

import (
	"path/filepath"
	"testing"
)

func TestLocalRootRuntime(t *testing.T) {
	root := t.TempDir()
	cfg, manager, err := loadRuntime(root)
	if err != nil {
		t.Fatal(err)
	}
	if manager.Paths().SDRoot != root || manager.Paths().MenuFolder != filepath.Join(root, "_Games") ||
		cfg.Systems.GamesFolder[0] != root {
		t.Fatalf("paths=%+v roots=%v", manager.Paths(), cfg.Systems.GamesFolder)
	}
	if cfg.IniPath != filepath.Join(root, "Scripts", "gamesmenu.ini") {
		t.Fatal(cfg.IniPath)
	}
}

func TestInvalidCLIArguments(t *testing.T) {
	for _, args := range [][]string{{"refresh"}, {"unexpected"}, {"--invalid"}, {"--root"}} {
		if _, err := parseCLI(args); err == nil {
			t.Errorf("accepted %v", args)
		}
	}
	options, err := parseCLI([]string{"--root", t.TempDir()})
	if err != nil || options.root == "" {
		t.Fatalf("root: %+v, %v", options, err)
	}
}
