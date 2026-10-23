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
	"reflect"
	"sync"
	"testing"
	"time"

	"github.com/fsnotify/fsnotify"
)

// countStarts counts the game-start events for target.
func countStarts(events []EventAction, target string) int {
	n := 0
	for i := range events {
		if events[i].Action == EventActionGameStart && events[i].Target == target {
			n++
		}
	}
	return n
}

// lockedDB is arcadeDB for a tracker that also runs a real index build.
type lockedDB struct {
	arcadeDB
	mu sync.Mutex
}

func (d *lockedDB) AddEvent(event *EventAction) error {
	d.mu.Lock()
	defer d.mu.Unlock()
	return d.arcadeDB.AddEvent(event)
}

func (d *lockedDB) snapshot() []EventAction {
	d.mu.Lock()
	defer d.mu.Unlock()
	return append([]EventAction(nil), d.events...)
}

// The tracker must keep handling core and game changes while the arcade
// folders are walked, and fill the arcade path in once the index is ready.
func TestIndexBuildRunsOffTheLock(t *testing.T) {
	root := t.TempDir()
	path := filepath.Join(root, "Pooyan.mra")
	writeArcade(t, path, "pooyan")
	db := &lockedDB{}
	tr := arcadeTracker(root, &db.arcadeDB)
	tr.Db = db
	tr.runBuild = nil // a real goroutine
	release := make(chan struct{})
	walking := make(chan struct{})
	tr.arcadeRoots = func() []string {
		close(walking)
		<-release
		return []string{root}
	}
	t.Cleanup(func() {
		select {
		case <-release:
		default:
			close(release)
		}
		tr.WaitArcadeIndex(5 * time.Second)
	})

	within := func(what string, fn func()) {
		t.Helper()
		done := make(chan struct{})
		go func() {
			fn()
			close(done)
		}()
		select {
		case <-done:
		case <-time.After(2 * time.Second):
			t.Fatalf("%s waited for the arcade index build", what)
		}
	}

	within("core change", func() {
		tr.mu.Lock()
		defer tr.mu.Unlock()
		tr.processCore("pooyan")
	})
	<-walking
	events := db.snapshot()
	if len(events) != 2 || events[0].Action != EventActionCoreStart || events[1].Action != EventActionGameStart ||
		events[1].TargetPath != "" {
		t.Fatalf("expected core and game start without a path yet, got %+v", events)
	}
	within("tick", func() {
		tr.tick(0)
		tr.tick(0)
	})
	within("ACTIVEGAME echo", func() {
		tr.mu.Lock()
		defer tr.mu.Unlock()
		tr.processGame("pooyan")
	})
	within("session read", func() { tr.SessionTimes() })
	within("playing read", func() {
		if now := tr.Playing(); now.Core != "pooyan" || now.System != ArcadeSystem || now.Game != "pooyan" {
			t.Errorf("playing during the build = %+v", now)
		}
	})

	close(release)
	if !tr.WaitArcadeIndex(5 * time.Second) {
		t.Fatal("arcade index never installed")
	}
	tr.mu.Lock()
	gamePath, recorded, total := tr.ActiveGamePath, tr.GameTimes["pooyan"].Path, tr.GameTimes["pooyan"].Time
	tr.mu.Unlock()
	if gamePath != path || recorded != path {
		t.Fatalf("path not filled in after the build: active=%q recorded=%q", gamePath, recorded)
	}
	if n := countStarts(db.snapshot(), "pooyan"); n != 1 {
		t.Fatalf("game started %d times, want 1", n)
	}
	if _, game := tr.SessionTimes(); game != 2 || total != 2 {
		t.Fatalf("session time reset by the path fill: session=%d total=%d", game, total)
	}
	if game, _ := tr.ActiveSession(); game.Path != path {
		t.Fatalf("running game path = %q", game.Path)
	}
	if now := tr.Playing(); now.GamePath != path || now.Game != "pooyan" {
		t.Fatalf("playing after the build = %+v", now)
	}
}

func TestReconcileSkipsAReplacedCore(t *testing.T) {
	root := t.TempDir()
	writeArcade(t, filepath.Join(root, "Pooyan.mra"), "pooyan")
	db := &arcadeDB{}
	tr := arcadeTracker(root, db)
	tr.NameMap = append(tr.NameMap, NameMapping{CoreName: "NES", System: "NES", Name: "NES"})
	tr.processCore("pooyan")
	tr.processCore("NES")
	events := len(db.events)
	state := func() [4]string {
		return [4]string{tr.ActiveCore, tr.ActiveSystem, tr.ActiveGame, tr.ActiveGamePath}
	}
	before := state()
	published := false
	tr.setActiveGame = func(string) error { published = true; return nil }
	drainIndex(t, tr)
	if after := state(); after != before || tr.ActiveCore != "NES" || tr.ActiveGame != "" {
		t.Fatalf("index install changed the NES state: %q, was %q", after, before)
	}
	if len(db.events) != events || published {
		t.Fatalf("index install emitted %d events, published=%t", len(db.events)-events, published)
	}
}

