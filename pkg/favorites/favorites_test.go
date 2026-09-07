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
package favorites

import (
	"archive/zip"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
	"testing"

	"github.com/wizzomafizzo/mrext/pkg/config"
	"github.com/wizzomafizzo/mrext/pkg/games"
)

func newTestManager(t *testing.T) (*Manager, *config.UserConfig, string) {
	t.Helper()
	root := filepath.Join(t.TempDir(), "fat")
	for _, folder := range []string{
		root,
		filepath.Join(root, "games"),
		filepath.Join(root, "_Arcade", "cores"),
		filepath.Join(root, "linux"),
	} {
		if err := os.MkdirAll(folder, 0o755); err != nil {
			t.Fatal(err)
		}
	}
	cfg := DefaultUserConfig()
	cfg.Systems.GamesFolder = []string{root}
	cfg.Favorites.ExternalFolder = filepath.Join(filepath.Dir(root), "usb0")
	paths := RuntimePaths{
		SDRoot:            root,
		StartupScript:     filepath.Join(root, "linux", "user-startup.sh"),
		ArcadeCoresFolder: filepath.Join(root, "_Arcade", "cores"),
	}
	return NewManagerWithPaths(cfg, paths), cfg, root
}

func TestCleanupKeepsUserFolderAfterSettingsChange(t *testing.T) {
	manager, cfg, root := newTestManager(t)
	userFolder := filepath.Join(root, "_Archive")
	if err := os.Mkdir(userFolder, 0o755); err != nil {
		t.Fatal(err)
	}
	created, err := manager.CreateDefaultFolder()
	if err != nil || !created {
		t.Fatalf("create default: %v, %v", created, err)
	}
	cfg.Favorites.DefaultFolder = "_Archive"
	if cleanupErr := manager.CleanupCreatedDefault(created); cleanupErr != nil {
		t.Fatal(cleanupErr)
	}
	if _, statErr := os.Stat(userFolder); statErr != nil {
		t.Fatalf("user folder removed: %v", statErr)
	}
	if _, statErr := os.Stat(filepath.Join(root, "_@Favorites")); !os.IsNotExist(statErr) {
		t.Fatalf("original default not cleaned: %v", statErr)
	}
}

func TestSymlinkedFavoriteRootWorkflows(t *testing.T) {
	manager, _, root := newTestManager(t)
	target := t.TempDir()
	nested := filepath.Join(target, "_Nested")
	if err := os.Mkdir(nested, 0o755); err != nil {
		t.Fatal(err)
	}
	link := filepath.Join(root, "_LinkedFavorites")
	if err := os.Symlink(target, link); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(target, filepath.Join(nested, "_Cycle")); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(nested, "game.mgl"), []byte("<mistergamedescription/>"), 0o644); err != nil {
		t.Fatal(err)
	}
	items, err := manager.List()
	if err != nil || len(items) != 1 || items[0].Path != filepath.Join(link, "_Nested", "game.mgl") {
		t.Fatalf("favorites: %#v, %v", items, err)
	}
	folders, err := manager.DestinationFolders(false, "")
	if err != nil || !slices.Equal(folders, []string{link, filepath.Join(link, "_Nested")}) {
		t.Fatalf("destinations: %v, %v", folders, err)
	}
	editable, err := manager.EditableItems()
	if err != nil || len(editable) != 2 {
		t.Fatalf("editable: %#v, %v", editable, err)
	}
	if setupErr := manager.SetupArcadeLinks(); setupErr != nil {
		t.Fatal(setupErr)
	}
	if _, statErr := os.Lstat(filepath.Join(nested, "cores")); statErr != nil {
		t.Fatal(statErr)
	}
	oldCore := filepath.Join(root, "Core_20260101.rbf")
	newCore := filepath.Join(root, "Core_20260201.rbf")
	if writeErr := os.WriteFile(newCore, nil, 0o644); writeErr != nil {
		t.Fatal(writeErr)
	}
	if linkErr := os.Symlink(oldCore, filepath.Join(nested, "Core_20260101.rbf")); linkErr != nil {
		t.Fatal(linkErr)
	}
	if refreshErr := manager.Refresh(); refreshErr != nil {
		t.Fatal(refreshErr)
	}
	if _, statErr := os.Stat(filepath.Join(nested, "Core_20260201.rbf")); statErr != nil {
		t.Fatal(statErr)
	}
}

func TestArchiveChildParentRemainsBrowsable(t *testing.T) {
	manager, _, root := newTestManager(t)
	archive := filepath.Join(root, "games", "SNES", "games.zip")
	if err := os.MkdirAll(filepath.Dir(archive), 0o755); err != nil {
		t.Fatal(err)
	}
	file, err := os.Create(archive)
	if err != nil {
		t.Fatal(err)
	}
	writer := zip.NewWriter(file)
	if _, createErr := writer.Create("sub/game.sfc"); createErr != nil {
		t.Fatal(createErr)
	}
	if closeErr := writer.Close(); closeErr != nil {
		t.Fatal(closeErr)
	}
	if closeErr := file.Close(); closeErr != nil {
		t.Fatal(closeErr)
	}
	parent, err := manager.ParentFolder(archive + "/sub")
	if err != nil {
		t.Fatal(err)
	}
	entries, err := manager.ListDirectory(parent)
	if err != nil || len(entries) != 1 || entries[0].Name != "sub" {
		t.Fatalf("archive parent entries: %#v, %v", entries, err)
	}
}

