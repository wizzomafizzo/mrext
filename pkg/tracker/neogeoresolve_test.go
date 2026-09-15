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

//nolint:gosec // Fixtures only access t.TempDir paths.
package tracker

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/wizzomafizzo/mrext/pkg/config"
)

func writeNeoGeoNames(t *testing.T, gamesRoot, title string) string {
	t.Helper()
	root := filepath.Join(gamesRoot, "NeoGeo")
	if err := os.MkdirAll(filepath.Join(root, "aof"), 0o750); err != nil {
		t.Fatal(err)
	}
	xml := `<romsets><romset name="aof" altname="` + title + `"/></romsets>`
	if err := os.WriteFile(filepath.Join(root, "romsets.xml"), []byte(xml), 0o600); err != nil {
		t.Fatal(err)
	}
	return root
}

func neoGeoTracker(gamesRoots ...string) *Tracker {
	return &Tracker{
		Db:     &arcadeDB{},
		Logger: silentLogger{},
		Config: &config.UserConfig{Systems: config.SystemsConfig{GamesFolder: gamesRoots}},
		NameMap: []NameMapping{{
			CoreName: NeoGeoSystem,
			System:   NeoGeoSystem,
			Name:     NeoGeoSystem,
		}},
		GameTimes:     make(map[string]GameTime),
		CoreTimes:     make(map[string]CoreTime),
		ActiveCore:    NeoGeoSystem,
		setActiveGame: func(string) error { return nil },
	}
}

func TestNeoGeoNameUsesFirstConfiguredRootAndReloads(t *testing.T) {
	games1 := filepath.Join(t.TempDir(), "games")
	games2 := filepath.Join(t.TempDir(), "games")
	root1 := writeNeoGeoNames(t, games1, "Art of Fighting")
	writeNeoGeoNames(t, games2, "Wrong Root")
	tr := neoGeoTracker(games1, games2)

	if got := tr.neoGeoName("AOF"); got != "Art of Fighting" {
		t.Fatalf("first configured root title = %q", got)
	}
	writeNeoGeoNames(t, games1, "Art of Fighting Reloaded")
	if got := tr.neoGeoName("aof"); got != "Art of Fighting" {
		t.Fatalf("cached title changed before reload: %q", got)
	}
	tr.ReloadNameMap()
	if got := tr.neoGeoName("aof"); got != "Art of Fighting Reloaded" {
		t.Fatalf("title after reload = %q; romsets path %q", got, root1)
	}
}

func TestProcessGameUsesNeoGeoSetTitle(t *testing.T) {
	gamesRoot := filepath.Join(t.TempDir(), "games")
	neoGeoRoot := writeNeoGeoNames(t, gamesRoot, "Art of Fighting")
	setPath := filepath.Join(neoGeoRoot, "aof")

	tests := []struct {
		name       string
		activeGame func(*testing.T) string
		want       string
	}{
		{
			name: "direct folder",
			activeGame: func(*testing.T) string {
				return setPath
			},
			want: "Art of Fighting",
		},
		{
			name: "MGL folder",
			activeGame: func(t *testing.T) string {
				mgl := filepath.Join(t.TempDir(), "aof.mgl")
				data := `<mistergamedescription><rbf>_Console/NeoGeo</rbf>` +
					`<file path="` + setPath + `"/></mistergamedescription>`
				if err := os.WriteFile(mgl, []byte(data), 0o600); err != nil {
					t.Fatal(err)
				}
				return mgl
			},
			want: "Art of Fighting",
		},
		{
			name: "unknown set",
			activeGame: func(t *testing.T) string {
				path := filepath.Join(neoGeoRoot, "unknown")
				if err := os.Mkdir(path, 0o750); err != nil {
					t.Fatal(err)
				}
				return path
			},
			want: "unknown",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			tr := neoGeoTracker(gamesRoot)
			tr.processGame(test.activeGame(t))
			if tr.ActiveGameName != test.want {
				t.Fatalf("active game name = %q, want %q", tr.ActiveGameName, test.want)
			}
		})
	}
}
