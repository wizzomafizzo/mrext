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
	"bytes"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/wizzomafizzo/mrext/pkg/config"
	"github.com/wizzomafizzo/mrext/pkg/games"
	"github.com/wizzomafizzo/mrext/pkg/utils"
	bolt "go.etcd.io/bbolt"
	"golang.org/x/sync/errgroup"
)

const (
	BucketNames       = "names"
	indexedSystemsKey = "meta:indexedSystems"
)

// Return the key for a name in the names index.
func NameKey(systemID, name string) string {
	return systemID + ":" + name
}

// Check if the gamesdb exists on disk.
func DBExists() bool {
	_, err := os.Stat(config.GamesDB)
	return err == nil
}

// Open the gamesdb with the given options. If the database does not exist it
// will be created and the buckets will be initialized.
func open(options *bolt.Options) (*bolt.DB, error) {
	if err := os.MkdirAll(filepath.Dir(config.GamesDB), 0o750); err != nil {
		return nil, fmt.Errorf("create games database directory: %w", err)
	}

	db, err := bolt.Open(config.GamesDB, 0o600, options)
	if err != nil {
		return nil, fmt.Errorf("open games database: %w", err)
	}

	if err := db.Update(func(txn *bolt.Tx) error {
		for _, bucket := range []string{BucketNames} {
			if _, err := txn.CreateBucketIfNotExists([]byte(bucket)); err != nil {
				return fmt.Errorf("create %s bucket: %w", bucket, err)
			}
		}
		return nil
	}); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("initialize games database: %w", err)
	}

	return db, nil
}

// Open the gamesdb with default options for generating names index.
func openNames() (*bolt.DB, error) {
	return open(&bolt.Options{
		NoSync:         true,
		NoFreelistSync: true,
	})
}

func readIndexedSystems(db *bolt.DB) ([]string, error) {
	var systems []string

	err := db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(BucketNames))
		v := b.Get([]byte(indexedSystemsKey))
		if v != nil {
			systems = strings.Split(string(v), ",")
		}
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("read indexed systems: %w", err)
	}
	return systems, nil
}

func writeIndexedSystems(db *bolt.DB, systems []string) error {
	if err := db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(BucketNames))
		v := b.Get([]byte(indexedSystemsKey))
		if v == nil {
			v = []byte(strings.Join(systems, ","))
			if err := b.Put([]byte(indexedSystemsKey), v); err != nil {
				return fmt.Errorf("write indexed systems: %w", err)
			}
			return nil
		}

		existing := strings.Split(string(v), ",")
		for _, system := range systems {
			if !utils.Contains(existing, system) {
				existing = append(existing, system)
			}
		}
		if err := b.Put([]byte(indexedSystemsKey), []byte(strings.Join(existing, ","))); err != nil {
			return fmt.Errorf("update indexed systems: %w", err)
		}
		return nil
	}); err != nil {
		return fmt.Errorf("update indexed systems transaction: %w", err)
	}
	return nil
}

type fileInfo struct {
	SystemID string
	Path     string
}

// Update the names index with the given files.
func updateNames(db *bolt.DB, files []fileInfo) error {
	if err := db.Batch(func(tx *bolt.Tx) error {
		bns := tx.Bucket([]byte(BucketNames))

		for _, file := range files {
			base := filepath.Base(file.Path)
			name := strings.TrimSuffix(base, filepath.Ext(base))

			nk := NameKey(file.SystemID, name)
			if err := bns.Put([]byte(nk), []byte(file.Path)); err != nil {
				return fmt.Errorf("index game name: %w", err)
			}
		}

		return nil
	}); err != nil {
		return fmt.Errorf("update names transaction: %w", err)
	}
	return nil
}

type IndexStatus struct {
	SystemID string
	Total    int
	Step     int
	Files    int
}

