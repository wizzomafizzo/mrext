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
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/wizzomafizzo/mrext/pkg/config"
	"github.com/wizzomafizzo/mrext/pkg/games"
	"github.com/wizzomafizzo/mrext/pkg/utils"
	bolt "go.etcd.io/bbolt"
)

const (
	indexBatchEntries = 256
	indexBatchBytes   = 256 * 1024
)

// NewNamesIndex refreshes requested systems without discarding unrelated
// systems needed by Search or Remote when LaunchSync indexes a playlist.
func NewNamesIndex(cfg *config.UserConfig, systems []games.System, update func(IndexStatus)) (int, error) {
	return generateNamesIndex(cfg, systems, update, false)
}

// RebuildNamesIndex replaces all names and system metadata after a successful
// scan. Failed regeneration leaves the previous index available to readers.
func RebuildNamesIndex(cfg *config.UserConfig, systems []games.System, update func(IndexStatus)) (int, error) {
	return generateNamesIndex(cfg, systems, update, true)
}

func generateNamesIndex(
	cfg *config.UserConfig, systems []games.System, update func(IndexStatus), replaceAll bool,
) (int, error) {
	unique := make([]games.System, 0)
	seen := make(map[string]bool)
	for i := range systems {
		system := &systems[i]
		if !seen[system.Id] {
			seen[system.Id] = true
			unique = append(unique, *system)
		}
	}
	if update != nil {
		update(IndexStatus{Step: 1, Total: len(unique) + 2})
	}
	return indexNames(config.GamesDB, unique, games.GetSystemPaths(cfg, unique), games.WalkFiles, update, replaceAll)
}

// indexLockTimeout bounds the wait for another indexer to finish. Long enough
// to ride out a handover, short enough to tell the user what is happening.
// A variable so tests need not wait it out.
var indexLockTimeout = 5 * time.Second

// ErrIndexBusy reports that another process is already writing the index.
var ErrIndexBusy = errors.New("another indexer is already running; wait for it to finish and try again")

// indexNames stages bounded batches beside the live database, never in /tmp.
// A separate persistent lock serializes writers across atomic file replacement;
// readers keep using the previous inode until they close their connection.
func indexNames(
	filename string, systems []games.System, paths []games.PathResult,
	scan func(string, string, func(string) error) error, update func(IndexStatus), replaceAll bool,
) (int, error) {
	if update == nil {
		update = func(IndexStatus) {}
	}
	if err := os.MkdirAll(filepath.Dir(filename), 0o750); err != nil {
		return 0, fmt.Errorf("create index directory: %w", err)
	}
	// bbolt's default flock timeout is zero, which means wait forever. Two
	// indexers, such as Remote's rebuild button and search.sh, would leave the
	// second one blocked here before its first progress callback, so its UI
	// sat on the opening message with nothing to show and no way to know why.
	lock, err := bolt.Open(filename+config.GamesDBLockSuffix, 0o600, &bolt.Options{Timeout: indexLockTimeout})
	if err != nil {
		if errors.Is(err, bolt.ErrTimeout) {
			return 0, ErrIndexBusy
		}
		return 0, fmt.Errorf("lock index writer: %w", err)
	}
	defer func() { _ = lock.Close() }()
	systemPaths, skipped := uniquePaths(paths)
	status := IndexStatus{Total: len(systemPaths) + 2, Step: 1, Skipped: skipped}
	update(status)
	temporary, err := os.CreateTemp(filepath.Dir(filename), ".games-index-*")
	if err != nil {
		return 0, fmt.Errorf("create staged index: %w", err)
	}
	stagedPath := temporary.Name()
	defer func() { _ = os.Remove(stagedPath) }()
	if closeErr := temporary.Close(); closeErr != nil {
		return 0, fmt.Errorf("close staged index file: %w", closeErr)
	}
	db, err := openAt(stagedPath, &bolt.Options{NoSync: true, NoFreelistSync: true})
	if err != nil {
		return 0, err
	}
	defer func() { _ = db.Close() }()
	writer := indexWriter{db: db}
	indexed := make(map[string]bool)
	if !replaceAll {
		if err := copyUnselected(filename, systems, &writer, indexed); err != nil {
			return 0, err
		}
	}
	for _, id := range utils.AlphaMapKeys(systemPaths) {
		status.SystemID = id
		status.Step++
		update(status)
		for _, path := range systemPaths[id] {
			if err := scan(id, path, func(file string) error {
				base := filepath.Base(file)
				name := strings.TrimSuffix(base, filepath.Ext(base))
				if err := writer.put(NameKey(id, name), file); err != nil {
					return err
				}
				status.Files++
				return nil
			}); err != nil {
				return status.Files, fmt.Errorf("scan %s: %w", path, err)
			}
		}
		indexed[id] = true
	}
	status.Step, status.SystemID = status.Total, ""
	update(status)
	if err := writer.put(indexedSystemsKey, strings.Join(utils.AlphaMapKeys(indexed), ",")); err != nil {
		return status.Files, err
	}
	if err := writer.flush(); err != nil {
		return status.Files, err
	}
	if err := db.Sync(); err != nil {
		return status.Files, fmt.Errorf("sync staged index: %w", err)
	}
	if err := db.Close(); err != nil {
		return status.Files, fmt.Errorf("close staged index: %w", err)
	}
	if err := os.Rename(stagedPath, filename); err != nil {
		return status.Files, fmt.Errorf("publish staged index: %w", err)
	}
	return status.Files, nil
}

