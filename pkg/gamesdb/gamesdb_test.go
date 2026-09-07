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

package gamesdb

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/wizzomafizzo/mrext/pkg/config"
	"github.com/wizzomafizzo/mrext/pkg/games"
	bolt "go.etcd.io/bbolt"
)

func indexFixture(t *testing.T) (string, []games.System, []games.PathResult) {
	t.Helper()
	root := t.TempDir()
	db := filepath.Join(root, "games.db")
	systems := make([]games.System, 0, 2)
	paths := make([]games.PathResult, 0, 2)
	for _, id := range []string{"NES", "SNES"} {
		system, lookupErr := games.GetSystem(id)
		if lookupErr != nil {
			t.Fatal(lookupErr)
		}
		folder := filepath.Join(root, id)
		if mkdirErr := os.Mkdir(folder, 0o750); mkdirErr != nil {
			t.Fatal(mkdirErr)
		}
		ext := ".nes"
		if id == "SNES" {
			ext = ".sfc"
		}
		for _, name := range []string{"deleted", "renamed"} {
			if writeErr := os.WriteFile(filepath.Join(folder, name+ext), nil, 0o600); writeErr != nil {
				t.Fatal(writeErr)
			}
		}
		systems = append(systems, *system)
		paths = append(paths, games.PathResult{System: *system, Path: folder})
	}
	if _, err := indexNames(db, systems, paths, games.WalkFiles, nil, true); err != nil {
		t.Fatal(err)
	}
	return db, systems, paths
}

func namesSnapshot(t *testing.T, path string) map[string]string {
	t.Helper()
	db, err := openAt(path, &bolt.Options{ReadOnly: true})
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = db.Close() }()
	result := make(map[string]string)
	if err := db.View(func(tx *bolt.Tx) error {
		return tx.Bucket([]byte(BucketNames)).ForEach(func(key, value []byte) error {
			result[string(key)] = string(value)
			return nil
		})
	}); err != nil {
		t.Fatal(err)
	}
	return result
}

func TestRebuildRemovesDeletedRenamedAndMissingSystems(t *testing.T) {
	db, systems, paths := indexFixture(t)
	if err := os.Remove(filepath.Join(paths[0].Path, "deleted.nes")); err != nil {
		t.Fatal(err)
	}
	oldName, newName := filepath.Join(paths[0].Path, "renamed.nes"), filepath.Join(paths[0].Path, "new.nes")
	if err := os.Rename(oldName, newName); err != nil {
		t.Fatal(err)
	}
	var progress []IndexStatus
	count, err := indexNames(db, systems, paths[:1], games.WalkFiles, func(status IndexStatus) {
		progress = append(progress, status)
	}, true)
	if err != nil || count != 1 {
		t.Fatalf("rebuild = %d, %v", count, err)
	}
	want := map[string]string{indexedSystemsKey: "NES", "NES:new": filepath.Join(paths[0].Path, "new.nes")}
	if got := namesSnapshot(t, db); !reflect.DeepEqual(got, want) {
		t.Fatalf("stale names or metadata: %v", got)
	}
	last := progress[len(progress)-1]
	if last.Step != last.Total || last.SystemID != "" || last.Files != 1 {
		t.Fatalf("final progress = %+v", last)
	}
	for _, status := range progress {
		if status.Step > status.Total {
			t.Fatalf("progress overflow: %+v", status)
		}
	}
}

func TestPartialIndexPreservesOtherSystems(t *testing.T) {
	db, systems, paths := indexFixture(t)
	before := namesSnapshot(t, db)
	if _, err := indexNames(db, systems[:1], nil, games.WalkFiles, nil, false); err != nil {
		t.Fatal(err)
	}
	want := map[string]string{
		indexedSystemsKey: "SNES", "SNES:deleted": before["SNES:deleted"], "SNES:renamed": before["SNES:renamed"],
	}
	if got := namesSnapshot(t, db); !reflect.DeepEqual(got, want) {
		t.Fatalf("partial refresh damaged unrelated names or retained missing system: %v", got)
	}
	if _, err := indexNames(db, systems[:1], paths[:1], games.WalkFiles, nil, false); err != nil {
		t.Fatal(err)
	}
	if got := namesSnapshot(t, db); !reflect.DeepEqual(got, before) {
		t.Fatalf("partial refresh failed to restore selected system: %v", got)
	}
}

func TestFailedRebuildRollsBackNamesAndMetadata(t *testing.T) {
	for _, replaceAll := range []bool{false, true} {
		db, systems, paths := indexFixture(t)
		before := namesSnapshot(t, db)
		failure := errors.New("fixture scan failed")
		scan := func(id, path string, visit func(string) error) error {
			if id == "SNES" {
				return failure
			}
			return visit(filepath.Join(path, "new.nes"))
		}
		if _, err := indexNames(db, systems, paths, scan, nil, replaceAll); !errors.Is(err, failure) {
			t.Fatalf("scan failure was hidden: %v", err)
		}
		if got := namesSnapshot(t, db); !reflect.DeepEqual(got, before) {
			t.Fatalf("failed scan changed working index: %v", got)
		}
		writeFailure := func(_, path string, visit func(string) error) error {
			return visit(filepath.Join(path, strings.Repeat("x", bolt.MaxKeySize+1)+".nes"))
		}
		_, writeErr := indexNames(db, systems, paths, writeFailure, nil, replaceAll)
		if !errors.Is(writeErr, bolt.ErrKeyTooLarge) {
			t.Fatalf("write failure was hidden: %v", writeErr)
		}
		if got := namesSnapshot(t, db); !reflect.DeepEqual(got, before) {
			t.Fatalf("failed write changed working index: %v", got)
		}
	}
}