func TestSettingsRetainManagedRootComments(t *testing.T) {
	for _, roots := range [][]string{{"/mnt/old"}, {"/mnt/new", "/mnt/second"}, nil} {
		path := filepath.Join(t.TempDir(), "favorites.ini")
		const initial = "[favorites]\n; root notes\ngames_folder = /mnt/old\n" +
			"; second root notes\ngames_folder = /mnt/second\n"
		if err := os.WriteFile(path, []byte(initial), 0o600); err != nil {
			t.Fatal(err)
		}
		settings := SettingsFromConfig(DefaultUserConfig())
		settings.GamesFolders = roots
		if err := SaveSettings(path, &settings); err != nil {
			t.Fatal(err)
		}
		data, err := os.ReadFile(path)
		if err != nil || !strings.Contains(string(data), "; root notes") ||
			!strings.Contains(string(data), "; second root notes") {
			t.Fatalf("lost notes: %s, %v", data, err)
		}
		loaded, err := LoadConfigAt(path, "favorites.sh")
		if err != nil || !slices.Equal(loaded.Favorites.GamesFolder, roots) {
			t.Fatalf("roots failed round trip: %v", err)
		}
	}
}

func TestNeoGeoNearestMappingIsAuthoritative(t *testing.T) {
	for _, local := range []string{
		`<romsets><romset name="other" altname="Other"/></romsets>`,
		`<romsets/>`,
		`<romsets><romset name="game" altname="Partial"/><broken`,
	} {
		manager, _, root := newTestManager(t)
		base := filepath.Join(root, "games", "NeoGeo")
		folder := filepath.Join(base, "sub")
		if err := os.MkdirAll(folder, 0o755); err != nil {
			t.Fatal(err)
		}
		const parentXML = `<romsets><romset name="game" altname="Parent"/></romsets>`
		if err := os.WriteFile(filepath.Join(base, "romsets.xml"), []byte(parentXML), 0o644); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(folder, "romsets.xml"), []byte(local), 0o644); err != nil {
			t.Fatal(err)
		}
		if title := manager.NeoGeoTitle(filepath.Join(folder, "game.zip")); title != "" {
			t.Fatalf("nearest XML not authoritative: %q", title)
		}
	}
}

func TestCustomDefaultFolderRemainsRecognized(t *testing.T) {
	manager, cfg, root := newTestManager(t)
	cfg.Favorites.DefaultFolder = "_Collection"
	if err := ValidateConfig(cfg); err != nil {
		t.Fatal(err)
	}
	created, err := manager.CreateDefaultFolder()
	if err != nil || !created {
		t.Fatalf("create custom default: %v, %v", created, err)
	}
	folders, err := manager.DestinationFolders(false, "")
	if err != nil || !slices.Equal(folders, []string{filepath.Join(root, "_Collection")}) {
		t.Fatalf("custom default destinations: %v, %v", folders, err)
	}
	if manager.IsFavoriteFolderName("_Other") || !manager.IsFavoriteFolderName("_Favorites") {
		t.Fatal("custom default changed matching rules for other folders")
	}
	if _, createErr := manager.CreateFolder(folders[0], "_Nested"); createErr != nil {
		t.Fatal(createErr)
	}
	restarted := NewManagerWithPaths(cfg, manager.paths)
	folders, err = restarted.DestinationFolders(false, "")
	if err != nil || len(folders) != 2 {
		t.Fatalf("restarted destinations: %v, %v", folders, err)
	}
}

func TestPlayerVariantsSelectCorrectCatalogCore(t *testing.T) {
	for _, test := range []struct{ folder, extension, systemID string }{
		{"GBA2P", ".gba", "GBA2P"},
		{"GBA", ".gba", "GBA"},
		{"GAMEBOY2P", ".gb", "Gameboy2P"},
		{"GAMEBOY", ".gb", "Gameboy"},
	} {
		t.Run(test.folder, func(t *testing.T) {
			manager, _, root := newTestManager(t)
			folder := filepath.Join(root, "games", test.folder)
			if err := os.MkdirAll(folder, 0o755); err != nil {
				t.Fatal(err)
			}
			media := filepath.Join(folder, "game"+test.extension)
			if err := os.WriteFile(media, nil, 0o644); err != nil {
				t.Fatal(err)
			}
			expected, err := games.GetSystem(test.systemID)
			if err != nil {
				t.Fatal(err)
			}
			want, err := manager.generateMGL(expected, media)
			if err != nil {
				t.Fatal(err)
			}
			for index := range 20 {
				entries, listErr := manager.ListDirectory(folder)
				if listErr != nil || len(entries) != 1 || entries[0].System == nil ||
					entries[0].System.Id != test.systemID {
					t.Fatalf("wrong player variant: %#v, %v", entries, listErr)
				}
				path, createErr := manager.CreateGameFavorite(entries[0].System, media, root, strconv.Itoa(index))
				if createErr != nil {
					t.Fatal(createErr)
				}
				data, readErr := os.ReadFile(path)
				if readErr != nil || string(data) != want {
					t.Fatalf("wrong variant MGL: %s, %v", data, readErr)
				}
			}
		})
	}
}

func TestConfigCreationPreservesExistingFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "favorites.ini")
	t.Setenv(config.UserConfigEnv, path)
	cfg, err := LoadConfig()
	if err != nil {
		t.Fatal(err)
	}
	if ensureErr := EnsureConfigFile(cfg); ensureErr != nil {
		t.Fatal(ensureErr)
	}
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(data), "[favorites]") || !strings.Contains(string(data), "[tui]") {
		t.Fatalf("default config missing sections:\n%s", data)
	}

	const custom = "[cores]\nall = yc\n"
	if writeErr := os.WriteFile(path, []byte(custom), 0o600); writeErr != nil {
		t.Fatal(writeErr)
	}
	if ensureErr := EnsureConfigFile(cfg); ensureErr != nil {
		t.Fatal(ensureErr)
	}
	data, err = os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != custom {
		t.Fatalf("existing config changed: %q", data)
	}
	loaded, err := LoadConfig()
	if err != nil {
		t.Fatal(err)
	}
	if loaded.FavoritesCores.All != "yc" || loaded.Favorites.DefaultFolder != "_@Favorites" {
		t.Fatalf("legacy config did not merge with defaults: %#v", loaded)
	}
}

func TestSettingsSavePreservesUnknownINIContent(t *testing.T) {
	path := filepath.Join(t.TempDir(), "favorites.ini")
	initial := "[cores]\nall =\n\n[custom]\n; keep this comment\nvalue = untouched\n"
	if err := os.WriteFile(path, []byte(initial), 0o600); err != nil {
		t.Fatal(err)
	}
	cfg := DefaultUserConfig()
	cfg.IniPath = path
	settings := SettingsFromConfig(cfg)
	settings.DefaultFolder = "_My Favorites"
	settings.FolderNameContains = []string{"fav", "collection"}
	settings.GamesFolders = []string{"/media/network", "/media/usb9"}
	settings.AlternateCore = "yc"
	settings.Theme = "high_contrast"
	settings.HideRootFiles = false
	if err := SaveSettings(path, &settings); err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	text := string(data)
	for _, expected := range []string{"[custom]", "; keep this comment", "value = untouched"} {
		if !strings.Contains(text, expected) {
			t.Fatalf("saved configuration lost %q:\n%s", expected, text)
		}
	}
	t.Setenv(config.UserConfigEnv, path)
	loaded, err := LoadConfig()
	if err != nil {
		t.Fatal(err)
	}
	if loaded.Favorites.DefaultFolder != "_My Favorites" ||
		!slices.Equal(loaded.Favorites.FolderNameContains, []string{"fav", "collection"}) ||
		!slices.Equal(loaded.Favorites.GamesFolder, []string{"/media/network", "/media/usb9"}) ||
		loaded.FavoritesCores.All != "yc" || loaded.TUI.Theme != "high_contrast" || loaded.Favorites.HideRootFiles {
		t.Fatalf("loaded settings = %#v", SettingsFromConfig(loaded))
	}
}

func TestSettingsRejectInvalidValues(t *testing.T) {
	settings := SettingsFromConfig(DefaultUserConfig())
	settings.DefaultFolder = "Favorites"
	if err := settings.Validate(); err == nil {
		t.Fatal("invalid default folder accepted")
	}
}

func TestCoreFavoriteCreateRenameAndMove(t *testing.T) {
	manager, _, root := newTestManager(t)
	folder := filepath.Join(root, "_@Favorites")
	sourceFolder := filepath.Join(root, "_Console")
	for _, path := range []string{folder, sourceFolder} {
		if err := os.MkdirAll(path, 0o750); err != nil {
			t.Fatal(err)
		}
	}
	source := filepath.Join(sourceFolder, "NES_20260101.rbf")
	if err := os.WriteFile(source, nil, 0o600); err != nil {
		t.Fatal(err)
	}
	favorite, err := manager.CreateCoreFavorite(source, folder, "Nintendo")
	if err != nil {
		t.Fatal(err)
	}
	if target, readErr := os.Readlink(favorite); readErr != nil || target != source {
		t.Fatalf("core favorite target = %q, err = %v", target, readErr)
	}
	renamed, err := manager.Rename(favorite, "Nintendo Entertainment System.rbf")
	if err != nil {
		t.Fatal(err)
	}
	if filepath.Ext(renamed) != ".rbf" || strings.HasSuffix(renamed, ".rbf.rbf") {
		t.Fatalf("renamed favorite = %q", renamed)
	}
	moved, err := manager.Move(renamed, root)
	if err != nil {
		t.Fatal(err)
	}
	if filepath.Dir(moved) != root {
		t.Fatalf("moved favorite = %q", moved)
	}
}

