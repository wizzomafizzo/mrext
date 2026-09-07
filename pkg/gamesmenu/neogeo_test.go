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
	"encoding/xml"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestGenerateNeoGeoDarksoftSets(t *testing.T) {
	m := fixture(t)
	external := t.TempDir()
	m.cfg.Systems.GamesFolder = append(m.cfg.Systems.GamesFolder, external)
	m.cfg.Systems.SetCore = []string{"NeoGeo:_Console/CustomNeoGeo"}
	// A definition on SD also describes sets stored on USB or network roots.
	put(t, filepath.Join(m.paths.SDRoot, "games", "NEOGEO", "ROMSETS.XML"),
		`<romsets><romset name="mslug, MS1" altname="Metal Slug"/>`+
			`<romset name="samsho" altname="Samurai/Shodown: Special"/></romsets>`)
	base := filepath.Join(external, "games", "NEOGEO")
	set := filepath.Join(base, `Picks & "More"`, "MS1")
	put(t, filepath.Join(set, "crom0"), "rom")
	put(t, filepath.Join(set, "prom.bin"), "component, not a NeoGeoCD game")
	put(t, filepath.Join(set, "nested", "wrong.neo"), "not a separate game")
	createZip(t, filepath.Join(base, "SAMSHO.ZIP"), "crom0", "prom.bin", "wrong.neo")
	createZip(t, filepath.Join(base, "collection.zip"), "loose.neo")
	put(t, filepath.Join(base, "ordinary.neo"), "rom")
	put(t, filepath.Join(base, "unknown", "crom0"), "rom")
	entries := discover(t, m)
	if len(entries) != 1 {
		t.Fatalf("entries: %+v", entries)
	}
	result := generate(t, m, entries)
	if result.Scan.Created != 4 {
		t.Fatalf("expected two sets and two ordinary games: %+v", result)
	}
	output := filepath.Join(m.paths.MenuFolder, entries[0].MenuFolder)
	checks := []struct{ shortcut, target string }{
		{filepath.Join(`_Picks & "More"`, "Metal Slug.mgl"), set},
		{"Samurai & Shodown  Special.mgl", filepath.Join(base, "SAMSHO.ZIP")},
		{"loose.mgl", filepath.Join(base, "collection.zip") + "/loose.neo"},
		{"ordinary.mgl", filepath.Join(base, "ordinary.neo")},
	}
	type mglFile struct {
		Path  string `xml:"path,attr"`
		Type  string `xml:"type,attr"`
		Index int    `xml:"index,attr"`
		Delay int    `xml:"delay,attr"`
	}
	for _, check := range checks {
		data := read(t, filepath.Join(output, check.shortcut))
		var doc struct {
			RBF  string  `xml:"rbf"`
			File mglFile `xml:"file"`
		}
		if err := xml.Unmarshal([]byte(data), &doc); err != nil {
			t.Fatalf("invalid MGL: %s: %v", data, err)
		}
		if doc.RBF != "_Console/CustomNeoGeo" || doc.File.Path != "../../../../.."+check.target ||
			doc.File.Type != "f" || doc.File.Index != 1 || doc.File.Delay != 1 {
			t.Fatalf("unexpected MGL: %s", data)
		}
	}
	result = generate(t, m, entries)
	if result.Scan.Created != 0 || result.Scan.Skipped != 4 {
		t.Fatalf("regeneration: %+v", result)
	}
	cleaned, err := m.CleanUp(nil)
	if err != nil || cleaned.Checked != 4 || cleaned.Removed != 0 || cleaned.Unreadable != 0 {
		t.Fatalf("existing sets must survive cleanup: %+v, %v", cleaned, err)
	}
	if removeErr := os.RemoveAll(set); removeErr != nil {
		t.Fatal(removeErr)
	}
	if removeErr := os.Remove(filepath.Join(base, "SAMSHO.ZIP")); removeErr != nil {
		t.Fatal(removeErr)
	}
	cleaned, err = m.CleanUp(nil)
	if err != nil || cleaned.Removed != 2 || cleaned.Unreadable != 0 {
		t.Fatalf("missing sets must be removed: %+v, %v", cleaned, err)
	}
}

func TestGenerateNeoGeoSymlinksCollisionsAndRootPriority(t *testing.T) {
	m := fixture(t)
	external := t.TempDir()
	m.cfg.Systems.GamesFolder = append(m.cfg.Systems.GamesFolder, external)
	base := filepath.Join(m.paths.SDRoot, "games", "NEOGEO")
	put(t, filepath.Join(base, "romsets.xml"),
		`<romsets><romset name="set,link.with.dot" altname="Game v1.0"/></romsets>`)
	put(t, filepath.Join(external, "games", "NEOGEO", "romsets.xml"),
		`<romsets><romset name="set" altname="Wrong title"/></romsets>`)
	put(t, filepath.Join(external, "games", "NEOGEO", "set", "crom0"), "rom")
	target := filepath.Join(external, "ROM data")
	put(t, filepath.Join(target, "crom0"), "rom")
	if err := os.Symlink(target, filepath.Join(base, "link.with.dot")); err != nil {
		t.Fatal(err)
	}
	createZip(t, filepath.Join(base, "set.zip"), "crom0")
	createZip(t, filepath.Join(base, ".hidden.zip"), "ignored.neo")
	entries := discover(t, m)
	result := generate(t, m, entries)
	if result.Scan.Created != 1 || result.Scan.Skipped != 2 {
		t.Fatalf("mapped aliases must collide without exposing components: %+v", result)
	}
	shortcut := filepath.Join(m.paths.MenuFolder, entries[0].MenuFolder, "Game v1.0.mgl")
	original := read(t, shortcut)
	if !strings.Contains(original, "link.with.dot") {
		t.Fatalf("expected linked directory mount, got %s", original)
	}
	put(t, shortcut, "user content")
	result = generate(t, m, entries)
	if result.Scan.Created != 0 || result.Scan.Skipped != 3 || read(t, shortcut) != "user content" {
		t.Fatalf("custom set shortcut must not be replaced: %+v", result)
	}
}

func TestGenerateNeoGeoWithoutValidMapping(t *testing.T) {
	for _, metadata := range []string{"", `<romsets><romset name="set" altname="Partial"/>`} {
		t.Run(metadata, func(t *testing.T) {
			m := fixture(t)
			base := filepath.Join(m.paths.SDRoot, "games", "NEOGEO")
			put(t, filepath.Join(base, "set", "crom0"), "rom")
			createZip(t, filepath.Join(base, "set.zip"), "crom0")
			put(t, filepath.Join(base, "ordinary.neo"), "rom")
			if metadata != "" {
				put(t, filepath.Join(base, "romsets.xml"), metadata)
			}
			result, err := m.Generate(discover(t, m), nil)
			if err != nil || result.Scan.Created != 1 {
				t.Fatalf("ordinary .neo must still work: %+v, %v", result, err)
			}
			if metadata == "" && result.Scan.Failed != 0 {
				t.Fatalf("missing optional mapping: %+v", result)
			}
			if metadata != "" && (result.Scan.Failed != 1 ||
				!strings.Contains(result.Scan.Errors[0].Error(), "romsets.xml")) {
				t.Fatalf("invalid mapping must be reported without partial set recognition: %+v", result)
			}
		})
	}
}
