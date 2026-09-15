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

package games

import (
	"archive/zip"
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"slices"
	"testing"
)

// GetFiles is now a thin wrapper over WalkFiles, so comparing the two would
// prove nothing. Assert the expected set directly: a plain game, a game reached
// through a symlinked directory that links back to the root, and the one
// matching member of a ZIP.
func TestWalkFilesFindsGamesThroughSymlinksAndArchives(t *testing.T) {
	root, external := t.TempDir(), t.TempDir()
	for _, path := range []string{filepath.Join(root, "game.nes"), filepath.Join(external, "linked.nes")} {
		if err := os.WriteFile(path, nil, 0o600); err != nil {
			t.Fatal(err)
		}
	}
	if err := os.Symlink(external, filepath.Join(root, "linked")); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(root, filepath.Join(external, "cycle")); err != nil {
		t.Fatal(err)
	}
	// #nosec G304 -- archive is created only beneath t.TempDir.
	file, err := os.Create(filepath.Join(root, "pack.zip"))
	if err != nil {
		t.Fatal(err)
	}
	archive := zip.NewWriter(file)
	for _, name := range []string{"nested/zip.nes", "ignored.sfc"} {
		if _, createErr := archive.Create(name); createErr != nil {
			t.Fatal(createErr)
		}
	}
	if closeErr := archive.Close(); closeErr != nil {
		t.Fatal(closeErr)
	}
	if closeErr := file.Close(); closeErr != nil {
		t.Fatal(closeErr)
	}
	want := []string{
		filepath.Join(root, "game.nes"),
		filepath.Join(root, "linked", "linked.nes"),
		filepath.Join(root, "pack.zip", "nested", "zip.nes"),
	}
	cwd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	var got []string
	walkErr := WalkFiles("NES", root, func(path string) error { got = append(got, path); return nil })
	if walkErr != nil {
		t.Fatal(walkErr)
	}
	slices.Sort(got)
	slices.Sort(want)
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("scanned paths = %v, want %v", got, want)
	}
	// GetFiles must agree, since it is the same scan collected into a slice.
	legacy, err := GetFiles("NES", root)
	if err != nil {
		t.Fatal(err)
	}
	slices.Sort(legacy)
	if !reflect.DeepEqual(legacy, want) {
		t.Fatalf("GetFiles = %v, want %v", legacy, want)
	}
	if after, cwdErr := os.Getwd(); cwdErr != nil || after != cwd {
		t.Fatalf("scanner changed working directory: %q, %v", after, cwdErr)
	}
	failure := errors.New("stop scan")
	calls := 0
	err = WalkFiles("NES", root, func(string) error { calls++; return failure })
	if !errors.Is(err, failure) || calls != 1 {
		t.Fatalf("callback failure not propagated immediately: %v, calls=%d", err, calls)
	}
}

// NeoGeo ROM sets are folders. A folder named in romsets.xml is a game and
// is emitted as one, without descending into its ROM files; any other folder
// is walked as usual. Without romsets.xml nothing changes.
func TestWalkFilesEmitsNeoGeoSetFolders(t *testing.T) {
	root := t.TempDir()
	romsets := `<romsets><romset name="aof,aofa" altname="Art of Fighting"/></romsets>`
	if err := os.WriteFile(filepath.Join(root, "romsets.xml"), []byte(romsets), 0o600); err != nil {
		t.Fatal(err)
	}
	for _, dir := range []string{"aof", "misc"} {
		if err := os.Mkdir(filepath.Join(root, dir), 0o750); err != nil {
			t.Fatal(err)
		}
	}
	for _, file := range []string{"aof/p1.p1", "misc/kof98.neo", "mslug.neo"} {
		if err := os.WriteFile(filepath.Join(root, file), nil, 0o600); err != nil {
			t.Fatal(err)
		}
	}

	var got []string
	if err := WalkFiles("NeoGeo", root, func(path string) error { got = append(got, path); return nil }); err != nil {
		t.Fatal(err)
	}
	slices.Sort(got)
	want := []string{
		filepath.Join(root, "aof"),
		filepath.Join(root, "misc", "kof98.neo"),
		filepath.Join(root, "mslug.neo"),
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("got %v, want %v", got, want)
	}

	// The same tree for a system without set folders walks into "aof".
	got = nil
	if err := WalkFiles("NES", root, func(path string) error { got = append(got, path); return nil }); err != nil {
		t.Fatal(err)
	}
	if slices.Contains(got, filepath.Join(root, "aof")) {
		t.Fatalf("NES walk emitted the aof folder as a game: %v", got)
	}
}

// A set folder reached through a symlink is the game itself too, so it is
// emitted under its link name and not walked into.
func TestWalkFilesEmitsLinkedNeoGeoSetFolders(t *testing.T) {
	root, external := t.TempDir(), t.TempDir()
	romsets := `<romsets><romset name="aofa" altname="Art of Fighting (set 2)"/></romsets>`
	if err := os.WriteFile(filepath.Join(root, "romsets.xml"), []byte(romsets), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(external, "p1.p1"), nil, 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(external, filepath.Join(root, "aofa")); err != nil {
		t.Fatal(err)
	}

	var got []string
	if err := WalkFiles("NeoGeo", root, func(path string) error { got = append(got, path); return nil }); err != nil {
		t.Fatal(err)
	}
	want := []string{filepath.Join(root, "aofa")}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("got %v, want %v", got, want)
	}
}
