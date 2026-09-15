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

package tracker

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/wizzomafizzo/mrext/pkg/games"
)

const NeoGeoSystem = "NeoGeo"

// neoGeoName returns the title MiSTer shows for a NeoGeo ROM set folder,
// read from the romsets.xml beside the games, or "" when the set is unknown.
// The names load on first use and are dropped with the name map, so a
// changed romsets.xml is picked up on the next reload. Only the NeoGeo game
// roots are read, and only their romsets.xml; no game folder is scanned.
func (tr *Tracker) neoGeoName(set string) string {
	if tr.neoGeoNames == nil {
		tr.neoGeoNames = tr.loadNeoGeoNames()
	}
	return tr.neoGeoNames[strings.ToLower(strings.TrimSpace(set))]
}

func (tr *Tracker) loadNeoGeoNames() map[string]string {
	names := make(map[string]string)
	system, err := games.GetSystem(NeoGeoSystem)
	if err != nil {
		return names
	}
	roots := games.GetSystemPaths(tr.Config, []games.System{*system})
	for i := range roots {
		path, findErr := games.FindFile(filepath.Join(roots[i].Path, "romsets.xml"))
		if findErr != nil {
			continue
		}
		// #nosec G304 -- path is romsets.xml under a configured NeoGeo game root.
		file, openErr := os.Open(path)
		if openErr != nil {
			continue
		}
		parsed, readErr := games.ReadNeoGeoNames(file)
		_ = file.Close()
		if readErr != nil {
			continue
		}
		for set, title := range parsed {
			if _, seen := names[set]; !seen {
				names[set] = title
			}
		}
	}
	return names
}