func TestFavoriteDiscoveryAndSafeFolderDeletion(t *testing.T) {
	manager, _, root := newTestManager(t)
	folder := filepath.Join(root, "_@Favorites")
	nested := filepath.Join(folder, "_RPG")
	if err := os.MkdirAll(nested, 0o755); err != nil {
		t.Fatal(err)
	}
	core := filepath.Join(root, "_Console", "NES_20260101.rbf")
	if err := os.MkdirAll(filepath.Dir(core), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(core, nil, 0o644); err != nil {
		t.Fatal(err)
	}
	favorite := filepath.Join(nested, filepath.Base(core))
	if err := os.Symlink(core, favorite); err != nil {
		t.Fatal(err)
	}

	items, err := manager.EditableItems()
	if err != nil {
		t.Fatal(err)
	}
	if len(items) != 2 || items[0].Kind != FavoriteFolder && items[1].Kind != FavoriteFolder {
		t.Fatalf("unexpected editable items: %#v", items)
	}
	if err := manager.Delete(nested); err == nil || err.Error() != "folder is not empty" {
		t.Fatalf("delete non-empty folder error = %v", err)
	}
	if err := manager.Delete(favorite); err != nil {
		t.Fatal(err)
	}
	if err := manager.Delete(nested); err != nil {
		t.Fatal(err)
	}
}

func TestRefreshUpdatesVersionedCoreSymlink(t *testing.T) {
	manager, _, root := newTestManager(t)
	folder := filepath.Join(root, "_@Favorites")
	coreFolder := filepath.Join(root, "_Console")
	if err := os.MkdirAll(folder, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(coreFolder, 0o755); err != nil {
		t.Fatal(err)
	}
	oldCore := filepath.Join(coreFolder, "SNES_20250101.rbf")
	staleCore := filepath.Join(coreFolder, "SNES_20250601.rbf")
	newCore := filepath.Join(coreFolder, "SNES_20260101.rbf")
	for _, core := range []string{staleCore, newCore} {
		if err := os.WriteFile(core, nil, 0o644); err != nil {
			t.Fatal(err)
		}
	}
	oldLink := filepath.Join(folder, filepath.Base(oldCore))
	if err := os.Symlink(oldCore, oldLink); err != nil {
		t.Fatal(err)
	}
	if err := manager.Refresh(); err != nil {
		t.Fatal(err)
	}
	newLink := filepath.Join(folder, filepath.Base(newCore))
	target, err := os.Readlink(newLink)
	if err != nil {
		t.Fatal(err)
	}
	if target != newCore {
		t.Fatalf("new core target = %q, want %q", target, newCore)
	}
	if _, err := os.Lstat(oldLink); !os.IsNotExist(err) {
		t.Fatalf("old link still exists: %v", err)
	}
	if _, err := os.Lstat(filepath.Join(folder, filepath.Base(staleCore))); !os.IsNotExist(err) {
		t.Fatalf("stale core was linked instead of newest: %v", err)
	}
}

func TestRefreshPreservesCustomCoreFavoriteName(t *testing.T) {
	for _, relative := range []bool{false, true} {
		t.Run(fmt.Sprintf("relative=%t", relative), func(t *testing.T) {
			manager, _, root := newTestManager(t)
			folder := filepath.Join(root, "_@Favorites")
			coreFolder := filepath.Join(root, "_Console")
			for _, path := range []string{folder, coreFolder} {
				if err := os.MkdirAll(path, 0o750); err != nil {
					t.Fatal(err)
				}
			}
			oldCore := filepath.Join(coreFolder, "NES_20250101.rbf")
			newCore := filepath.Join(coreFolder, "NES_20260101.rbf")
			if err := os.WriteFile(newCore, nil, 0o600); err != nil {
				t.Fatal(err)
			}
			if relative {
				oldCore = filepath.Join("..", "_Console", filepath.Base(oldCore))
			}
			link := filepath.Join(folder, "My_Nintendo.rbf")
			if err := os.Symlink(oldCore, link); err != nil {
				t.Fatal(err)
			}
			if err := manager.Refresh(); err != nil {
				t.Fatal(err)
			}
			target, err := os.Readlink(link)
			if err != nil || target != newCore {
				t.Fatalf("custom favorite lost: target=%q, err=%v", target, err)
			}
		})
	}
}

func TestRefreshUpdatesRelativeVersionedCoreSymlink(t *testing.T) {
	manager, _, root := newTestManager(t)
	folder := filepath.Join(root, "_@Favorites")
	coreFolder := filepath.Join(root, "_Console")
	for _, path := range []string{folder, coreFolder} {
		if err := os.MkdirAll(path, 0o750); err != nil {
			t.Fatal(err)
		}
	}
	newCore := filepath.Join(coreFolder, "NES_20260101.rbf")
	if err := os.WriteFile(newCore, nil, 0o600); err != nil {
		t.Fatal(err)
	}
	oldLink := filepath.Join(folder, "NES_20250101.rbf")
	if err := os.Symlink(filepath.Join("..", "_Console", "NES_20250101.rbf"), oldLink); err != nil {
		t.Fatal(err)
	}
	if err := manager.Refresh(); err != nil {
		t.Fatal(err)
	}
	newLink := filepath.Join(folder, filepath.Base(newCore))
	target, err := os.Readlink(newLink)
	if err != nil {
		t.Fatal(err)
	}
	if target != newCore {
		t.Fatalf("new relative-core target = %q, want %q", target, newCore)
	}
}

func TestCatalogBackedMGLGeneration(t *testing.T) {
	manager, _, root := newTestManager(t)
	folder := filepath.Join(root, "_@Favorites")
	gameFolder := filepath.Join(root, "games", "Jaguar")
	if err := os.MkdirAll(folder, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(gameFolder, 0o755); err != nil {
		t.Fatal(err)
	}
	gamePath := filepath.Join(gameFolder, `Alien & Predator.j64`)
	if err := os.WriteFile(gamePath, nil, 0o644); err != nil {
		t.Fatal(err)
	}
	system, err := games.GetSystem("Jaguar")
	if err != nil {
		t.Fatal(err)
	}
	launcherPath, err := manager.CreateGameFavorite(system, gamePath, folder, "Alien vs Predator")
	if err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(launcherPath)
	if err != nil {
		t.Fatal(err)
	}
	launcher := string(data)
	if !strings.Contains(launcher, "<rbf>_Console/Jaguar</rbf>") {
		t.Fatalf("missing Jaguar RBF:\n%s", launcher)
	}
	if !strings.Contains(launcher, "Alien &amp; Predator.j64") {
		t.Fatalf("media path was not XML-escaped:\n%s", launcher)
	}
	if !strings.Contains(launcher, `<reset delay="1" hold="1"/>`) {
		t.Fatalf("missing canonical reset timing:\n%s", launcher)
	}
}

func TestGeneratedMGLIsDiscoveredWithOriginalTarget(t *testing.T) {
	manager, _, root := newTestManager(t)
	folder := filepath.Join(root, "_@Favorites")
	gameFolder := filepath.Join(root, "games", "SNES")
	for _, path := range []string{folder, gameFolder} {
		if err := os.MkdirAll(path, 0o750); err != nil {
			t.Fatal(err)
		}
	}
	gamePath := filepath.Join(gameFolder, "Chrono Trigger.sfc")
	if err := os.WriteFile(gamePath, nil, 0o600); err != nil {
		t.Fatal(err)
	}
	system, err := games.GetSystem("SNES")
	if err != nil {
		t.Fatal(err)
	}
	launcherPath, err := manager.CreateGameFavorite(system, gamePath, folder, "Chrono Trigger")
	if err != nil {
		t.Fatal(err)
	}
	items, err := manager.List()
	if err != nil {
		t.Fatal(err)
	}
	if len(items) != 1 || items[0].Path != launcherPath || items[0].Target != gamePath {
		t.Fatalf("discovered Favorites = %#v", items)
	}
}

func TestNeoGeoTitleAndRelativeLauncher(t *testing.T) {
	manager, _, root := newTestManager(t)
	folder := filepath.Join(root, "_@Favorites")
	gameFolder := filepath.Join(root, "games", "NeoGeo")
	if err := os.MkdirAll(folder, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(gameFolder, 0o755); err != nil {
		t.Fatal(err)
	}
	romsets := `<romsets><romset name="mslug,mslugx" altname="Metal Slug &amp; Friends"/></romsets>`
	if err := os.WriteFile(filepath.Join(gameFolder, "romsets.xml"), []byte(romsets), 0o644); err != nil {
		t.Fatal(err)
	}
	gamePath := filepath.Join(gameFolder, "mslug.zip")
	if err := createZip(gamePath, map[string]string{"game.bin": "data"}); err != nil {
		t.Fatal(err)
	}
	system, err := games.GetSystem("NeoGeo")
	if err != nil {
		t.Fatal(err)
	}
	if got := manager.DefaultName(system, gamePath); got != "Metal Slug & Friends" {
		t.Fatalf("NeoGeo title = %q", got)
	}
	launcherPath, err := manager.CreateGameFavorite(system, gamePath, folder, "Metal Slug")
	if err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(launcherPath)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(data), `path="mslug.zip"`) {
		t.Fatalf("NeoGeo path is not relative:\n%s", data)
	}
}

func TestBrowseInsideZIPUsesCatalog(t *testing.T) {
	manager, _, root := newTestManager(t)
	gameFolder := filepath.Join(root, "games", "SNES")
	if err := os.MkdirAll(gameFolder, 0o755); err != nil {
		t.Fatal(err)
	}
	archive := filepath.Join(gameFolder, "collection.zip")
	if err := createZip(archive, map[string]string{
		"RPG/Chrono Trigger.sfc": "data",
		"notes.txt":              "ignore",
	}); err != nil {
		t.Fatal(err)
	}
	entries, err := manager.ListDirectory(archive + string(filepath.Separator))
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 1 || !entries[0].IsDirectory || entries[0].Name != "RPG" {
		t.Fatalf("archive root entries = %#v", entries)
	}
	entries, err = manager.ListDirectory(filepath.Join(archive, "RPG"))
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 1 || entries[0].System == nil || entries[0].System.Id != "SNES" {
		t.Fatalf("archive game entries = %#v", entries)
	}
}

func TestBrowseFollowsDirectorySymlinksAndMarksNeoGeoZIP(t *testing.T) {
	manager, _, root := newTestManager(t)
	externalGames := filepath.Join(root, "external-games")
	if err := os.MkdirAll(filepath.Join(externalGames, "SNES"), 0o750); err != nil {
		t.Fatal(err)
	}
	gamesLink := filepath.Join(root, "games")
	if err := os.Remove(gamesLink); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(externalGames, gamesLink); err != nil {
		t.Fatal(err)
	}
	entries, err := manager.ListDirectory(root)
	if err != nil {
		t.Fatal(err)
	}
	foundGamesLink := false
	for _, entry := range entries {
		if entry.Path == gamesLink && entry.IsDirectory {
			foundGamesLink = true
		}
	}
	if !foundGamesLink {
		t.Fatalf("symlinked games folder entries = %#v", entries)
	}

	neoGeoTarget := filepath.Join(externalGames, "NeoGeo")
	if mkdirErr := os.MkdirAll(neoGeoTarget, 0o750); mkdirErr != nil {
		t.Fatal(mkdirErr)
	}
	neoGeoFolder := filepath.Join(gamesLink, "NeoGeo")
	archive := filepath.Join(neoGeoFolder, "mslug.zip")
	if createErr := createZip(archive, map[string]string{"nested/game.neo": "data"}); createErr != nil {
		t.Fatal(createErr)
	}
	entries, err = manager.ListDirectory(neoGeoFolder)
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 1 || entries[0].System == nil || !entries[0].IsArchive || entries[0].IsDirectory {
		t.Fatalf("NeoGeo ZIP entry = %#v", entries)
	}
	inside, err := manager.ListDirectory(archive + string(filepath.Separator))
	if err != nil {
		t.Fatal(err)
	}
	if len(inside) != 1 || !inside[0].IsDirectory {
		t.Fatalf("NeoGeo ZIP contents = %#v", inside)
	}
}

func TestCanonicalCatalogRecognizesEveryLegacyGamesFolder(t *testing.T) {
	_, cfg, root := newTestManager(t)
	legacyFolders := []string{
		"Amiga", "Arcadia", "AVision", "Astrocade", "ATARI2600", "ATARI5200", "ATARI7800", "AtariLynx",
		"C64", "ChannelF", "Coleco", "CreatiVision", "GAMEBOY2P", "GAMEBOY", "GBC", "Gamate",
		"GameNWatch", "GameGear", "GBA2P", "GBA", "MegaDrive", "Genesis", "Intellivision", "MegaCD",
		"N64", "NeoGeo-CD", "NeoGeo", "NES", "ODYSSEY2", "PSX", "PocketChallengeV2", "PokemonMini",
		"Saturn", "S32X", "SG1000", "SGB", "SMS", "SNES", "SuperVision", "TGFX16-CD", "TGFX16",
		"VC4000", "VECTREX", "WonderSwan", "WonderSwanColor",
	}
	for _, folder := range legacyFolders {
		t.Run(folder, func(t *testing.T) {
			path := filepath.Join(root, "games", folder) + string(filepath.Separator)
			if systems := games.FolderToSystems(cfg, path); len(systems) == 0 {
				t.Fatalf("legacy games folder is not in canonical catalog: %s", folder)
			}
		})
	}
}

func TestCanonicalCatalogCoversLegacyAndNewSystems(t *testing.T) {
	manager, _, root := newTestManager(t)
	folder := filepath.Join(root, "_@Favorites")
	if err := os.MkdirAll(folder, 0o750); err != nil {
		t.Fatal(err)
	}
	tests := []struct {
		id       string
		filename string
		wantRBF  string
		wantFile string
	}{
		{id: "Genesis", filename: "Sonic.md", wantRBF: "_Console/MegaDrive", wantFile: `type="f" index="1"`},
		{id: "CDI", filename: "Hotel Mario.chd", wantRBF: "_Console/CDi", wantFile: `type="s" index="1"`},
		{id: "CoCo2", filename: "Mega-Bug.ccc", wantRBF: "_Computer/CoCo2", wantFile: `type="f" index="1"`},
	}
	for _, test := range tests {
		t.Run(test.id, func(t *testing.T) {
			system, err := games.GetSystem(test.id)
			if err != nil {
				t.Fatal(err)
			}
			mediaPath := filepath.Join(root, "games", system.Folder[0], test.filename)
			if mkdirErr := os.MkdirAll(filepath.Dir(mediaPath), 0o750); mkdirErr != nil {
				t.Fatal(mkdirErr)
			}
			if writeErr := os.WriteFile(mediaPath, nil, 0o600); writeErr != nil {
				t.Fatal(writeErr)
			}
			launcherPath, err := manager.CreateGameFavorite(system, mediaPath, folder, test.id)
			if err != nil {
				t.Fatal(err)
			}
			data, err := os.ReadFile(launcherPath)
			if err != nil {
				t.Fatal(err)
			}
			launcher := string(data)
			if !strings.Contains(launcher, "<rbf>"+test.wantRBF+"</rbf>") ||
				!strings.Contains(launcher, test.wantFile) {
				t.Fatalf("unexpected %s launcher:\n%s", test.id, launcher)
			}
		})
	}
}

func TestRACoreMappingsAndFallback(t *testing.T) {
	manager, cfg, root := newTestManager(t)
	cfg.FavoritesCores.All = "ra"
	cfg.Favorites.CorePrefix = "_Custom/"
	if err := ValidateConfig(cfg); err != nil {
		t.Fatal(err)
	}
	for id, variant := range raCores {
		t.Run(id, func(t *testing.T) {
			source, err := games.GetSystem(id)
			if err != nil {
				t.Fatal(err)
			}
			corePath := filepath.Join(root, config.RACoresFolder, variant.name+"_20260101.rbf")
			if err = os.MkdirAll(filepath.Dir(corePath), 0o750); err != nil {
				t.Fatal(err)
			}
			if err = os.WriteFile(corePath, nil, 0o600); err != nil {
				t.Fatal(err)
			}
			resolved := *source
			if err = manager.configureRACore(&resolved); err != nil {
				t.Fatal(err)
			}
			if resolved.Rbf != filepath.Join(config.RACoresFolder, variant.name) ||
				resolved.SetName != variant.setName || resolved.SetNameSameDir != variant.sameDir {
				t.Fatalf("RA metadata: %+v", resolved)
			}
			if strings.HasPrefix(source.SetName, "RA_") {
				t.Fatal("shared catalog mutated")
			}
			if err = os.Remove(corePath); err != nil {
				t.Fatal(err)
			}
			fallback := *source
			fallback.Rbf = manager.resolveCore(source)
			if err = manager.configureRACore(&fallback); err != nil {
				t.Fatal(err)
			}
			if fallback.Rbf != "_Custom/"+source.Rbf || fallback.SetName != source.SetName {
				t.Fatalf("fallback changed metadata: %+v", fallback)
			}
		})
	}
}

func TestRAIgnoresNonmatchingCoreFiles(t *testing.T) {
	manager, cfg, root := newTestManager(t)
	cfg.FavoritesCores.All = "ra"
	folder := filepath.Join(root, config.RACoresFolder)
	if err := os.MkdirAll(folder, 0o750); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(folder, "NES2.rbf"), nil, 0o600); err != nil {
		t.Fatal(err)
	}
	source, err := games.GetSystem("NES")
	if err != nil {
		t.Fatal(err)
	}
	resolved := *source
	if err = manager.configureRACore(&resolved); err != nil {
		t.Fatal(err)
	}
	if resolved.Rbf != source.Rbf || resolved.SetName != source.SetName {
		t.Fatal("selected unrelated RA core")
	}
}

func TestRAGameFavoritesUseSetNamesAndCatalogSlots(t *testing.T) {
	for _, tc := range []struct{ id, file, core, setname, slot string }{
		{"NES", "game.nes", "NES", `<setname same_dir="1">RA_NES</setname>`, `type="f" index="1"`},
		{"FDS", "game.fds", "NES", `<setname>RA_FDS</setname>`, `type="f" index="1"`},
		{"GameboyColor", "game.gbc", "Gameboy", `<setname>RA_GBC</setname>`, `type="f" index="1"`},
		{"Atari2600", "game.a26", "Atari7800", `<setname same_dir="1">RA_Atari7800</setname>`, `type="f" index="1"`},
		{"Genesis", "game.md", "MegaDrive", `<setname same_dir="1">RA_MegaDrive</setname>`, `type="f" index="1"`},
	} {
		t.Run(tc.id, func(t *testing.T) {
			manager, cfg, root := newTestManager(t)
			cfg.FavoritesCores.All = "ra"
			folder := filepath.Join(root, "_@Favorites")
			cores := filepath.Join(root, config.RACoresFolder)
			for _, dir := range []string{folder, cores} {
				if err := os.MkdirAll(dir, 0o750); err != nil {
					t.Fatal(err)
				}
			}
			if err := os.WriteFile(filepath.Join(cores, tc.core+".rbf"), nil, 0o600); err != nil {
				t.Fatal(err)
			}
			system, err := games.GetSystem(tc.id)
			if err != nil {
				t.Fatal(err)
			}
			media := filepath.Join(root, "games", system.Folder[0], tc.file)
			path, err := manager.CreateGameFavorite(system, media, folder, "RA game")
			if err != nil {
				t.Fatal(err)
			}
			data, err := os.ReadFile(path)
			if err != nil {
				t.Fatal(err)
			}
			for _, part := range []string{"<rbf>_RA_Cores/Cores/" + tc.core + "</rbf>", tc.setname, tc.slot, media} {
				if !strings.Contains(string(data), part) {
					t.Fatalf("missing %s in %s", part, data)
				}
			}
		})
	}
}

func TestAlternateCoreAndCorePrefix(t *testing.T) {
	manager, cfg, root := newTestManager(t)
	folder := filepath.Join(root, "_@Favorites")
	gameFolder := filepath.Join(root, "games", "Genesis")
	llapiFolder := filepath.Join(root, "_LLAPI")
	for _, path := range []string{folder, gameFolder, llapiFolder} {
		if err := os.MkdirAll(path, 0o750); err != nil {
			t.Fatal(err)
		}
	}
	mediaPath := filepath.Join(gameFolder, "Sonic.md")
	if err := os.WriteFile(mediaPath, nil, 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(llapiFolder, "Genesis_LLAPI_20260101.rbf"), nil, 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(llapiFolder, "GenesisYC_20260101.rbf"), nil, 0o600); err != nil {
		t.Fatal(err)
	}
	system, err := games.GetSystem("Genesis")
	if err != nil {
		t.Fatal(err)
	}
	cfg.FavoritesCores.All = "llapi"
	launcherPath, err := manager.CreateGameFavorite(system, mediaPath, folder, "LLAPI")
	if err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(launcherPath)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(data), "<rbf>_LLAPI/Genesis_LLAPI</rbf>") {
		t.Fatalf("alternate core not selected:\n%s", data)
	}

	cfg.FavoritesCores.All = "yc"
	launcherPath, err = manager.CreateGameFavorite(system, mediaPath, folder, "YC")
	if err != nil {
		t.Fatal(err)
	}
	data, err = os.ReadFile(launcherPath)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(data), "<rbf>_LLAPI/GenesisYC</rbf>") {
		t.Fatalf("YC core not selected:\n%s", data)
	}

	cfg.FavoritesCores.All = ""
	cfg.Favorites.CorePrefix = "_Alt/"
	launcherPath, err = manager.CreateGameFavorite(system, mediaPath, folder, "Prefixed")
	if err != nil {
		t.Fatal(err)
	}
	data, err = os.ReadFile(launcherPath)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(data), "<rbf>_Alt/_Console/MegaDrive</rbf>") {
		t.Fatalf("core prefix not applied:\n%s", data)
	}
}