func TestRebuiltIndexSupportsReadOnlySearchAndMetadata(t *testing.T) {
	path, systems, _ := indexFixture(t)
	results, err := searchNamesAt(path, systems, "renamed", strings.Contains)
	if err != nil || len(results) != 2 {
		t.Fatalf("read-only search = %v, %v", results, err)
	}
	reader, err := openAt(path, &bolt.Options{ReadOnly: true})
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = reader.Close() }()
	indexed, err := readIndexedSystems(reader)
	if err != nil || !reflect.DeepEqual(indexed, []string{"NES", "SNES"}) {
		t.Fatalf("read-only metadata = %v, %v", indexed, err)
	}
}

func TestIndexDeduplicatesResolvedSystemPaths(t *testing.T) {
	db, systems, paths := indexFixture(t)
	alias := filepath.Join(filepath.Dir(db), "alias")
	if err := os.Symlink(paths[0].Path, alias); err != nil {
		t.Fatal(err)
	}
	repeated := make([]games.PathResult, 0, 20001)
	for range 20000 {
		repeated = append(repeated, paths[0])
	}
	repeated = append(repeated, games.PathResult{System: systems[0], Path: alias})
	calls := 0
	count, err := indexNames(db, systems[:1], repeated, func(_, path string, visit func(string) error) error {
		calls++
		for i := range 10000 {
			if err := visit(filepath.Join(path, fmt.Sprintf("game-%05d.nes", i))); err != nil {
				return err
			}
		}
		return nil
	}, nil, true)
	if err != nil || count != 10000 || calls != 1 {
		t.Fatalf("duplicate scan: count=%d calls=%d err=%v", count, calls, err)
	}
}

func TestIndexWriterBoundsPendingBatch(t *testing.T) {
	db, err := openAt(filepath.Join(t.TempDir(), "stage.db"), &bolt.Options{NoSync: true})
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = db.Close() }()
	writer := indexWriter{db: db}
	for i := range 2000 {
		if err := writer.put(fmt.Sprintf("NES:%d", i), strings.Repeat("x", 2000)); err != nil {
			t.Fatal(err)
		}
		if len(writer.batch) > indexBatchEntries || writer.bytes > indexBatchBytes {
			t.Fatalf("unbounded pending batch: entries=%d bytes=%d", len(writer.batch), writer.bytes)
		}
	}
	if err := writer.flush(); err != nil {
		t.Fatal(err)
	}
	if len(writer.batch) != 0 || writer.bytes != 0 {
		t.Fatal("flushed batch retained pending data")
	}
	if err := writer.put("NES:oversized", strings.Repeat("x", indexBatchBytes)); err == nil {
		t.Fatal("oversized entry exceeded batch bound")
	}
}

func TestPreviousIndexReadableDuringRebuild(t *testing.T) {
	db, systems, paths := indexFixture(t)
	before := namesSnapshot(t, db)
	_, err := indexNames(db, systems, paths, func(_, path string, visit func(string) error) error {
		if got := namesSnapshot(t, db); !reflect.DeepEqual(got, before) {
			t.Fatalf("live index changed before publication: %v", got)
		}
		return visit(filepath.Join(path, "new.nes"))
	}, nil, true)
	if err != nil {
		t.Fatal(err)
	}
}

func TestFailedInitialBuildLeavesNoIndexOrStagedFiles(t *testing.T) {
	root := t.TempDir()
	db := filepath.Join(root, "games.db")
	failure := errors.New("fixture failure")
	_, err := indexNames(db, nil, []games.PathResult{{System: games.System{Id: "NES"}, Path: root}},
		func(string, string, func(string) error) error { return failure }, nil, true)
	if !errors.Is(err, failure) {
		t.Fatalf("scan error = %v", err)
	}
	if _, statErr := os.Stat(db); !os.IsNotExist(statErr) {
		t.Fatalf("failed initial build left an index: %v", statErr)
	}
	matches, err := filepath.Glob(filepath.Join(root, ".games-index-*"))
	if err != nil || len(matches) != 0 {
		t.Fatalf("staged files leaked: %v, %v", matches, err)
	}
}

