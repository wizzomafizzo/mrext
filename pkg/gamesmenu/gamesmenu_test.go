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
	"archive/zip"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"sync"
	"testing"

	"github.com/wizzomafizzo/mrext/pkg/mister"
)

func fixture(t *testing.T) *Manager {
	t.Helper()
	root := t.TempDir()
	cfg := DefaultUserConfig()
	cfg.Systems.GamesFolder = []string{root}
	manager, err := NewManagerWithPaths(cfg, RootedPaths(root))
	if err != nil {
		t.Fatal(err)
	}
	return manager
}

func put(t *testing.T, filename, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(filename), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filename, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

func read(t *testing.T, filename string) string {
	t.Helper()
	data, err := os.ReadFile(filename)
	if err != nil {
		t.Fatal(err)
	}
	return string(data)
}

func createZip(t *testing.T, filename string, members ...string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(filename), 0o755); err != nil {
		t.Fatal(err)
	}
	file, err := os.Create(filename)
	if err != nil {
		t.Fatal(err)
	}
	writer := zip.NewWriter(file)
	for _, member := range members {
		output, err := writer.Create(member)
		if err != nil {
			t.Fatal(err)
		}
		if _, err := output.Write([]byte("game")); err != nil {
			t.Fatal(err)
		}
	}
	if err := writer.Close(); err != nil {
		t.Fatal(err)
	}
	if err := file.Close(); err != nil {
		t.Fatal(err)
	}
}

func discover(t *testing.T, manager *Manager) []Entry {
	t.Helper()
	entries, err := manager.Discover()
	if err != nil {
		t.Fatal(err)
	}
	return entries
}

func generate(t *testing.T, manager *Manager, entries []Entry) GenerateResult {
	t.Helper()
	result, err := manager.Generate(entries, nil)
	if err != nil {
		t.Fatal(err)
	}
	if result.Scan.Failed != 0 {
		t.Fatalf("scan errors: %v", result.Scan.Errors)
	}
	return result
}

func TestDiscoverKeysRootsCaseAndSelection(t *testing.T) {
	m := fixture(t)
	root := m.paths.SDRoot
	external := filepath.Join(root, "external")
	m.cfg.Systems.GamesFolder = append(m.cfg.Systems.GamesFolder, root, external)
	put(t, filepath.Join(root, "games", "SNES", "a.sfc"), "")
	put(t, filepath.Join(root, "SNES", "b.sfc"), "")
	put(t, filepath.Join(external, "games", "SNES", "c.sfc"), "")
	put(t, filepath.Join(root, "games", "atari2600", "a.a26"), "")
	put(t, filepath.Join(root, "games", "NeoGeo", "a.neo"), "")
	entries := discover(t, m)
	if len(entries) != 3 {
		t.Fatalf("entries: %+v", entries)
	}
	for _, entry := range entries {
		if !entry.Selected {
			t.Fatalf("missing menu should select %s", entry.Key)
		}
	}
	put(t, filepath.Join(m.paths.MenuFolder, "_NeoGeo", "keep.mgl"), "custom")
	put(t, filepath.Join(m.paths.MenuFolder, "_ATARI2600", "keep.mgl"), "custom")
	entries = discover(t, m)
	if entries[0].MenuFolder != "_ATARI2600" || !entries[0].Selected ||
		entries[1].MenuFolder != "_NeoGeo" || !entries[1].Selected || entries[2].Selected {
		t.Fatalf("case/selection: %+v", entries)
	}
	want := []string{
		filepath.Join(root, "games", "SNES"), filepath.Join(root, "SNES"), filepath.Join(external, "games", "SNES"),
	}
	if !reflect.DeepEqual(entries[2].Paths, want) {
		t.Fatalf("paths %v, want %v", entries[2].Paths, want)
	}
}

func TestGenerateSharedFolderAndNeverOverwrite(t *testing.T) {
	m := fixture(t)
	folder := filepath.Join(m.paths.SDRoot, "games", "GAMEBOY")
	for _, name := range []string{"Mario & Luigi.gb", "color.gbc", "duck.bin"} {
		put(t, filepath.Join(folder, name), "")
	}
	entries := discover(t, m)
	result := generate(t, m, entries)
	if result.Scan.Created != 3 {
		t.Fatalf("result: %+v", result)
	}
	output := filepath.Join(m.paths.MenuFolder, "_GAMEBOY")
	checks := []struct{ name, fragment string }{
		{"Mario & Luigi", "<rbf>_Console/Gameboy</rbf>"},
		{"color", "<setname>GBC</setname>"},
		{"duck", `index="1"`},
	}
	for _, check := range checks {
		data := read(t, filepath.Join(output, check.name+".mgl"))
		if !strings.Contains(data, check.fragment) || !strings.Contains(data, "../../../../..") {
			t.Errorf("MGL: %s", data)
		}
	}
	if !strings.Contains(read(t, filepath.Join(output, "Mario & Luigi.mgl")), "&amp;") {
		t.Fatal("path not XML escaped")
	}
	custom := filepath.Join(output, "color.mgl")
	put(t, custom, "custom bytes")
	result = generate(t, m, entries)
	if result.Scan.Created != 0 || result.Scan.Skipped != 3 || read(t, custom) != "custom bytes" {
		t.Fatalf("overwrite: %+v", result)
	}
}