func TestArcadeLinksExposeNewCoresWithoutRefresh(t *testing.T) {
	manager, _, root := newTestManager(t)
	folder := filepath.Join(root, "_@Favorites")
	nested := filepath.Join(folder, "_Arcade Picks")
	if err := os.MkdirAll(nested, 0o750); err != nil {
		t.Fatal(err)
	}
	if err := manager.SetupArcadeLinks(); err != nil {
		t.Fatal(err)
	}
	created, err := manager.CreateFolder(folder, "_New Picks")
	if err != nil {
		t.Fatal(err)
	}

	// Downloader adds cores after Favorites has already prepared its menu folders.
	const coreName = "RTypeII_20220918.rbf"
	const coreData = "new arcade core"
	corePath := filepath.Join(manager.paths.ArcadeCoresFolder, coreName)
	if err := os.WriteFile(corePath, []byte(coreData), 0o600); err != nil {
		t.Fatal(err)
	}
	for _, destination := range []string{root, folder, nested, created} {
		link := filepath.Join(destination, "cores")
		target, readErr := os.Readlink(link)
		if readErr != nil {
			t.Fatal(readErr)
		}
		if target != manager.paths.ArcadeCoresFolder {
			t.Fatalf("%s target = %q, want %q", link, target, manager.paths.ArcadeCoresFolder)
		}
		data, readErr := os.ReadFile(filepath.Join(link, coreName))
		if readErr != nil {
			t.Fatal(readErr)
		}
		if string(data) != coreData {
			t.Fatalf("new core unavailable through %s: %q", link, data)
		}
	}
}

