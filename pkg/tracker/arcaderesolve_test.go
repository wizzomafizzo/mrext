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
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/wizzomafizzo/mrext/pkg/config"
)

func writeArcade(t *testing.T, path, setName string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	data := "<misterromdescription><name>Same Name</name><setname>" + setName + "</setname></misterromdescription>"
	if err := os.WriteFile(path, []byte(data), 0o600); err != nil {
		t.Fatal(err)
	}
}

// Copyright (c) 2026 The Zaparoo Project Contributors.
// Matching/ambiguity cases adapted from Core's
// pkg/platforms/mister/tracker/arcaderesolve_test.go, using filesystem fixtures
// instead of MediaDB mocks so they remain suitable for future library extraction.
func TestResolveArcadeSetName(t *testing.T) {
	root := t.TempDir()
	wanted := filepath.Join(root, "nested", "game.mra")
	writeArcade(t, wanted, "POOYAN")
	writeArcade(t, filepath.Join(root, "clone.mra"), "pooyana")
	if err := os.Symlink(wanted, filepath.Join(root, "alias.mra")); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(root, filepath.Join(root, "nested", "cycle")); err != nil {
		t.Fatal(err)
	}
	path, ok := resolveArcadeSetName([]string{root, root}, "pooyan")
	if !ok {
		t.Fatal("nested exact set-name match was not resolved")
	}
	canonical, err := filepath.EvalSymlinks(path)
	if err != nil || canonical != wanted {
		t.Fatalf("path=%q err=%v", path, err)
	}
	writeArcade(t, filepath.Join(root, "duplicate.mra"), "pooyan")
	if _, found := resolveArcadeSetName([]string{root}, "pooyan"); found {
		t.Fatal("ambiguous set name was guessed")
	}
	if _, found := resolveArcadeSetName([]string{root}, "missing"); found {
		t.Fatal("unmatched set name resolved")
	}
	if _, found := resolveArcadeSetName([]string{root}, ""); found {
		t.Fatal("empty set name resolved")
	}
}

func TestResolverRejectsMalformedAndMissingFiles(t *testing.T) {
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "game.mra"), []byte("<bad>"), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, ok := resolveArcadeSetName([]string{root, filepath.Join(root, "missing")}, "game"); ok {
		t.Fatal("malformed single candidate was guessed")
	}
}

type silentLogger struct{}

func (silentLogger) Info(string, ...any)  {}
func (silentLogger) Warn(string, ...any)  {}
func (silentLogger) Error(string, ...any) {}

type arcadeDB struct {
	Db
	events []EventAction
}

func (d *arcadeDB) AddEvent(event *EventAction) error {
	d.events = append(d.events, *event)
	return nil
}

func (*arcadeDB) GetCore(name string) (CoreTime, error) { return CoreTime{Name: name}, nil }
func (*arcadeDB) GetGame(string) (GameTime, error)      { return GameTime{}, errors.New("missing") }
func (*arcadeDB) NoResults(err error) bool              { return err != nil }
func (*arcadeDB) UpdateGame(GameTime) error             { return nil }
func (*arcadeDB) UpdateCore(CoreTime) error             { return nil }

func arcadeTracker(root string, db *arcadeDB) *Tracker {
	return &Tracker{
		Db: db, Logger: silentLogger{}, Config: &config.UserConfig{},
		NameMap:   []NameMapping{{CoreName: "pooyan", System: ArcadeSystem, Name: ArcadeSystem, ArcadeName: "Pooyan"}},
		GameTimes: make(map[string]GameTime), CoreTimes: make(map[string]CoreTime),
		arcadeRoots:   func() []string { return []string{root} },
		setActiveGame: func(string) error { return nil },
	}
}

