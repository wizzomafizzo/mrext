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
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"sync/atomic"
	"testing"

	"github.com/wizzomafizzo/mrext/pkg/config"
)

// snapshotDB checks, as each event is saved, that Playing already reports
// the state the event carries.
type snapshotDB struct {
	tr *Tracker
	arcadeDB
	mismatches []string
}

func (d *snapshotDB) AddEvent(ev *EventAction) error {
	now := d.tr.Playing()
	if now.Core != ev.ActiveCore.Core || now.System != ev.ActiveCore.System ||
		now.SystemName != ev.ActiveCore.SystemName || now.GamePath != ev.ActiveGame.Path ||
		now.GameName != ev.ActiveGame.Name {
		d.mismatches = append(d.mismatches, fmt.Sprintf("event %d %q: playing %+v", ev.Action, ev.Target, now))
	}
	return d.arcadeDB.AddEvent(ev)
}

func TestPlayingSnapshot(t *testing.T) {
	root := t.TempDir()
	writeArcade(t, filepath.Join(root, "Pooyan.mra"), "pooyan")
	nes := filepath.Join(root, "games", "NES", "game.nes")
	if err := os.MkdirAll(filepath.Dir(nes), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(nes, nil, 0o600); err != nil {
		t.Fatal(err)
	}
	db := &snapshotDB{}
	tr := arcadeTracker(root, &db.arcadeDB)
	db.tr = tr
	tr.Db = db
	tr.Config.Systems.GamesFolder = []string{root}
	tr.NameMap = append(tr.NameMap, NameMapping{CoreName: "NES", System: "NES", Name: "NES"})
	if now := tr.Playing(); now != (Playing{}) {
		t.Fatalf("playing before the first update = %+v", now)
	}
	readyIndex(t, tr)

	// A reader polling as a Remote client would.
	var stop atomic.Bool
	var torn atomic.Value
	var wg sync.WaitGroup
	wg.Go(func() {
		for !stop.Load() {
			now := tr.Playing()
			if now.Core == "pooyan" && now.System != ArcadeSystem ||
				now.Core == "NES" && now.Game != "" && now.System != "NES" {
				torn.Store(now)
			}
		}
	})
	update := func(change func()) {
		tr.mu.Lock()
		defer tr.mu.Unlock()
		defer tr.updateSessions()
		change()
	}
	for range 50 {
		update(func() { tr.processCore("NES") })
		update(func() { tr.processGame(nes) })
		update(func() { tr.processCore("pooyan") })
		update(func() { tr.processGame("pooyan") })
		update(func() { tr.processCore(config.MenuCore) })
	}
	update(func() { tr.processCore("pooyan") })
	stop.Store(true)
	wg.Wait()

	if now, ok := torn.Load().(Playing); ok {
		t.Fatalf("a reader saw a core without its system: %+v", now)
	}
	if len(db.mismatches) > 0 {
		t.Fatalf("events went out before Playing matched them: %v", db.mismatches[:min(3, len(db.mismatches))])
	}
	want := Playing{
		Core: "pooyan", System: ArcadeSystem, SystemName: ArcadeSystem,
		Game: "pooyan", GameName: "Pooyan", GamePath: filepath.Join(root, "Pooyan.mra"),
	}
	if now := tr.Playing(); now != want {
		t.Fatalf("playing = %+v, want %+v", now, want)
	}
}
