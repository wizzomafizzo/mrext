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
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/wizzomafizzo/mrext/pkg/mister"
)

func TestCleanUpLegacyGeneratedZipAndUnknownTargets(t *testing.T) {
	m := fixture(t)
	folder := filepath.Join(m.paths.SDRoot, "games", "NES")
	game := filepath.Join(folder, "valid & game.nes")
	put(t, game, "")
	createZip(t, filepath.Join(folder, "set.zip"), "valid.nes")
	put(t, filepath.Join(folder, "corrupt.zip"), "invalid zip")
	generate(t, m, discover(t, m))
	output := filepath.Join(m.paths.MenuFolder, "_NES")
	legacy := func(name, target string) {
		put(t, filepath.Join(output, name+".mgl"),
			`<mistergamedescription><rbf>_Console/NES</rbf><file path="`+target+`"/></mistergamedescription>`)
	}
	legacy("legacy valid", game)
	legacy("legacy missing", filepath.Join(folder, "gone & game.nes"))
	legacy("zip missing", filepath.Join(folder, "set.zip", "gone.nes"))
	legacy("archive missing", filepath.Join(folder, "absent.zip", "game.nes"))
	legacy("corrupt archive", filepath.Join(folder, "corrupt.zip", "game.nes"))
	legacy("relative", "games/NES/missing.nes")
	legacy("empty target", "")
	put(t, filepath.Join(output, "trailing.mgl"), `<mistergamedescription><file path="`+
		folder+`/gone.nes"/></mistergamedescription>garbage`)
	put(t, filepath.Join(output, "core.mgl"), `<mistergamedescription><rbf>_Console/NES</rbf></mistergamedescription>`)
	put(t, filepath.Join(output, "malformed.mgl"), `<mistergamedescription><file`)
	put(t, filepath.Join(output, "wrong-root.mgl"), `<other><file path="`+folder+`/gone.nes"/></other>`)
	// If any target is ambiguous, never delete a multi-file custom launcher.
	put(t, filepath.Join(output, "multi.mgl"), `<mistergamedescription><file path="`+folder+
		`/gone.nes"/><file path="relative.nes"/></mistergamedescription>`)
	entries := discover(t, m)
	system := matchSystem(extensionRules(entries[0].Systems), "deleted.nes")
	data, err := mister.GenerateMgl(m.cfg, system, filepath.Join(folder, "deleted.nes"), "")
	if err != nil {
		t.Fatal(err)
	}
	put(t, filepath.Join(output, "_empty", "_nested", "generated.mgl"), data)
	// Do not follow menu symlinks or remove content outside _Games.
	outside := t.TempDir()
	put(t, filepath.Join(outside, "keep.mgl"), data)
	if linkErr := os.Symlink(outside, filepath.Join(output, "link")); linkErr != nil {
		t.Fatal(linkErr)
	}
	var last CleanupProgress
	result, err := m.CleanUp(func(p CleanupProgress) { last = p })
	if err != nil {
		t.Fatal(err)
	}
	if result.Removed != 4 || result.PrunedFolders != 2 || result.Unreadable != 7 {
		t.Fatalf("result: %+v", result)
	}
	if last.Checked != result.Checked || last.Removed != result.Removed {
		t.Fatalf("progress: %+v, result: %+v", last, result)
	}
	retained := []string{
		"valid & game", "valid", "legacy valid", "corrupt archive", "relative",
		"core", "malformed", "wrong-root", "multi", "empty target", "trailing",
	}
	for _, name := range retained {
		if _, err := os.Stat(filepath.Join(output, name+".mgl")); err != nil {
			t.Errorf("retained %s: %v", name, err)
		}
	}
	if _, err := os.Stat(filepath.Join(outside, "keep.mgl")); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(output); err != nil {
		t.Fatal("system folder pruned:", err)
	}
}

func TestCleanUpKeepsAmbiguousPathsBeforeCheckingExistence(t *testing.T) {
	m := fixture(t)
	folder := filepath.Join(m.paths.SDRoot, "games", "NES")
	targets := []string{
		folder + "/missing.zip/../game.nes",
		folder + "/missing.zip/dir/../game.nes",
		folder + "/missing.zip//absolute.nes",
		folder + `/missing.zip/dir\game.nes`,
		folder + "/./missing.zip/game.nes",
		folder + "/./missing.nes",
	}
	for index, target := range targets {
		put(t, filepath.Join(m.paths.MenuFolder, "_NES", fmt.Sprintf("keep%d.mgl", index)),
			`<mistergamedescription><file path="`+target+`"/></mistergamedescription>`)
	}
	result, err := m.CleanUp(nil)
	if err != nil || result.Checked != len(targets) || result.Removed != 0 || result.Unreadable != len(targets) {
		t.Fatalf("ambiguous shortcuts not retained: %+v, %v", result, err)
	}
}

func TestCleanUpStatErrorsAreNotMissing(t *testing.T) {
	m := fixture(t)
	put(t, filepath.Join(m.paths.SDRoot, "not-directory"), "")
	output := filepath.Join(m.paths.MenuFolder, "_NES", "keep.mgl")
	put(t, output, `<mistergamedescription><file path="`+m.paths.SDRoot+
		`/not-directory/game.nes"/></mistergamedescription>`)
	result, err := m.CleanUp(nil)
	if err != nil {
		t.Fatal(err)
	}
	if result.Removed != 0 || result.Unreadable != 1 {
		t.Fatalf("result: %+v", result)
	}
	if _, err := os.Stat(output); err != nil {
		t.Fatal(err)
	}
}
