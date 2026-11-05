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

//nolint:gosec // Tests only operate on temporary fixture paths.
package gamesmenu

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/wizzomafizzo/mrext/pkg/config"
)

func TestCustomMenuFolder(t *testing.T) {
	root := t.TempDir()
	cfg := DefaultUserConfig()
	cfg.Systems.GamesFolder = []string{root}
	cfg.GamesMenu.MenuFolder = "_My Games"
	m, err := NewManagerWithPaths(cfg, RootedPaths(root))
	if err != nil {
		t.Fatal(err)
	}
	if want := filepath.Join(root, "_My Games"); m.Paths().MenuFolder != want {
		t.Fatalf("MenuFolder = %q, want %q", m.Paths().MenuFolder, want)
	}
	put(t, filepath.Join(root, "games", "NES", "game.nes"), "")
	put(t, filepath.Join(root, "_Games", "_SNES", "old.mgl"), "old")
	generate(t, m, discover(t, m))
	read(t, filepath.Join(root, "_My Games", "_NES", "game.mgl"))
	if read(t, filepath.Join(root, "_Games", "_SNES", "old.mgl")) != "old" {
		t.Fatal("old _Games folder changed")
	}
}

func TestMenuFolderValidation(t *testing.T) {
	valid := []string{"_Games", "_My Games", "_Juegos"}
	for _, name := range valid {
		if err := ValidateMenuFolder(name); err != nil {
			t.Errorf("%q: %v", name, err)
		}
	}
	invalid := []string{"Games", "_", "_a/b", `_a\b`, "_Console", "_ARCADE", "_@Favorites", "../_x"}
	for _, name := range invalid {
		if err := ValidateMenuFolder(name); err == nil {
			t.Errorf("%q accepted", name)
		}
	}

	cfg := DefaultUserConfig()
	cfg.GamesMenu.MenuFolder = "_Console"
	if err := ValidateConfig(cfg); err == nil {
		t.Fatal("ValidateConfig accepted _Console")
	}
	if _, err := NewManagerWithPaths(cfg, RootedPaths(t.TempDir())); err == nil {
		t.Fatal("NewManagerWithPaths accepted _Console")
	}
}

func TestCustomMenuFolderRefusesCoreFolder(t *testing.T) {
	root := t.TempDir()
	cfg := DefaultUserConfig()
	cfg.Systems.GamesFolder = []string{root}
	cfg.GamesMenu.MenuFolder = "_Cores"
	m, err := NewManagerWithPaths(cfg, RootedPaths(root))
	if err != nil {
		t.Fatal(err)
	}
	put(t, filepath.Join(root, "_Cores", "NES_20260101.rbf"), "core")
	put(t, filepath.Join(root, "_Cores", "_Sub", "keep.txt"), "keep")
	if _, err := m.Generate(nil, nil); err == nil || !strings.Contains(err.Error(), "contains cores") {
		t.Fatalf("Generate err = %v", err)
	}
	if err := m.RemoveAll(); err == nil {
		t.Fatal("RemoveAll accepted a core folder")
	}
	if read(t, filepath.Join(root, "_Cores", "_Sub", "keep.txt")) != "keep" {
		t.Fatal("core folder changed")
	}
}

func TestGenerateExclude(t *testing.T) {
	m := fixture(t)
	m.exclude = compileExclude([]string{"*(beta)*", "*[h]*", " ", "skip?.nes", "excluded.zip"})
	nes := filepath.Join(m.paths.SDRoot, "games", "NES")
	put(t, filepath.Join(nes, "Game (USA).nes"), "")
	put(t, filepath.Join(nes, "Game (Beta).nes"), "")
	put(t, filepath.Join(nes, "Hack [h].nes"), "")
	put(t, filepath.Join(nes, "Hash.nes"), "")
	put(t, filepath.Join(nes, "Skip1.nes"), "")
	put(t, filepath.Join(nes, "Skip10.nes"), "")
	createZip(t, filepath.Join(nes, "pack.zip"), "Zipped.nes", "dir/Zipped (Beta).nes")
	createZip(t, filepath.Join(nes, "excluded.zip"), "Hidden.nes")
	generate(t, m, discover(t, m))

	output := filepath.Join(m.paths.MenuFolder, "_NES")
	var got []string
	err := filepath.WalkDir(output, func(name string, entry os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !entry.IsDir() {
			rel, _ := filepath.Rel(output, name)
			got = append(got, filepath.ToSlash(rel))
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	want := "Game (USA).mgl,Hash.mgl,Skip10.mgl,Zipped.mgl"
	if strings.Join(got, ",") != want {
		t.Fatalf("shortcuts = %v, want %s", got, want)
	}
}

func TestLoadGamesMenuOptions(t *testing.T) {
	t.Setenv(config.SharedTUIConfigEnv, filepath.Join(t.TempDir(), "tui.ini"))
	root := t.TempDir()
	filename := filepath.Join(root, IniFilename)
	put(t, filename, "[gamesmenu]\nmenu_folder = _Juegos\nexclude = *(Beta)*\nexclude = *(USA, Europe)*\n")
	cfg, err := LoadConfigAt(filename, filepath.Join(root, AppFilename))
	if err != nil {
		t.Fatal(err)
	}
	if cfg.GamesMenu.MenuFolder != "_Juegos" ||
		strings.Join(cfg.GamesMenu.Exclude, "|") != "*(Beta)*|*(USA, Europe)*" {
		t.Fatalf("GamesMenu = %+v", cfg.GamesMenu)
	}
}
