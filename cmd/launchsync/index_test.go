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
	"testing"

	"github.com/wizzomafizzo/mrext/pkg/games"
)

func TestRequiredSystemsDeduplicatesLargePlaylist(t *testing.T) {
	nes, err := games.GetSystem("NES")
	if err != nil {
		t.Fatal(err)
	}
	snes, err := games.GetSystem("SNES")
	if err != nil {
		t.Fatal(err)
	}
	playlist := syncFile{games: make([]syncFileGame, 20000)}
	for index := range playlist.games {
		playlist.games[index].system = nes
	}
	systems := requiredSystems([]syncFile{playlist, playlist, {games: []syncFileGame{{system: snes}}}})
	if len(systems) != 2 || systems[0].Id != "NES" || systems[1].Id != "SNES" {
		t.Fatalf("repeated playlist entries expanded system list: %+v", systems)
	}
	if len(requiredSystems(nil)) != 0 {
		t.Fatal("empty playlist should not request indexing")
	}
}