func TestArcadeLinksDoNotReplaceExistingContent(t *testing.T) {
	manager, _, root := newTestManager(t)
	folder := filepath.Join(root, "_@Favorites")
	nested := filepath.Join(folder, "_Arcade Picks")
	if err := os.MkdirAll(nested, 0o750); err != nil {
		t.Fatal(err)
	}
	rootCores := filepath.Join(root, "cores")
	if err := os.WriteFile(rootCores, []byte("user content"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := manager.SetupArcadeLinks(); err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(rootCores)
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != "user content" {
		t.Fatalf("root cores content changed: %q", data)
	}
	for _, path := range []string{filepath.Join(folder, "cores"), filepath.Join(nested, "cores")} {
		info, err := os.Lstat(path)
		if err != nil {
			t.Fatal(err)
		}
		if info.Mode()&os.ModeSymlink == 0 {
			t.Fatalf("arcade cores path is not a symlink: %s", path)
		}
	}
}

func TestStartupEntryUsesConfiguredApplicationPath(t *testing.T) {
	manager, cfg, root := newTestManager(t)
	cfg.AppPath = filepath.Join(root, "Custom Scripts", "favorites.sh")
	startup := filepath.Join(root, "linux", "user-startup.sh")
	if err := os.MkdirAll(filepath.Dir(startup), 0o750); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(startup, []byte("#!/bin/bash\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := manager.TryAddToStartup(); err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(startup)
	if err != nil {
		t.Fatal(err)
	}
	quotedPath := `"` + cfg.AppPath + `"`
	if !strings.Contains(string(data), "[[ -e "+quotedPath+" ]] && "+quotedPath+" refresh") {
		t.Fatalf("startup entry = %q", data)
	}
}

func TestStartupEntryIsIdempotent(t *testing.T) {
	manager, _, _ := newTestManager(t)
	if err := os.WriteFile(manager.paths.StartupScript, []byte("#!/bin/bash\n"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := manager.TryAddToStartup(); err != nil {
		t.Fatal(err)
	}
	if err := manager.TryAddToStartup(); err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(manager.paths.StartupScript)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Count(string(data), startupMarker) != 1 {
		t.Fatalf("startup marker count = %d", strings.Count(string(data), startupMarker))
	}
}

func createZip(path string, files map[string]string) error {
	file, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("create ZIP fixture: %w", err)
	}
	writer := zip.NewWriter(file)
	for name, content := range files {
		entry, createErr := writer.Create(name)
		if createErr != nil {
			_ = writer.Close()
			_ = file.Close()
			return fmt.Errorf("create ZIP entry: %w", createErr)
		}
		if _, writeErr := entry.Write([]byte(content)); writeErr != nil {
			_ = writer.Close()
			_ = file.Close()
			return fmt.Errorf("write ZIP entry: %w", writeErr)
		}
	}
	if err := writer.Close(); err != nil {
		_ = file.Close()
		return fmt.Errorf("close ZIP writer: %w", err)
	}
	if err := file.Close(); err != nil {
		return fmt.Errorf("close ZIP fixture: %w", err)
	}
	return nil
}

func TestRefreshKeepsFavoritesWhenStorageIsDetached(t *testing.T) {
	manager, _, root := newTestManager(t)
	folder := filepath.Join(root, "_@Favorites")
	if err := os.MkdirAll(folder, 0o755); err != nil {
		t.Fatal(err)
	}

	// Refresh also runs from user-startup.sh, before USB and network mounts
	// settle, so a favorite pointing at an unplugged drive must survive.
	media := t.TempDir()
	attached := filepath.Join(media, "usb0")
	detached := filepath.Join(media, "usb1")
	if err := os.MkdirAll(filepath.Join(attached, "games"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.Mkdir(detached, 0o755); err != nil {
		t.Fatal(err)
	}

	previous := config.StorageRoots
	config.StorageRoots = []string{attached, detached}
	t.Cleanup(func() { config.StorageRoots = previous })

	onDetached := filepath.Join(folder, "Unplugged.mgl")
	if err := os.Symlink(filepath.Join(detached, "games", "game.mgl"), onDetached); err != nil {
		t.Fatal(err)
	}
	onAttached := filepath.Join(folder, "Deleted.mgl")
	if err := os.Symlink(filepath.Join(attached, "games", "gone.mgl"), onAttached); err != nil {
		t.Fatal(err)
	}

	if err := manager.Refresh(); err != nil {
		t.Fatal(err)
	}

	if _, err := os.Lstat(onDetached); err != nil {
		t.Errorf("favorite on an unplugged drive was removed: %v", err)
	}
	if _, err := os.Lstat(onAttached); !os.IsNotExist(err) {
		t.Errorf("favorite for a genuinely deleted game survived: %v", err)
	}
}
