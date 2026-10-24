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
	"io"
	"os"
	"strings"
	"time"
)

// mraHeadLimit bounds how much of an .mra is read for its set name. The
// <setname> sits near the top, ahead of the <rom> parts, which can run to
// megabytes; a set name that is not within this much counts as absent.
const mraHeadLimit = 64 << 10

// openMRA opens an .mra for reading; replaced in tests.
var openMRA = func(path string) (io.ReadCloser, error) {
	// #nosec G304 -- path comes from installed arcade roots or the resolver cache.
	return os.Open(path)
}

// Copyright (c) 2026 The Zaparoo Project Contributors.
// Adapted from Zaparoo Core's pkg/platforms/mister/tracker/arcaderesolve.go
// (ResolveArcadeSetName) and pkg/platforms/mister/mgls/mgls.go (MRA/ReadMRA).
// Keep this boundary small so a future dependency-light Core library can replace
// it. mrext discovers candidates on disk instead of using Core's MediaDB, and
// rejects unreadable candidates rather than using Core's single-candidate guess.
//
// readMRASetName returns the <setname> of an .mra's <misterromdescription>
// root, or "" when the file cannot be read, is not an MRA or has none. It
// stops at the set name rather than decoding the whole file, and reads no
// further than mraHeadLimit, so a truncated head means no set name rather
// than a guess.
func readMRASetName(path string) string {
	file, err := openMRA(path)
	if err != nil {
		return ""
	}
	defer func() { _ = file.Close() }()
	dec := xml.NewDecoder(io.LimitReader(file, mraHeadLimit))
	for {
		tok, err := dec.Token()
		if err != nil {
			return ""
		}
		if start, ok := tok.(xml.StartElement); ok {
			if start.Name.Local != "misterromdescription" {
				return ""
			}
			break
		}
	}
	for {
		tok, err := dec.Token()
		if err != nil {
			return ""
		}
		switch elem := tok.(type) {
		case xml.StartElement:
			if elem.Name.Local == "setname" {
				return readElementText(dec)
			}
			if err := dec.Skip(); err != nil {
				return ""
			}
		case xml.EndElement:
			// The root closed without a set name.
			return ""
		}
	}
}

// readElementText collects the character data of the element just opened,
// which can arrive in several pieces, up to its end tag.
func readElementText(dec *xml.Decoder) string {
	var text strings.Builder
	for {
		tok, err := dec.Token()
		if err != nil {
			return ""
		}
		switch elem := tok.(type) {
		case xml.CharData:
			_, _ = text.Write(elem)
		case xml.StartElement:
			if err := dec.Skip(); err != nil {
				return ""
			}
		case xml.EndElement:
			return strings.TrimSpace(text.String())
		}
	}
}

func mraMatches(path, setName string) bool {
	return setName != "" && strings.EqualFold(readMRASetName(path), setName)
}

// resolveArcadeSetName returns the only .mra under roots that carries
// setName. Where Core looks set names up in its MediaDB, mrext walks the
// arcade folders into an index, which the tracker builds off its lock.
func resolveArcadeSetName(roots []string, setName string) (string, bool) {
	if setName == "" {
		return "", false
	}
	paths := buildArcadeIndex(roots).paths(setName)
	if len(paths) != 1 {
		return "", false
	}
	return paths[0], true
}

// lookupArcadeSetPath returns the installed .mra for setName from the
// arcade index, without touching the disk beyond confirming that file.
// known is false while the first index is still being built, and the path
// is empty for a set no single file carries. A miss asks for a fresh index,
// at most once per rebuildInterval, in case the files changed.
func (tr *Tracker) lookupArcadeSetPath(setName string) (path string, known bool) {
	return tr.findArcadeSetPath(setName, true)
}

// findArcadeSetPath runs under the tracker lock; mayRequest is false for
// lookups made while installing an index, which never ask for another.
func (tr *Tracker) findArcadeSetPath(setName string, mayRequest bool) (path string, known bool) {
	if setName == "" {
		return "", true
	}
	idx := tr.arcadeIdx.Load()
	if idx == nil {
		if mayRequest {
			tr.requestArcadeIndex()
		}
		return "", false
	}
	if paths := idx.paths(setName); len(paths) == 1 && mraMatches(paths[0], setName) {
		return paths[0], true
	}
	if mayRequest && !tr.indexBusy.Load() && time.Since(tr.lastBuildStart) >= tr.rebuildInterval {
		tr.requestArcadeIndex()
	}
	return "", true
}

// startArcade keeps the legacy set-name identity and ACTIVEGAME payload while
// supplying the installed MRA path to event consumers such as LastPlayed.
// A set already running never starts again: a path found later only fills
// in the running game, so its session time carries on.
func (tr *Tracker) startArcade(mapping NameMapping) {
	id := tr.ActiveCore
	path, _ := tr.lookupArcadeSetPath(id)
	if tr.ActiveGame == id {
		if path != "" && path != tr.ActiveGamePath {
			tr.setArcadePath(id, path)
		}
		return
	}
	tr.stopGame()
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

// setArcadePath gives the running arcade game the path found for it,
// without an event.
func (tr *Tracker) setArcadePath(id, path string) {
	tr.ActiveGamePath = path
	game := tr.GameTimes[id]
	game.Path = path
	tr.GameTimes[id] = game
}
