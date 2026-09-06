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
	"encoding/xml"
	"os"
	"path/filepath"
	"strings"

	"github.com/wizzomafizzo/mrext/pkg/games"
)

// Copyright (c) 2026 The Zaparoo Project Contributors.
// Adapted from Zaparoo Core's pkg/platforms/mister/tracker/arcaderesolve.go
// (ResolveArcadeSetName) and pkg/platforms/mister/mgls/mgls.go (MRA/ReadMRA).
// Keep this boundary small so a future dependency-light Core library can replace
// it. mrext discovers candidates on disk instead of using Core's MediaDB, and
// rejects unreadable candidates rather than using Core's single-candidate guess.
type arcadeMRA struct {
	XMLName xml.Name `xml:"misterromdescription"`
	SetName string   `xml:"setname"`
}

func mraMatches(path, setName string) bool {
	// #nosec G304 -- path comes from installed arcade roots or the resolver cache.
	file, err := os.Open(path)
	if err != nil {
		return false
	}
	defer func() { _ = file.Close() }()
	var mra arcadeMRA
	return xml.NewDecoder(file).Decode(&mra) == nil && strings.EqualFold(strings.TrimSpace(mra.SetName), setName)
}

func resolveArcadeSetName(roots []string, setName string) (string, bool) {
	if setName == "" {
		return "", false
	}
	seen := make(map[string]bool)
	var matches []string
	var visit func(string)
	visit = func(path string) {
		canonical, err := filepath.EvalSymlinks(path)
		if err != nil || seen[canonical] {
			return
		}
		seen[canonical] = true
		info, err := os.Stat(canonical)
		if err != nil {
			return
		}
		if !info.IsDir() {
			if info.Mode().IsRegular() && strings.EqualFold(filepath.Ext(path), ".mra") && mraMatches(path, setName) {
				matches = append(matches, path)
			}
			return
		}
		entries, err := os.ReadDir(path)
		if err != nil {
			return
		}
		for _, entry := range entries {
			visit(filepath.Join(path, entry.Name()))
		}
	}
	for _, root := range roots {
		visit(root)
	}
	if len(matches) != 1 {
		return "", false
	}
	return matches[0], true
}

func (tr *Tracker) lookupArcadeSetPath(setName string) string {
	if path := tr.arcadePaths[setName]; path != "" && mraMatches(path, setName) {
		return path
	}
	var roots []string
	if tr.arcadeRoots != nil {
		roots = tr.arcadeRoots()
	} else {
		system, err := games.GetSystem(ArcadeSystem)
		if err != nil {
			return ""
		}
		paths := games.GetSystemPaths(tr.Config, []games.System{*system})
		for i := range paths {
			roots = append(roots, paths[i].Path)
		}
	}
	path, ok := resolveArcadeSetName(roots, setName)
	if !ok {
		delete(tr.arcadePaths, setName)
		return ""
	}
	if tr.arcadePaths == nil {
		tr.arcadePaths = make(map[string]string)
	}
	tr.arcadePaths[setName] = path
	return path
}

// startArcade keeps the legacy set-name identity and ACTIVEGAME payload while
// supplying the installed MRA path to event consumers such as LastPlayed.
func (tr *Tracker) startArcade(mapping NameMapping) {
	id := tr.ActiveCore
	path := tr.lookupArcadeSetPath(id)
	if tr.ActiveGame == id && tr.ActiveGamePath == path {
		return
	}
	if tr.ActiveGame != id {
		tr.stopGame()
	}
	game, exists := tr.GameTimes[id]
	if !exists {
		loaded, err := tr.Db.GetGame(id)
		if err == nil {
			game = loaded
		} else if !tr.Db.NoResults(err) {
			tr.Logger.Error("error loading arcade game time: %s", err)
		}
	}
	game.Id, game.Path, game.Name = id, path, mapping.ArcadeName
	tr.GameTimes[id] = game
	tr.ActiveGame, tr.ActiveGamePath, tr.ActiveGameName = id, path, mapping.ArcadeName
	tr.addEvent(EventActionGameStart, id)
}
