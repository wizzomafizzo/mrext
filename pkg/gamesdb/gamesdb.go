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
	bolt "go.etcd.io/bbolt"
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
	return openAt(config.GamesDB, options)
}

func openAt(path string, options *bolt.Options) (*bolt.DB, error) {
	readOnly := options != nil && options.ReadOnly
	if !readOnly {
		if err := os.MkdirAll(filepath.Dir(path), 0o750); err != nil {
			return nil, fmt.Errorf("create games database directory: %w", err)
		}
	}

	db, err := bolt.Open(path, 0o600, options)
	if err != nil {
		return nil, fmt.Errorf("open games database: %w", err)
	}

	if readOnly {
		if err := db.View(func(tx *bolt.Tx) error {
			if tx.Bucket([]byte(BucketNames)) == nil {
				return errors.New("games database has no names index")
			}
			return nil
		}); err != nil {
			_ = db.Close()
			return nil, fmt.Errorf("validate games database: %w", err)
		}
		return db, nil
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

func readIndexedSystems(db *bolt.DB) ([]string, error) {
	var systems []string

	err := db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(BucketNames))
		v := b.Get([]byte(indexedSystemsKey))
		if len(v) != 0 {
			systems = strings.Split(string(v), ",")
		}
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("read indexed systems: %w", err)
	}
	return systems, nil
}

type IndexStatus struct {
	SystemID string
	Total    int
	Step     int
	Files    int
	// Skipped counts game roots that could not be scanned, such as a folder
	// on a drive that is not attached. Those roots are reported rather than
	// failing the whole run.
	Skipped int
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
	return searchNamesAt(config.GamesDB, systems, query, test)
}

func searchNamesAt(
	path string, systems []games.System, query string, test func(string, string) bool,
) ([]SearchResult, error) {
	db, err := openAt(path, &bolt.Options{ReadOnly: true})
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
