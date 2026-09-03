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
	"io/fs"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/ZaparooProject/zaparoo-core/mister/catalog"
)

type missingSource struct{}

func (missingSource) ReadFile(string) ([]byte, error) {
	return nil, fs.ErrNotExist
}

func TestParseAliases(t *testing.T) {
	t.Parallel()

	source := []byte(`package systemdefs
const SystemGenesis = "Genesis"
var Systems = map[string]System{
	SystemGenesis: {
		ID: SystemGenesis,
		Aliases: []string{"MegaDrive", "Mega Drive"},
	},
}
`)
	aliases, err := parseAliases(source)
	if err != nil {
		t.Fatal(err)
	}
	want := map[string][]string{"Genesis": {"MegaDrive", "Mega Drive"}}
	if !reflect.DeepEqual(aliases, want) {
		t.Fatalf("aliases = %#v, want %#v", aliases, want)
	}
}

func TestMetadataForSystemFallsBackWhenAssetIsAbsent(t *testing.T) {
	t.Parallel()

	system := &catalog.Core{ID: "LegacyOnly"}
	entry, err := metadataForSystem(
		missingSource{},
		system,
		map[string][]string{"LegacyOnly": {"Legacy Alias"}},
	)
	if err != nil {
		t.Fatal(err)
	}
	if entry.Name != system.ID || entry.Category != "" {
		t.Fatalf("fallback metadata = %#v", entry)
	}
	if !reflect.DeepEqual(entry.Aliases, []string{"Legacy Alias"}) {
		t.Fatalf("fallback aliases = %#v", entry.Aliases)
	}
}

func TestGeneratedOutputCurrent(t *testing.T) {
	t.Parallel()

	path := filepath.Join(t.TempDir(), "metadata.json")
	data := []byte(`{"source":"` + pinnedSourceLabel() + `","format":1,"systems":{}}`)
	if err := os.WriteFile(path, data, 0o600); err != nil {
		t.Fatal(err)
	}
	if !generatedOutputCurrent(path) {
		t.Fatal("current generated output was not recognized")
	}
}