func TestGenerateZipFlatteningNamesDotfilesAndOverride(t *testing.T) {
	m := fixture(t)
	put(t, m.paths.NamesFile, "NES:Famicom/Disk\ndir:Renamed\n")
	var err error
	m.names, err = LoadNames(m.paths.NamesFile)
	if err != nil {
		t.Fatal(err)
	}
	m.cfg.Systems.SetCore = []string{"NES:_Console/Custom"}
	packs := filepath.Join(m.paths.SDRoot, "games", "NES", "Packs")
	createZip(t, filepath.Join(packs, "a.zip"),
		"dir/sub/game.nes", "top.nes", "readme.txt", ".hidden.nes", "dir/.hidden/game.nes")
	createZip(t, filepath.Join(packs, "b.zip"), "top.nes")
	put(t, filepath.Join(packs, "invalid.zip"), "not zip")
	put(t, filepath.Join(packs, ".hidden.nes"), "")
	put(t, filepath.Join(packs, ".folder", "game.nes"), "")
	result := generate(t, m, discover(t, m))
	if result.Scan.Created != 2 || result.Scan.Skipped != 1 {
		t.Fatalf("result: %+v", result)
	}
	output := filepath.Join(m.paths.MenuFolder, "_Famicom & Disk", "_Packs")
	data := read(t, filepath.Join(output, "_Renamed", "_sub", "game.mgl"))
	if !strings.Contains(data, "<rbf>_Console/Custom</rbf>") || !strings.Contains(data, "a.zip/dir/sub/game.nes") {
		t.Fatalf("MGL: %s", data)
	}
}

func TestGenerateProcessesFilesBeforeSubfolders(t *testing.T) {
	m := fixture(t)
	folder := filepath.Join(m.paths.SDRoot, "games", "NES")
	put(t, filepath.Join(folder, "A", "game.nes"), "loose game")
	createZip(t, filepath.Join(folder, "z.zip"), "A/game.nes")
	entries := discover(t, m)
	result := generate(t, m, entries)
	if result.Scan.Created != 1 || result.Scan.Skipped != 1 {
		t.Fatalf("result: %+v", result)
	}
	shortcut := filepath.Join(m.paths.MenuFolder, "_NES", "_A", "game.mgl")
	data := read(t, shortcut)
	if !strings.Contains(data, "z.zip/A/game.nes") {
		t.Fatalf("Python processes the parent ZIP before the loose subfolder game: %s", data)
	}
	result = generate(t, m, entries)
	if result.Scan.Created != 0 || result.Scan.Skipped != 2 || read(t, shortcut) != data {
		t.Fatalf("regeneration changed the collision winner: %+v", result)
	}
}

func TestGenerateNoncanonicalZIPMembersAndCleanup(t *testing.T) {
	m := fixture(t)
	archive := filepath.Join(m.paths.SDRoot, "games", "NES", "Packs", "set.zip")
	createZip(t, archive, "./game.nes", "Folder//second.nes", "Folder/./third.nes", "Folder//Sub/fourth.nes",
		"./.hidden.nes", "Folder/./.hidden/game.nes")
	result := generate(t, m, discover(t, m))
	if result.Scan.Created != 4 {
		t.Fatalf("result: %+v", result)
	}
	// Python prefixes literal dot and interior empty components before joining
	// output paths. Preserve those menu folders and the exact ZIP member keys.
	checks := []struct{ output, member string }{
		{"_./game.mgl", "./game.nes"},
		{"_Folder/second.mgl", "Folder//second.nes"},
		{"_Folder/_./third.mgl", "Folder/./third.nes"},
		{"_Folder/_/_Sub/fourth.mgl", "Folder//Sub/fourth.nes"},
	}
	for _, check := range checks {
		data := read(t, filepath.Join(m.paths.MenuFolder, "_NES", "_Packs", check.output))
		if !strings.Contains(data, archive+"/"+check.member) {
			t.Fatalf("ZIP member identity changed: %s", data)
		}
	}
	cleanup, err := m.CleanUp(nil)
	if err != nil || cleanup.Checked != 4 || cleanup.Removed != 0 || cleanup.Unreadable != 0 {
		t.Fatalf("valid members not recognized: %+v, %v", cleanup, err)
	}
	createZip(t, archive, "./game.nes")
	cleanup, err = m.CleanUp(nil)
	if err != nil || cleanup.Removed != 3 || cleanup.Unreadable != 0 {
		t.Fatalf("missing members not removed: %+v, %v", cleanup, err)
	}
	if _, err := os.Stat(filepath.Join(m.paths.MenuFolder, "_NES", "_Packs", "_.", "game.mgl")); err != nil {
		t.Fatal("cleanup removed retained dot-prefixed member:", err)
	}
}

