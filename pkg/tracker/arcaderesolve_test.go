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
	"fmt"
	"io"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"sync"
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
	tr := &Tracker{
		Db: db, Logger: silentLogger{}, Config: &config.UserConfig{},
		NameMap:   []NameMapping{{CoreName: "pooyan", System: ArcadeSystem, Name: ArcadeSystem, ArcadeName: "Pooyan"}},
		GameTimes: make(map[string]GameTime), CoreTimes: make(map[string]CoreTime),
		arcadeRoots:   func() []string { return []string{root} },
		setActiveGame: func(string) error { return nil },
		rbfName:       func() string { return "" },
		currentCore:   func() string { return "" },
	}
	queueBuilds(tr)
	return tr
}

// buildQueue holds the index builds a test tracker asks for until the test
// runs them with drainIndex, outside the tracker lock.
type buildQueue struct {
	jobs []func()
	mu   sync.Mutex
}

var buildQueues sync.Map // *Tracker to *buildQueue

func queueBuilds(tr *Tracker) {
	queue := &buildQueue{}
	buildQueues.Store(tr, queue)
	tr.runBuild = func(job func()) {
		queue.mu.Lock()
		defer queue.mu.Unlock()
		queue.jobs = append(queue.jobs, job)
	}
}

func trackerQueue(t *testing.T, tr *Tracker) *buildQueue {
	t.Helper()
	queue, ok := buildQueues.Load(tr)
	if !ok {
		t.Fatal("tracker has no build queue")
	}
	typed, ok := queue.(*buildQueue)
	if !ok {
		t.Fatal("tracker build queue has the wrong type")
	}
	return typed
}

// pendingBuilds counts the index builds asked for and not yet run.
func pendingBuilds(t *testing.T, tr *Tracker) int {
	t.Helper()
	queue := trackerQueue(t, tr)
	queue.mu.Lock()
	defer queue.mu.Unlock()
	return len(queue.jobs)
}

// drainIndex runs the queued index builds.
func drainIndex(t *testing.T, tr *Tracker) {
	t.Helper()
	queue := trackerQueue(t, tr)
	for {
		queue.mu.Lock()
		if len(queue.jobs) == 0 {
			queue.mu.Unlock()
			return
		}
		job := queue.jobs[0]
		queue.jobs = queue.jobs[1:]
		queue.mu.Unlock()
		job()
	}
}

// readyIndex builds and installs the arcade index before the test starts.
func readyIndex(t *testing.T, tr *Tracker) {
	t.Helper()
	tr.mu.Lock()
	tr.warmArcadeIndex()
	tr.mu.Unlock()
	drainIndex(t, tr)
}

