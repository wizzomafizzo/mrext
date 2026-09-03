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

package mister

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/wizzomafizzo/mrext/pkg/config"
	"github.com/wizzomafizzo/mrext/pkg/games"
)

func TestGenerateMglUsesCatalog(t *testing.T) {
	t.Parallel()

	system, err := games.GetSystem("NES")
	if err != nil {
		t.Fatal(err)
	}
	got, err := GenerateMgl(&config.UserConfig{}, system, `/media/fat/games/NES/Mario & "Luigi".nes`, "")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(got, "<rbf>_Console/NES</rbf>") {
		t.Fatalf("catalog RBF missing: %s", got)
	}
	if !strings.Contains(got, `Mario &amp; &quot;Luigi&quot;.nes`) {
		t.Fatalf("path was not XML escaped: %s", got)
	}
}

func TestGenerateMglPreservesRBFOverride(t *testing.T) {
	t.Parallel()

	system, err := games.GetSystem("NES")
	if err != nil {
		t.Fatal(err)
	}
	cfg := &config.UserConfig{Systems: config.SystemsConfig{SetCore: []string{"nes:_Console/CustomNES"}}}
	got, err := GenerateMgl(cfg, system, "/media/fat/games/NES/Mario.nes", "")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(got, "<rbf>_Console/CustomNES</rbf>") {
		t.Fatalf("custom RBF missing: %s", got)
	}
}

func TestGenerateMglPreservesSetNameOverrideAndReset(t *testing.T) {
	t.Parallel()

	system, err := games.GetSystem("Jaguar")
	if err != nil {
		t.Fatal(err)
	}
	system.SetName = `Alt & "Jaguar"`
	system.SetNameSameDir = true
	got, err := GenerateMgl(&config.UserConfig{}, system, "/media/fat/games/Jaguar/game.jag", "")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(got, `<setname same_dir="1">Alt &amp; &quot;Jaguar&quot;</setname>`) {
		t.Fatalf("setname missing: %s", got)
	}
	if !strings.Contains(got, `<reset delay="1" hold="1"/>`) {
		t.Fatalf("reset missing: %s", got)
	}
}

func TestTrySetupArcadeCoresLinkUsesOutermostMenuFolder(t *testing.T) {
	t.Parallel()

	root := filepath.Join(t.TempDir(), "_Favorites")
	nested := filepath.Join(root, "Arcade", "Action")
	if err := os.MkdirAll(nested, 0o750); err != nil {
		t.Fatal(err)
	}
	launcher := filepath.Join(nested, "Game.mra")
	if err := os.WriteFile(launcher, nil, 0o600); err != nil {
		t.Fatal(err)
	}

	if err := TrySetupArcadeCoresLink(nested); err != nil {
		t.Fatal(err)
	}
	rootLink := filepath.Join(root, filepath.Base(config.ArcadeCoresFolder))
	info, err := os.Lstat(rootLink)
	if err != nil {
		t.Fatal(err)
	}
	if info.Mode()&os.ModeSymlink == 0 {
		t.Fatalf("%s is not a symlink", rootLink)
	}
	if _, err := os.Lstat(filepath.Join(nested, filepath.Base(config.ArcadeCoresFolder))); !os.IsNotExist(err) {
		t.Fatalf("nested cores link exists: %v", err)
	}

	if err := os.Remove(launcher); err != nil {
		t.Fatal(err)
	}
	if err := TrySetupArcadeCoresLink(nested); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Lstat(rootLink); !os.IsNotExist(err) {
		t.Fatalf("unused root cores link remains: %v", err)
	}
}

func TestGenerateMglPreservesHookOverride(t *testing.T) {
	t.Parallel()

	system, err := games.GetSystem("NES")
	if err != nil {
		t.Fatal(err)
	}
	override := "\t<file type=\"f\" path=\"custom\"/>\n"
	got, err := GenerateMgl(&config.UserConfig{}, system, "ignored.nes", override)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(got, override) || strings.Contains(got, "ignored.nes") {
		t.Fatalf("override changed: %s", got)
	}
}
