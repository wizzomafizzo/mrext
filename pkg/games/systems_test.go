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
	"reflect"
	"strings"
	"testing"

	"github.com/ZaparooProject/zaparoo-core/mister/catalog"
)

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