func TestArcadeLifecycleProvidesPathOnceAndPreservesIdentity(t *testing.T) {
	root := t.TempDir()
	path := filepath.Join(root, "nested", "Pooyan.mra")
	writeArcade(t, path, "pooyan")
	db := &arcadeDB{}
	tr := arcadeTracker(root, db)
	var signal string
	tr.setActiveGame = func(value string) error { signal = value; return nil }
	readyIndex(t, tr)
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
	if n := pendingBuilds(t, tr); n != 0 {
		t.Fatalf("console launch asked for %d arcade index builds", n)
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
	if n := pendingBuilds(t, tr); n != 0 {
		t.Fatalf("unmapped core asked for %d arcade index builds", n)
	}
}

func TestArcadeMissRetriedAndRemovedCachedPathDiscarded(t *testing.T) {
	root := t.TempDir()
	db := &arcadeDB{}
	tr := arcadeTracker(root, db)
	tr.processCore("pooyan")
	drainIndex(t, tr)
	if tr.GameTimes["pooyan"].Path != "" || tr.ActiveGamePath != "" {
		t.Fatal("missing MRA produced a target")
	}
	events := len(db.events)
	path := filepath.Join(root, "Pooyan.mra")
	writeArcade(t, path, "pooyan")
	// The ACTIVEGAME echo misses and asks for a fresh index.
	tr.processGame("pooyan")
	drainIndex(t, tr)
	if tr.ActiveGamePath != path || tr.GameTimes["pooyan"].Path != path {
		t.Fatalf("miss was cached permanently: path=%q", tr.ActiveGamePath)
	}
	if len(db.events) != events {
		t.Fatalf("a path found later started the running game again: %+v", db.events[events:])
	}
	if err := os.Remove(path); err != nil {
		t.Fatal(err)
	}
	if got, known := tr.lookupArcadeSetPath("pooyan"); got != "" || !known {
		t.Fatalf("removed cached MRA retained: %q known=%t", got, known)
	}
	if n := pendingBuilds(t, tr); n != 1 {
		t.Fatalf("a removed MRA asked for %d index builds, want 1", n)
	}
}

// A handful of arcade database rows carry no setname. Left in the name map
// they give every lookup of an empty core name an arcade hit, so the tracker
// reports whichever of those rows came first as the running game.
func TestEmptyCoreNameMatchesNothing(t *testing.T) {
	db := &arcadeDB{}
	tr := arcadeTracker(t.TempDir(), db)
	tr.NameMap = []NameMapping{
		{CoreName: "", System: ArcadeSystem, Name: ArcadeSystem, ArcadeName: "Arkanoid [hb]"},
		{CoreName: "pooyan", System: ArcadeSystem, Name: ArcadeSystem, ArcadeName: "Pooyan"},
	}
	if got := tr.LookupName("", "/media/fat/_Arcade/Whatever.mra"); got != (NameMapping{}) {
		t.Fatalf("empty core name matched %+v", got)
	}
	if got := tr.LookupName("pooyan", ""); got.ArcadeName != "Pooyan" {
		t.Fatalf("real lookup regressed: %+v", got)
	}
}

func TestReadMRASetName(t *testing.T) {
	root := t.TempDir()
	cases := []struct {
		name, data, want string
	}{
		{
			"declaration comment and name first",
			`<?xml version="1.0" encoding="UTF-8"?><!-- generated --><misterromdescription>` +
				`<name>Pooyan</name><setname>pooyan</setname></misterromdescription>`, "pooyan",
		},
		{
			"uppercase with whitespace",
			"<misterromdescription>\n\t<setname>\n  POOYAN  \n</setname></misterromdescription>", "POOYAN",
		},
		{"split by an entity", "<misterromdescription><setname>pooy&#97;n</setname></misterromdescription>", "pooyan"},
		{"wrong root", "<bad>", ""},
		{"wrong root with setname", "<other><setname>pooyan</setname></other>", ""},
		{"nested setname", "<misterromdescription><rom><setname>pooyan</setname></rom></misterromdescription>", ""},
		{"no setname", "<misterromdescription><name>Pooyan</name></misterromdescription>", ""},
		{"empty file", "", ""},
		{
			"malformed tail after setname",
			"<misterromdescription><setname>pooyan</setname><rom><part></rom>", "pooyan",
		},
	}
	for i, tc := range cases {
		path := filepath.Join(root, fmt.Sprintf("%d.mra", i))
		if err := os.WriteFile(path, []byte(tc.data), 0o600); err != nil {
			t.Fatal(err)
		}
		if got := readMRASetName(path); got != tc.want {
			t.Errorf("%s: set name = %q, want %q", tc.name, got, tc.want)
		}
	}
	if got := readMRASetName(filepath.Join(root, "missing.mra")); got != "" {
		t.Errorf("missing file: set name = %q", got)
	}
}

type countingReader struct {
	io.ReadCloser
	read *int
}

func (c countingReader) Read(p []byte) (int, error) {
	n, err := c.ReadCloser.Read(p)
	*c.read += n
	return n, err //nolint:wrapcheck // A reader passes io.EOF through as is.
}

func TestReadMRASetNameReadsOnlyTheHead(t *testing.T) {
	root := t.TempDir()
	big := "<misterromdescription><setname>pooyan</setname><rom>" +
		strings.Repeat("<part>00 01 02 03 04 05 06 07</part>\n", 5<<20/37) + "</rom></misterromdescription>"
	path := filepath.Join(root, "big.mra")
	if err := os.WriteFile(path, []byte(big), 0o600); err != nil {
		t.Fatal(err)
	}
	read := 0
	original := openMRA
	openMRA = func(path string) (io.ReadCloser, error) {
		file, err := original(path)
		if err != nil {
			return nil, err
		}
		return countingReader{ReadCloser: file, read: &read}, nil
	}
	t.Cleanup(func() { openMRA = original })
	if got := readMRASetName(path); got != "pooyan" {
		t.Fatalf("set name = %q", got)
	}
	if read > mraHeadLimit+4<<10 {
		t.Fatalf("read %d bytes of a %d byte MRA for its set name", read, len(big))
	}

	late := "<misterromdescription><name>" + strings.Repeat("x", mraHeadLimit) +
		"</name><setname>pooyan</setname></misterromdescription>"
	if err := os.WriteFile(path, []byte(late), 0o600); err != nil {
		t.Fatal(err)
	}
	if got := readMRASetName(path); got != "" {
		t.Fatalf("set name past the head limit = %q", got)
	}
}

func TestUnlistedArcadeSetIsTrackedFromItsMRA(t *testing.T) {
	root := t.TempDir()
	path := filepath.Join(root, "Bomb Jack Dualbit.mra")
	writeArcade(t, path, "bombjack_dualbit")
	db := &arcadeDB{}
	tr := arcadeTracker(root, db)
	tr.NameMap = nil
	tr.rbfName = func() string { return "BombJack" }
	var signal string
	eventsAtSignal := -1
	tr.setActiveGame = func(value string) error {
		signal, eventsAtSignal = value, len(db.events)
		return nil
	}
	readyIndex(t, tr)
	tr.processCore("bombjack_dualbit")
	if tr.ActiveSystem != ArcadeSystem || tr.ActiveGame != "bombjack_dualbit" ||
		tr.ActiveGamePath != path || tr.ActiveGameName != "Bomb Jack Dualbit" {
		t.Fatalf("unlisted arcade set not tracked: system=%q game=%q path=%q name=%q",
			tr.ActiveSystem, tr.ActiveGame, tr.ActiveGamePath, tr.ActiveGameName)
	}
	if len(db.events) != 2 || db.events[1].Action != EventActionGameStart || db.events[1].TargetPath != path {
		t.Fatalf("expected core and game start, got %+v", db.events)
	}
	if signal != "bombjack_dualbit" || eventsAtSignal != 0 {
		t.Fatalf("ACTIVEGAME = %q after %d events, want the set name before the core start", signal, eventsAtSignal)
	}
	tr.processGame("bombjack_dualbit")
	if len(db.events) != 2 {
		t.Fatal("ACTIVEGAME notification duplicated the unlisted arcade start")
	}
}

func TestUnlistedCoreWithoutMRAIsScannedOnce(t *testing.T) {
	root := t.TempDir()
	db := &arcadeDB{}
	tr := arcadeTracker(root, db)
	tr.NameMap = nil
	tr.rbfName = func() string { return "SomeCore" }
	scans := 0
	tr.arcadeRoots = func() []string {
		scans++
		return []string{root}
	}
	tr.processCore("homebrew_thing")
	drainIndex(t, tr)
	tr.processCore(config.MenuCore)
	tr.processCore("homebrew_thing")
	drainIndex(t, tr)
	if scans != 1 {
		t.Fatalf("arcade folders scanned %d times for one unknown core", scans)
	}
	if tr.ActiveGame != "" || !tr.unknownSets["homebrew_thing"] {
		t.Fatalf("unknown core: game=%q unknown=%v", tr.ActiveGame, tr.unknownSets)
	}

	// An index that later finds the set forgets it was unknown.
	path := filepath.Join(root, "Homebrew Thing.mra")
	writeArcade(t, path, "homebrew_thing")
	tr.processCore(config.MenuCore)
	tr.mu.Lock()
	tr.requestArcadeIndex()
	tr.mu.Unlock()
	drainIndex(t, tr)
	if tr.unknownSets["homebrew_thing"] {
		t.Fatal("a set the new index found is still unknown")
	}
	tr.processCore("homebrew_thing")
	if tr.ActiveSystem != ArcadeSystem || tr.ActiveGamePath != path {
		t.Fatalf("set found later not tracked: system=%q path=%q", tr.ActiveSystem, tr.ActiveGamePath)
	}
}

func TestUnlistedCoreNamedLikeItsFileIsNotScanned(t *testing.T) {
	db := &arcadeDB{}
	tr := arcadeTracker(t.TempDir(), db)
	tr.NameMap = nil
	tr.rbfName = func() string { return "MyCore" }
	tr.processCore("MyCore")
	if tr.ActiveGame != "" || len(db.events) != 1 {
		t.Fatalf("unexpected state: %+v", db.events)
	}
	if n := pendingBuilds(t, tr); n != 0 {
		t.Fatalf("a core running from its own file asked for %d arcade index builds", n)
	}
}

func TestKnownCoreWithoutGameIsNotScanned(t *testing.T) {
	db := &arcadeDB{}
	tr := arcadeTracker(t.TempDir(), db)
	tr.NameMap = []NameMapping{{CoreName: "RA_NES", System: "NES", Name: "NES"}}
	tr.rbfName = func() string { return "NES" }
	tr.processCore("RA_NES")
	if tr.ActiveGame != "" {
		t.Fatalf("unexpected game: %q", tr.ActiveGame)
	}
	if n := pendingBuilds(t, tr); n != 0 {
		t.Fatalf("a core the name map knows asked for %d arcade index builds", n)
	}
}

func TestArcadeCoreReplacedMeanwhileIsNotPublished(t *testing.T) {
	db := &arcadeDB{}
	tr := arcadeTracker(t.TempDir(), db)
	published := false
	tr.setActiveGame = func(value string) error {
		if value != "" {
			published = true
		}
		return nil
	}
	tr.currentCore = func() string { return "MegaDrive" }
	tr.processCore("pooyan")
	if published || tr.ActiveGame != "" {
		t.Fatalf("a core already replaced published its set: game=%q", tr.ActiveGame)
	}
}

func TestUnlistedSetWhileIndexPending(t *testing.T) {
	root := t.TempDir()
	path := filepath.Join(root, "Bomb Jack Dualbit.mra")
	writeArcade(t, path, "bombjack_dualbit")
	pending := func() (*Tracker, *arcadeDB, *[]string) {
		db := &arcadeDB{}
		tr := arcadeTracker(root, db)
		tr.NameMap = nil
		tr.rbfName = func() string { return "BombJack" }
		var published []string
		tr.setActiveGame = func(value string) error { published = append(published, value); return nil }
		tr.processCore("bombjack_dualbit")
		if tr.ActiveSystem != "" || tr.ActiveGame != "" || len(db.events) != 1 {
			t.Fatalf("pending set not tracked as its core only: system=%q game=%q events=%+v",
				tr.ActiveSystem, tr.ActiveGame, db.events)
		}
		if tr.unknownSets["bombjack_dualbit"] {
			t.Fatal("a set was taken for unknown before the index was ready")
		}
		return tr, db, &published
	}

	tr, db, published := pending()
	drainIndex(t, tr)
	if tr.ActiveSystem != ArcadeSystem || tr.ActiveGame != "bombjack_dualbit" ||
		tr.ActiveGamePath != path || tr.ActiveGameName != "Bomb Jack Dualbit" {
		t.Fatalf("pending set not taken up: system=%q game=%q path=%q name=%q",
			tr.ActiveSystem, tr.ActiveGame, tr.ActiveGamePath, tr.ActiveGameName)
	}
	if len(db.events) != 2 || db.events[1].Action != EventActionGameStart || db.events[1].TargetPath != path ||
		db.events[1].ActiveCore.System != ArcadeSystem {
		t.Fatalf("expected one late game start, got %+v", db.events)
	}
	if len(*published) != 0 {
		t.Fatalf("a late take-up wrote ACTIVEGAME: %q", *published)
	}

	// MiSTer moved on before the index was ready.
	tr, db, published = pending()
	tr.currentCore = func() string { return "MegaDrive" }
	drainIndex(t, tr)
	if tr.ActiveSystem != "" || tr.ActiveGame != "" || len(db.events) != 1 || len(*published) != 0 {
		t.Fatalf("a replaced core was taken up: system=%q game=%q events=%d published=%q",
			tr.ActiveSystem, tr.ActiveGame, len(db.events), *published)
	}

	// A game already tracked for the core, as from MiSTer's recent files.
	tr, db, _ = pending()
	tr.processGame(path)
	game, events := tr.ActiveGame, len(db.events)
	if game == "" {
		t.Fatal("fixture: the recorded .mra was not tracked")
	}
	drainIndex(t, tr)
	if tr.ActiveGame != game || len(db.events) != events || tr.pendingSet != "" {
		t.Fatalf("take-up replaced a tracked game: game=%q, was %q; events=%d, were %d",
			tr.ActiveGame, game, len(db.events), events)
	}

	// The set name in ACTIVEGAME, stale or written by another tracker,
	// is not a file: the install still takes the set up.
	tr, _, _ = pending()
	tr.processGame("bombjack_dualbit")
	if tr.ActiveGame != "" {
		t.Fatalf("a bare set name became a game: %q", tr.ActiveGame)
	}
	drainIndex(t, tr)
	if tr.ActiveGame != "bombjack_dualbit" || tr.ActiveGamePath != path {
		t.Fatalf("pending set not taken up after its name: game=%q path=%q", tr.ActiveGame, tr.ActiveGamePath)
	}
}

func TestSetNameIsNotTakenForAConsoleGame(t *testing.T) {
	root := t.TempDir()
	db := &arcadeDB{}
	tr := arcadeTracker(root, db)
	tr.Config.Systems.GamesFolder = []string{root}
	tr.NameMap = []NameMapping{{CoreName: "MegaDrive", System: "Genesis", Name: "Genesis"}}
	tr.processCore("MegaDrive")
	tr.processGame("bombjack_dualbit")
	if tr.ActiveGame != "" || tr.ActiveGamePath != "" {
		t.Fatalf("an arcade set name became a game: %q %q", tr.ActiveGame, tr.ActiveGamePath)
	}
}

// With recents enabled, MiSTer's menu records a launched .mra in
// cores_recent.cfg and the tracker publishes that path to ACTIVEGAME, so
// the game can arrive before or after its core change. Either way an
// unlisted set the index resolves is tracked under its set name.
func TestUnlistedSetWithItsRecordedMRA(t *testing.T) {
	root := t.TempDir()
	path := filepath.Join(root, "Bomb Jack Dualbit.mra")
	writeArcade(t, path, "bombjack_dualbit")
	ready := func() (*Tracker, *arcadeDB, *[]string) {
		db := &arcadeDB{}
		tr := arcadeTracker(root, db)
		tr.NameMap = nil
		tr.rbfName = func() string { return "BombJack" }
		var published []string
		tr.setActiveGame = func(value string) error { published = append(published, value); return nil }
		readyIndex(t, tr)
		return tr, db, &published
	}
	actions := func(events []EventAction) []string {
		got := make([]string, 0, len(events))
		for i := range events {
			got = append(got, fmt.Sprintf("%d:%s", events[i].Action, events[i].Target))
		}
		return got
	}

	// The recorded .mra first, under the menu.
	tr, db, published := ready()
	tr.processGame(path)
	recorded := tr.ActiveGame
	if recorded == "" || recorded == "bombjack_dualbit" {
		t.Fatalf("fixture: the recorded .mra was tracked as %q", recorded)
	}
	tr.processCore("bombjack_dualbit")
	tr.processGame("bombjack_dualbit") // the tracker's own ACTIVEGAME echo
	want := []string{
		fmt.Sprintf("%d:%s", EventActionGameStart, recorded),
		fmt.Sprintf("%d:%s", EventActionCoreStart, "bombjack_dualbit"),
		fmt.Sprintf("%d:%s", EventActionGameStop, recorded),
		fmt.Sprintf("%d:%s", EventActionGameStart, "bombjack_dualbit"),
	}
	if got := actions(db.events); !reflect.DeepEqual(got, want) {
		t.Fatalf("events = %q, want %q", got, want)
	}
	if !reflect.DeepEqual(*published, []string{"bombjack_dualbit"}) {
		t.Fatalf("ACTIVEGAME writes = %q, want the set name", *published)
	}
	if tr.ActiveGame != "bombjack_dualbit" || tr.ActiveGamePath != path || tr.ActiveGameName != "Bomb Jack Dualbit" {
		t.Fatalf("game=%q path=%q name=%q", tr.ActiveGame, tr.ActiveGamePath, tr.ActiveGameName)
	}

	// The core change first.
	tr, db, published = ready()
	tr.processCore("bombjack_dualbit")
	tr.processGame(path)
	want = []string{
		fmt.Sprintf("%d:%s", EventActionCoreStart, "bombjack_dualbit"),
		fmt.Sprintf("%d:%s", EventActionGameStart, "bombjack_dualbit"),
	}
	if got := actions(db.events); !reflect.DeepEqual(got, want) {
		t.Fatalf("events = %q, want %q", got, want)
	}
	if !reflect.DeepEqual(*published, []string{"bombjack_dualbit"}) {
		t.Fatalf("ACTIVEGAME writes = %q, want the set name", *published)
	}
	if tr.ActiveGame != "bombjack_dualbit" || tr.ActiveGamePath != path {
		t.Fatalf("game=%q path=%q", tr.ActiveGame, tr.ActiveGamePath)
	}
}