func TestGenerateDeselectedRemovalAndRemoveAll(t *testing.T) {
	m := fixture(t)
	put(t, filepath.Join(m.paths.SDRoot, "games", "GBA", "game.gba"), "")
	put(t, filepath.Join(m.paths.MenuFolder, "_gba", "keep.mgl"), "keep")
	put(t, filepath.Join(m.paths.MenuFolder, "_NES", "old.mgl"), "old")
	put(t, filepath.Join(m.paths.MenuFolder, "User Folder", "note"), "old")
	put(t, filepath.Join(m.paths.MenuFolder, "notes.txt"), "keep")
	result := generate(t, m, discover(t, m))
	if len(result.Removed) != 2 {
		t.Fatalf("removed: %v", result.Removed)
	}
	if read(t, filepath.Join(m.paths.MenuFolder, "_gba", "keep.mgl")) != "keep" ||
		read(t, filepath.Join(m.paths.MenuFolder, "notes.txt")) != "keep" {
		t.Fatal("lost retained content")
	}
	if err := m.RemoveAll(); err != nil {
		t.Fatal(err)
	}
	if m.MenuExists() {
		t.Fatal("menu remains")
	}
}

func TestGenerateFollowsSymlinkedFoldersOnce(t *testing.T) {
	m := fixture(t)
	library := filepath.Join(m.paths.SDRoot, "library")
	put(t, filepath.Join(library, "game.sfc"), "")
	if err := os.MkdirAll(filepath.Join(m.paths.SDRoot, "games"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(library, filepath.Join(m.paths.SDRoot, "games", "SNES")); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(library, filepath.Join(library, "loop")); err != nil {
		t.Fatal(err)
	}
	result := generate(t, m, discover(t, m))
	if result.Scan.Created != 1 {
		t.Fatalf("result: %+v", result)
	}
}

func TestGenerateRejectsZipTraversalAndOutputSymlink(t *testing.T) {
	m := fixture(t)
	createZip(t, filepath.Join(m.paths.SDRoot, "games", "NES", "games.zip"),
		"../escape.nes", "/absolute.nes", "inside/../escape2.nes", ".././escape3.nes", `inside\escape4.nes`, "safe.nes")
	result, err := m.Generate(discover(t, m), nil)
	if err != nil {
		t.Fatal(err)
	}
	if result.Scan.Failed != 5 || result.Scan.Created != 1 {
		t.Fatalf("result: %+v", result)
	}
	outside := t.TempDir()
	if removeErr := os.RemoveAll(filepath.Join(m.paths.MenuFolder, "_NES")); removeErr != nil {
		t.Fatal(removeErr)
	}
	if linkErr := os.Symlink(outside, filepath.Join(m.paths.MenuFolder, "_NES")); linkErr != nil {
		t.Fatal(linkErr)
	}
	result, err = m.Generate(discover(t, m), nil)
	if err != nil {
		t.Fatal(err)
	}
	if result.Scan.Failed == 0 {
		t.Fatal("escaped output root")
	}
	children, err := os.ReadDir(outside)
	if err != nil || len(children) != 0 {
		t.Fatalf("outside files: %v, %v", children, err)
	}
}

func TestGenerateConcurrentNeverOverwrites(t *testing.T) {
	m := fixture(t)
	put(t, filepath.Join(m.paths.SDRoot, "games", "NES", "game.nes"), "")
	entries := discover(t, m)
	var wg sync.WaitGroup
	results := make(chan GenerateResult, 2)
	errs := make(chan error, 2)
	for range 2 {
		wg.Go(func() { result, err := m.Generate(entries, nil); results <- result; errs <- err })
	}
	wg.Wait()
	close(results)
	close(errs)
	for err := range errs {
		if err != nil {
			t.Fatal(err)
		}
	}
	created, skipped := 0, 0
	for result := range results {
		if result.Scan.Failed != 0 {
			t.Fatal(result.Scan.Errors)
		}
		created += result.Scan.Created
		skipped += result.Scan.Skipped
	}
	if created != 1 || skipped != 1 {
		t.Fatalf("created=%d skipped=%d", created, skipped)
	}
	if _, err := mister.ReadMGL(filepath.Join(m.paths.MenuFolder, "_NES", "game.mgl")); err != nil {
		t.Fatal(err)
	}
}

func TestGenerateProgress(t *testing.T) {
	m := fixture(t)
	for _, system := range []string{"NES", "SNES", "GBA"} {
		put(t, filepath.Join(m.paths.SDRoot, "games", system, "readme.txt"), "")
	}
	var percentages []int
	_, err := m.Generate(discover(t, m), func(p Progress) {
		percentages = append(percentages, (p.Index*100+p.Total-1)/p.Total)
	})
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(percentages, []int{0, 34, 67, 100}) {
		t.Fatalf("progress: %v", percentages)
	}
}