func TestArcadeLifecycleProvidesPathOnceAndPreservesIdentity(t *testing.T) {
	root := t.TempDir()
	path := filepath.Join(root, "nested", "Pooyan.mra")
	writeArcade(t, path, "pooyan")
	db := &arcadeDB{}
	tr := arcadeTracker(root, db)
	var signal string
	tr.setActiveGame = func(value string) error { signal = value; return nil }
	tr.processCore("pooyan")
	if signal != "pooyan" {
		t.Fatalf("legacy ACTIVEGAME payload changed: %q", signal)
	}
	if len(db.events) != 2 {
		t.Fatalf("expected core and game start, got %v", db.events)
	}
	event := db.events[1]
	if event.TargetPath != path || event.Target != "pooyan" || event.ActiveGame.Path != path ||
		event.ActiveCore.System != ArcadeSystem || event.ActiveCore.Core != "pooyan" {
		t.Fatalf("incomplete arcade start event: %+v", event)
	}
	tr.processGame("pooyan")
	if len(db.events) != 2 {
		t.Fatal("ACTIVEGAME notification duplicated arcade start")
	}
	game := tr.GameTimes["pooyan"]
	game.Time = 47
	tr.GameTimes["pooyan"] = game
	tr.processCore(config.MenuCore)
	if tr.ActiveGame != "" || tr.ActiveGamePath != "" || signal != "" {
		t.Fatal("arcade state survived return to Menu")
	}
	if len(db.events) != 4 || db.events[2].Action != EventActionGameStop || db.events[2].TargetPath != path {
		t.Fatalf("arcade stop missing: %+v", db.events)
	}
	tr.processCore("pooyan")
	if tr.GameTimes["pooyan"].Time != 47 {
		t.Fatal("returning to arcade game lost accumulated play time")
	}
}

func TestConsoleTrackingDoesNotScanArcadeRoots(t *testing.T) {
	root := t.TempDir()
	db := &arcadeDB{}
	tr := arcadeTracker(root, db)
	tr.Config.Systems.GamesFolder = []string{root}
	tr.NameMap = []NameMapping{{CoreName: "NES", System: "NES", Name: "NES"}}
	tr.arcadeRoots = func() []string {
		t.Fatal("console launch scanned arcade files")
		return nil
	}
	path := filepath.Join(root, "games", "NES", "game.nes")
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, nil, 0o600); err != nil {
		t.Fatal(err)
	}
	tr.processCore("NES")
	tr.processGame(path)
	if tr.ActiveSystem != "NES" || tr.ActiveGamePath != path || len(db.events) != 2 {
		t.Fatalf("console tracking regressed: %+v", db.events)
	}
	tr.processGame(path)
	if len(db.events) != 2 {
		t.Fatal("same console game emitted duplicate start")
	}
}

func TestMissingArcadeMappingDoesNotInventGame(t *testing.T) {
	db := &arcadeDB{}
	tr := arcadeTracker(t.TempDir(), db)
	tr.NameMap = nil
	tr.arcadeRoots = func() []string {
		t.Fatal("unmapped core triggered arcade scan")
		return nil
	}
	tr.processCore("pooyan")
	if tr.ActiveGame != "" || len(db.events) != 1 || db.events[0].Action != EventActionCoreStart {
		t.Fatalf("missing Arcade Database mapping invented game state: %+v", db.events)
	}
}

func TestArcadeMissRetriedAndRemovedCachedPathDiscarded(t *testing.T) {
	root := t.TempDir()
	tr := arcadeTracker(root, &arcadeDB{})
	tr.processCore("pooyan")
	if tr.GameTimes["pooyan"].Path != "" {
		t.Fatal("missing MRA produced a target")
	}
	path := filepath.Join(root, "Pooyan.mra")
	writeArcade(t, path, "pooyan")
	tr.processGame("pooyan")
	if tr.ActiveGamePath != path {
		t.Fatal("miss was cached permanently")
	}
	if err := os.Remove(path); err != nil {
		t.Fatal(err)
	}
	if got := tr.lookupArcadeSetPath("pooyan"); got != "" {
		t.Fatalf("removed cached MRA retained: %q", got)
	}
}
