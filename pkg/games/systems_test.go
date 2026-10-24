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
	"encoding/json"
	"path/filepath"
	"reflect"
	"slices"
	"strings"
	"testing"

	"github.com/ZaparooProject/zaparoo-core/mister/catalog"
	"github.com/wizzomafizzo/mrext/pkg/config"
)

func TestFolderMatchingUsesComponentBoundaries(t *testing.T) {
	root := filepath.Join(t.TempDir(), "library")
	cfg := &config.UserConfig{Systems: config.SystemsConfig{GamesFolder: []string{root}}}
	if found := FolderToSystems(cfg, filepath.Join(root+"backup", "GBA", "game.gba")); len(found) != 0 {
		t.Fatalf("matched sibling root: %#v", found)
	}
	if found := FolderToSystems(cfg, filepath.Join(root, "GBAunknown", "game.gba")); len(found) != 0 {
		t.Fatalf("matched partial system folder: %#v", found)
	}
	cfg.Systems.GamesFolder = append(cfg.Systems.GamesFolder, filepath.Join(root, "nested"))
	found := FolderToSystems(cfg, filepath.Join(root, "nested", "GBA", "game.gba"))
	if len(found) != 1 || found[0].Id != "GBA" {
		t.Fatalf("nested root shadowed by parent: %#v", found)
	}
}

func TestSystemsUseCatalogOperationalData(t *testing.T) {
	t.Parallel()

	definitions := catalog.All()
	if len(Systems) != len(definitions) {
		t.Fatalf("system count: want %d, got %d", len(definitions), len(Systems))
	}

	for i := range definitions {
		definition := &definitions[i]
		system, ok := Systems[definition.ID]
		if !ok {
			t.Fatalf("catalog system missing from mrext: %s", definition.ID)
		}
		if !reflect.DeepEqual(system.Folder, definition.Folders) {
			t.Fatalf("%s folders: want %#v, got %#v", definition.ID, definition.Folders, system.Folder)
		}
		if system.Rbf != definition.RBF || system.SetName != definition.SetName ||
			system.SetNameSameDir != definition.SetNameSameDir {
			t.Fatalf("%s launch metadata differs from catalog", definition.ID)
		}
		if !reflect.DeepEqual(CatalogCore(&system).Slots, definition.Slots) {
			t.Fatalf("%s slots differ from catalog", definition.ID)
		}
		if !reflect.DeepEqual(system.extensions, definition.Extensions) {
			t.Fatalf("%s scan extensions differ from catalog", definition.ID)
		}
	}
}

func TestSystemJSONShapeRemainsLegacyCompatible(t *testing.T) {
	t.Parallel()

	data, err := json.Marshal(Systems["Jaguar"])
	if err != nil {
		t.Fatal(err)
	}
	encoded := string(data)
	if !strings.Contains(encoded, `"Mgl":{"Delay":1,"Method":"f","Index":0}`) {
		t.Fatalf("legacy MGL field names changed: %s", encoded)
	}
	if strings.Contains(encoded, `"resetDelay"`) || strings.Contains(encoded, `"extensions"`) {
		t.Fatalf("internal catalog fields leaked into legacy JSON: %s", encoded)
	}
}

func TestSystemsUseGeneratedCoreMetadataAndAliases(t *testing.T) {
	t.Parallel()

	genesis := Systems["Genesis"]
	if genesis.Name != "Genesis" || genesis.Manufacturer != ManufacturerSega {
		t.Fatalf("generated Core metadata missing: %#v", genesis)
	}

	resolved, err := LookupSystem("MegaDrive")
	if err != nil {
		t.Fatal(err)
	}
	if resolved.Id != "Genesis" {
		t.Fatalf("alias resolved to %s", resolved.Id)
	}
}

func TestSystemsIncludeNewCatalogEntries(t *testing.T) {
	t.Parallel()

	for _, id := range []string{
		"AppleIIGS", "AppleLisa", "GameGear2P", "JaguarCD", "MegaVGMDrive",
		"NeoGeoPocket", "NeoGeoPocketColor", "OpenBOR", "Pico8", "VirtualBoy",
	} {
		system, ok := Systems[id]
		if !ok {
			t.Fatalf("new catalog entry missing: %s", id)
		}
		if system.Name == "" {
			t.Fatalf("new catalog entry has no display name: %s", id)
		}
	}
}