// uniquePaths deduplicates game roots by their resolved location. A root that
// cannot be resolved, such as a symlink into a drive that is not attached, is
// skipped and counted: one bad symlink used to abort the run before anything
// was indexed, including the systems that would have scanned cleanly.
func uniquePaths(paths []games.PathResult) (roots map[string][]string, skippedRoots int) {
	roots = make(map[string][]string)
	seen := make(map[string]bool)
	for i := range paths {
		path := &paths[i]
		resolved, err := filepath.EvalSymlinks(path.Path)
		if err != nil {
			skippedRoots++
			continue
		}
		key := path.System.Id + ":" + resolved
		if !seen[key] {
			seen[key] = true
			roots[path.System.Id] = append(roots[path.System.Id], path.Path)
		}
	}
	return roots, skippedRoots
}

func copyUnselected(filename string, systems []games.System, writer *indexWriter, indexed map[string]bool) error {
	db, err := openAt(filename, &bolt.Options{ReadOnly: true})
	if errors.Is(err, os.ErrNotExist) {
		return nil
	}
	if err != nil {
		return fmt.Errorf("read previous index: %w", err)
	}
	defer func() { _ = db.Close() }()
	selected := make(map[string]bool, len(systems))
	for i := range systems {
		selected[systems[i].Id] = true
	}
	err = db.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(BucketNames))
		for _, id := range strings.Split(string(bucket.Get([]byte(indexedSystemsKey))), ",") {
			if id != "" && !selected[id] {
				indexed[id] = true
			}
		}
		return bucket.ForEach(func(key, value []byte) error {
			name := string(key)
			id, _, _ := strings.Cut(name, ":")
			if name == indexedSystemsKey || selected[id] {
				return nil
			}
			return writer.put(name, string(value))
		})
	})
	if err != nil {
		return fmt.Errorf("copy retained index entries: %w", err)
	}
	return nil
}

type indexEntry struct{ key, value string }

type indexWriter struct {
	db    *bolt.DB
	batch []indexEntry
	bytes int
}

func (w *indexWriter) put(key, value string) error {
	if len(key) > bolt.MaxKeySize {
		return fmt.Errorf("index key: %w", bolt.ErrKeyTooLarge)
	}
	size := len(key) + len(value)
	if size > indexBatchBytes {
		return fmt.Errorf("index entry exceeds batch limit: %d bytes", size)
	}
	if len(w.batch) >= indexBatchEntries || w.bytes+size > indexBatchBytes {
		if err := w.flush(); err != nil {
			return err
		}
	}
	w.batch = append(w.batch, indexEntry{key: key, value: value})
	w.bytes += size
	return nil
}

func (w *indexWriter) flush() error {
	if len(w.batch) == 0 {
		return nil
	}
	if err := w.db.Update(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(BucketNames))
		for _, entry := range w.batch {
			if err := bucket.Put([]byte(entry.key), []byte(entry.value)); err != nil {
				return fmt.Errorf("index game name: %w", err)
			}
		}
		return nil
	}); err != nil {
		return fmt.Errorf("write index batch: %w", err)
	}
	clear(w.batch)
	w.batch, w.bytes = w.batch[:0], 0
	return nil
}