func TestRebuildHoldsStableWriterLock(t *testing.T) {
	db, systems, paths := indexFixture(t)
	ctx, release := context.WithCancel(context.Background())
	defer release()
	started := make(chan struct{})
	done := make(chan error, 1)
	go func() {
		_, err := indexNames(db, systems[:1], paths[:1], func(_, path string, visit func(string) error) error {
			close(started)
			<-ctx.Done()
			return visit(filepath.Join(path, "new.nes"))
		}, nil, true)
		done <- err
	}()
	select {
	case <-started:
	case <-time.After(5 * time.Second):
		t.Fatal("writer did not start")
	}
	lock, err := bolt.Open(db+config.GamesDBLockSuffix, 0o600, &bolt.Options{Timeout: time.Millisecond})
	if lock != nil {
		_ = lock.Close()
	}
	release()
	if !errors.Is(err, bolt.ErrTimeout) {
		t.Fatalf("concurrent writer was not excluded: %v", err)
	}
	select {
	case buildErr := <-done:
		if buildErr != nil {
			t.Fatal(buildErr)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("writer did not finish")
	}
	if _, statErr := os.Stat(db + config.GamesDBLockSuffix); statErr != nil {
		t.Fatalf("publication removed stable lock file: %v", statErr)
	}
}

func TestEmptyFullRebuildClearsIndex(t *testing.T) {
	db, systems, _ := indexFixture(t)
	if _, err := indexNames(db, systems, nil, games.WalkFiles, nil, true); err != nil {
		t.Fatal(err)
	}
	if got := namesSnapshot(t, db); !reflect.DeepEqual(got, map[string]string{indexedSystemsKey: ""}) {
		t.Fatalf("empty rebuild retained names: %v", got)
	}
	reader, err := openAt(db, &bolt.Options{ReadOnly: true})
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = reader.Close() }()
	indexed, err := readIndexedSystems(reader)
	if err != nil || len(indexed) != 0 {
		t.Fatalf("empty metadata = %v, %v", indexed, err)
	}
}

func TestSecondIndexerReportsBusyInsteadOfHanging(t *testing.T) {
	db, systems, paths := indexFixture(t)
	ctx, release := context.WithCancel(context.Background())
	defer release()
	started := make(chan struct{})
	done := make(chan error, 1)
	go func() {
		_, err := indexNames(db, systems[:1], paths[:1], func(_, path string, visit func(string) error) error {
			close(started)
			<-ctx.Done()
			return visit(filepath.Join(path, "new.nes"))
		}, nil, true)
		done <- err
	}()
	select {
	case <-started:
	case <-time.After(5 * time.Second):
		t.Fatal("first writer did not start")
	}

	previous := indexLockTimeout
	indexLockTimeout = 50 * time.Millisecond
	t.Cleanup(func() { indexLockTimeout = previous })

	// bbolt's zero default timeout waits forever, so this call used to block
	// before its first progress callback: Remote's rebuild and search.sh at the
	// same time left one of them on a frozen screen with nothing to explain it.
	progressed := false
	_, err := indexNames(db, systems[:1], paths[:1],
		func(_, path string, visit func(string) error) error { return visit(filepath.Join(path, "other.nes")) },
		func(IndexStatus) { progressed = true }, true)
	if !errors.Is(err, ErrIndexBusy) {
		t.Fatalf("second indexer error = %v, want ErrIndexBusy", err)
	}
	if progressed {
		t.Error("second indexer reported progress it never made")
	}

	release()
	select {
	case buildErr := <-done:
		if buildErr != nil {
			t.Fatal(buildErr)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("first writer did not finish")
	}
}

func TestDanglingSymlinksDoNotAbortTheIndex(t *testing.T) {
	db, systems, paths := indexFixture(t)

	// One broken link in a library used to abort the whole run with
	// "resolve game symlink", so nothing was indexed at all, including every
	// system that would have scanned cleanly.
	broken := filepath.Join(paths[0].Path, "unplugged.nes")
	if err := os.Symlink(filepath.Join(t.TempDir(), "never-existed.nes"), broken); err != nil {
		t.Fatal(err)
	}

	files, err := indexNames(db, systems, paths, games.WalkFiles, nil, true)
	if err != nil {
		t.Fatalf("a dangling symlink aborted the index: %v", err)
	}
	names := namesSnapshot(t, db)
	if len(names) == 0 {
		t.Fatal("nothing was indexed")
	}
	for _, name := range names {
		if strings.Contains(name, "unplugged") {
			t.Errorf("a broken link was indexed as a game: %s", name)
		}
	}
	if files == 0 {
		t.Fatal("no files counted")
	}
}

func TestUnresolvableRootsAreSkippedAndCounted(t *testing.T) {
	root := t.TempDir()
	present := filepath.Join(root, "NES")
	if err := os.Mkdir(present, 0o750); err != nil {
		t.Fatal(err)
	}
	system, err := games.GetSystem("NES")
	if err != nil {
		t.Fatal(err)
	}
	paths := []games.PathResult{
		{System: *system, Path: present},
		{System: *system, Path: filepath.Join(root, "not-mounted")},
	}

	unique, skipped := uniquePaths(paths)
	if skipped != 1 {
		t.Errorf("skipped = %d, want 1", skipped)
	}
	if got := len(unique["NES"]); got != 1 {
		t.Errorf("kept %d roots, want the one that resolves", got)
	}
}