func TestMatchSystemFileUsesCatalogScanExtensions(t *testing.T) {
	t.Parallel()

	nes := Systems["NES"]
	if !MatchSystemFile(&nes, "shortcut.mgl") {
		t.Fatal("catalog-added MGL scan extension was not used")
	}
	group, err := GetGroup("Jaguar")
	if err != nil {
		t.Fatal(err)
	}
	if !MatchSystemFile(&group, "game.cdi") {
		t.Fatal("group did not merge catalog scan extensions")
	}
}

// The Atari7800 core loads .bin for both consoles, and the two share the
// ATARI7800 folder. A .bin is a 2600 game only under an Atari2600 folder, so
// 7800 dumps keep resolving to Atari7800 even though Atari2600 has a setname.
func TestAtari2600BinMatchesOnlyItsOwnFolder(t *testing.T) {
	root := filepath.Join(t.TempDir(), "games")
	cfg := &config.UserConfig{Systems: config.SystemsConfig{GamesFolder: []string{root}}}
	atari2600, atari7800 := Systems["Atari2600"], Systems["Atari7800"]
	tests := []struct {
		path string
		want string
	}{
		{filepath.Join(root, "Atari2600", "Adventure.bin"), "Atari2600"},
		{filepath.Join(root, "Atari2600", "Pitfall.a26"), "Atari2600"},
		{filepath.Join(root, "ATARI7800", "Atari 2600", "Combat.bin"), "Atari2600"},
		{filepath.Join(root, "ATARI7800", "Asteroids.bin"), "Atari7800"},
		{filepath.Join(root, "ATARI7800", "Joust.a78"), "Atari7800"},
	}
	for _, test := range tests {
		got, err := BestSystemMatch(cfg, test.path)
		if err != nil {
			t.Fatal(err)
		}
		if got.Id != test.want {
			t.Errorf("BestSystemMatch(%s) = %s, want %s", test.path, got.Id, test.want)
		}
	}
	if MatchSystemFile(&atari2600, filepath.Join(root, "ATARI7800", "Asteroids.bin")) {
		t.Error("Atari2600 matched a .bin outside an Atari2600 folder")
	}
	if !MatchSystemFile(&atari7800, filepath.Join(root, "Atari2600", "Adventure.bin")) {
		t.Error("Atari7800 lost its own .bin extension")
	}
	if MatchSystemFile(&atari2600, filepath.Join(root, "Atari2600", ".Adventure.bin")) {
		t.Error("Atari2600 matched a dot file")
	}
}

func TestLaunchCatalogCoreAddsBinOnlyForAtari2600Folders(t *testing.T) {
	t.Parallel()

	atari2600 := Systems["Atari2600"]
	core := LaunchCatalogCore(&atari2600, "/media/fat/games/Atari2600/Adventure.bin")
	params, err := catalog.PathToMGLDef(&core, "/media/fat/games/Atari2600/Adventure.bin")
	if err != nil {
		t.Fatal(err)
	}
	want, err := catalog.PathToMGLDef(&core, "Adventure.a26")
	if err != nil {
		t.Fatal(err)
	}
	if *params != *want {
		t.Fatalf(".bin params = %#v, want the .a26 slot %#v", params, want)
	}
	outside := LaunchCatalogCore(&atari2600, "/media/fat/games/ATARI7800/Asteroids.bin")
	if !reflect.DeepEqual(outside, CatalogCore(&atari2600)) {
		t.Fatalf("core changed outside an Atari2600 folder: %#v", outside)
	}
	if slots := CatalogCore(&atari2600).Slots; len(slots) == 0 || slices.Contains(slots[0].Exts, ".bin") {
		t.Fatal("LaunchCatalogCore changed the shared catalog slots")
	}
}

func TestCoreGroupsUseCatalogMembership(t *testing.T) {
	t.Parallel()

	group, err := GetGroup("Jaguar")
	if err != nil {
		t.Fatal(err)
	}
	jaguar := Systems["Jaguar"]
	jaguarCD := Systems["JaguarCD"]
	if len(group.Slots) != len(jaguar.Slots)+len(jaguarCD.Slots) {
		t.Fatalf("Jaguar group has %d slots", len(group.Slots))
	}
}