// Given a list of systems, index all valid game files on disk and write a
// names index to the DB. Overwrites any existing names index, but does not
// clean up old missing files.
//
// Takes a function which will be called with the current status of the index
// during key steps.
//
// Returns the total number of files indexed.
func NewNamesIndex(
	cfg *config.UserConfig,
	systems []games.System,
	update func(IndexStatus),
) (int, error) {
	status := IndexStatus{
		Total: len(systems) + 1,
		Step:  1,
	}

	db, err := openNames()
	if err != nil {
		return status.Files, fmt.Errorf("error opening gamesdb: %w", err)
	}
	defer func() { _ = db.Close() }()

	update(status)
	systemPaths := make(map[string][]string, 0)
	paths := games.GetSystemPaths(cfg, systems)
	for i := range paths {
		systemPaths[paths[i].System.Id] = append(systemPaths[paths[i].System.Id], paths[i].Path)
	}

	g := new(errgroup.Group)

	for _, k := range utils.AlphaMapKeys(systemPaths) {
		status.SystemID = k
		status.Step++
		update(status)

		files := make([]fileInfo, 0)

		for _, path := range systemPaths[k] {
			pathFiles, filesErr := games.GetFiles(k, path)
			if filesErr != nil {
				return status.Files, fmt.Errorf("error getting files: %w", filesErr)
			}

			if len(pathFiles) == 0 {
				continue
			}

			for pf := range pathFiles {
				files = append(files, fileInfo{SystemID: k, Path: pathFiles[pf]})
			}
		}

		if len(files) == 0 {
			continue
		}

		status.Files += len(files)

		g.Go(func() error {
			return updateNames(db, files)
		})
	}

	status.Step++
	status.SystemID = ""
	update(status)

	err = g.Wait()
	if err != nil {
		return status.Files, fmt.Errorf("error updating names index: %w", err)
	}

	err = writeIndexedSystems(db, utils.AlphaMapKeys(systemPaths))
	if err != nil {
		return status.Files, fmt.Errorf("error writing indexed systems: %w", err)
	}

	err = db.Sync()
	if err != nil {
		return status.Files, fmt.Errorf("error syncing database: %w", err)
	}

	return status.Files, nil
}

type SearchResult struct {
	SystemID string
	Name     string
	Path     string
}

// Iterate all indexed names and return matches to test func against query.
func searchNamesGeneric(
	systems []games.System,
	query string,
	test func(string, string) bool,
) ([]SearchResult, error) {
	if !DBExists() {
		return nil, errors.New("gamesdb does not exist")
	}

	db, err := open(&bolt.Options{ReadOnly: true})
	if err != nil {
		return nil, fmt.Errorf("open games database for search: %w", err)
	}
	defer func() { _ = db.Close() }()

	var results []SearchResult

	err = db.View(func(tx *bolt.Tx) error {
		bn := tx.Bucket([]byte(BucketNames))

		for i := range systems {
			system := &systems[i]
			pre := []byte(system.Id + ":")
			nameIdx := bytes.IndexByte(pre, ':')

			c := bn.Cursor()
			for k, v := c.Seek(pre); k != nil && bytes.HasPrefix(k, pre); k, v = c.Next() {
				keyName := string(k[nameIdx+1:])

				if test(query, keyName) {
					results = append(results, SearchResult{
						SystemID: system.Id,
						Name:     keyName,
						Path:     string(v),
					})
				}
			}
		}

		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("search names index: %w", err)
	}

	return results, nil
}

// Return indexed names partially matching query (case insensitive).
func SearchNamesPartial(systems []games.System, query string) ([]SearchResult, error) {
	return searchNamesGeneric(systems, query, func(query, keyName string) bool {
		return strings.Contains(strings.ToLower(keyName), strings.ToLower(query))
	})
}

// Return indexed names that include every word in query (case insensitive).
func SearchNamesWords(systems []games.System, query string) ([]SearchResult, error) {
	return searchNamesGeneric(systems, query, func(query, keyName string) bool {
		qWords := strings.Fields(strings.ToLower(query))

		for _, word := range qWords {
			if !strings.Contains(strings.ToLower(keyName), word) {
				return false
			}
		}

		return true
	})
}

// Return indexed names matching query using regular expression.
func SearchNamesRegexp(systems []games.System, query string) ([]SearchResult, error) {
	return searchNamesGeneric(systems, query, func(query, keyName string) bool {
		r, err := regexp.Compile(query)
		if err != nil {
			return false
		}

		return r.MatchString(keyName)
	})
}

// Return all systems indexed in the gamesdb
func IndexedSystems() ([]string, error) {
	if !DBExists() {
		return nil, errors.New("gamesdb does not exist")
	}

	db, err := open(&bolt.Options{ReadOnly: true})
	if err != nil {
		return nil, fmt.Errorf("open games database for metadata: %w", err)
	}
	defer func() { _ = db.Close() }()

	systems, err := readIndexedSystems(db)
	if err != nil {
		return nil, err
	}

	return systems, nil
}