func TestIndexSkipsLinksIntoRoots(t *testing.T) {
	dir, links, outside := canonicalTempDir(t), canonicalTempDir(t), canonicalTempDir(t)
	// The root is reached through a link, as /var is on macOS.
	root := filepath.Join(links, "arcade")
	if err := os.Symlink(dir, root); err != nil {
		t.Fatal(err)
	}
	timber := filepath.Join(dir, "Timber.mra")
	writeArcade(t, timber, "timber")
	writeArcade(t, filepath.Join(outside, "Other.mra"), "other")
	writeArcade(t, filepath.Join(dir, "._Timber.mra"), "appledouble")
	organized := filepath.Join(dir, "_Organized", "T")
	if err := os.MkdirAll(organized, 0o755); err != nil {
		t.Fatal(err)
	}
	inRoot := map[string]string{
		filepath.Join(organized, "Relative.mra"):  filepath.Join("..", "..", "Timber.mra"),
		filepath.Join(organized, "Canonical.mra"): timber,
		filepath.Join(organized, "Raw.mra"):       filepath.Join(root, "Timber.mra"),
		filepath.Join(organized, "Folder"):        root,
	}
	for link, target := range inRoot {
		if err := os.Symlink(target, link); err != nil {
			t.Fatal(err)
		}
	}
	if err := os.Symlink(filepath.Join(outside, "Other.mra"), filepath.Join(dir, "Other.mra")); err != nil {
		t.Fatal(err)
	}
	// A link that leaves the roots and loops back on itself must end.
	if err := os.Symlink(outside, filepath.Join(outside, "loop")); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(outside, filepath.Join(dir, "Outside")); err != nil {
		t.Fatal(err)
	}

	var resolved []string
	original := evalSymlinks
	evalSymlinks = func(path string) (string, error) {
		resolved = append(resolved, path)
		return original(path)
	}
	t.Cleanup(func() { evalSymlinks = original })

	idx := buildArcadeIndex([]string{root})
	for _, path := range resolved {
		rel, err := filepath.Rel(root, path)
		if err != nil {
			t.Fatal(err)
		}
		if _, ok := inRoot[filepath.Join(dir, rel)]; ok {
			t.Errorf("a link into the roots was resolved: %s", path)
		}
	}
	if got, want := idx.paths("timber"), []string{filepath.Join(root, "Timber.mra")}; !reflect.DeepEqual(got, want) {
		t.Errorf("timber = %v, want %v", got, want)
	}
	if got, want := idx.paths("other"), []string{filepath.Join(root, "Other.mra")}; !reflect.DeepEqual(got, want) {
		t.Errorf("other = %v, want %v (the Outside folder holds the same file)", got, want)
	}
	if got := idx.paths("appledouble"); got != nil {
		t.Errorf("AppleDouble file indexed: %v", got)
	}
}

// canonicalTempDir is t.TempDir with any link in its path resolved.
func canonicalTempDir(t *testing.T) string {
	t.Helper()
	dir, err := filepath.EvalSymlinks(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	return dir
}

func TestAmbiguousAndMissingSetsDoNotLoop(t *testing.T) {
	ambiguous := t.TempDir()
	writeArcade(t, filepath.Join(ambiguous, "Pooyan.mra"), "pooyan")
	writeArcade(t, filepath.Join(ambiguous, "Pooyan (alt).mra"), "pooyan")
	for name, root := range map[string]string{"ambiguous": ambiguous, "missing": t.TempDir()} {
		db := &arcadeDB{}
		tr := arcadeTracker(root, db)
		tr.rebuildInterval = time.Hour
		readyIndex(t, tr)
		tr.processCore("pooyan")
		for range 10 {
			tr.processGame("pooyan")
		}
		if n := pendingBuilds(t, tr); n != 0 {
			t.Errorf("%s set: %d index builds within the rebuild interval", name, n)
		}
		if n := countStarts(db.events, "pooyan"); n != 1 || tr.ActiveGamePath != "" {
			t.Errorf("%s set: %d game starts, path %q", name, n, tr.ActiveGamePath)
		}
		// Once the interval has passed, the echoes ask for one build between
		// them, not one each.
		tr.lastBuildStart = time.Time{}
		for range 10 {
			tr.processGame("pooyan")
		}
		if n := pendingBuilds(t, tr); n != 1 {
			t.Errorf("%s set: %d index builds after the interval, want 1", name, n)
		}
	}
}

// A menu change arrives on the watcher goroutine while the index install
// updates the running game on the build goroutine; run with -race.
func TestMenuEventDuringIndexInstall(t *testing.T) {
	dir := t.TempDir()
	full, current := filepath.Join(dir, "FULLPATH"), filepath.Join(dir, "CURRENTPATH")
	for _, file := range []string{full, current} {
		if err := os.WriteFile(file, []byte("games"), 0o600); err != nil {
			t.Fatal(err)
		}
	}
	oldFull, oldCurrent := fullPathFile, currentPathFile
	fullPathFile, currentPathFile = full, current
	t.Cleanup(func() { fullPathFile, currentPathFile = oldFull, oldCurrent })

	root := t.TempDir()
	path := filepath.Join(root, "Pooyan.mra")
	writeArcade(t, path, "pooyan")
	for range 5 {
		tr := arcadeTracker(root, &arcadeDB{})
		tr.processCore("pooyan")
		watch := &fileWatch{logger: silentLogger{}, files: map[string]func(){current: tr.trackMenu}}
		done := make(chan struct{})
		go func() {
			defer close(done)
			for range 100 {
				watch.dispatch(fsnotify.Event{Name: current, Op: fsnotify.Write})
			}
		}()
		drainIndex(t, tr)
		<-done
		if tr.ActiveGamePath != path {
			t.Fatalf("path = %q, want %q", tr.ActiveGamePath, path)
		}
	}
}
