package gamesdb

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	bolt "go.etcd.io/bbolt"
	"golang.org/x/sync/errgroup"

	"github.com/wizzomafizzo/mrext/pkg/config"
	"github.com/wizzomafizzo/mrext/pkg/games"
	"github.com/wizzomafizzo/mrext/pkg/utils"
)

const (
	BucketNames = "names"
)

// Return the key for a name in the names index.
func NameKey(systemId string, name string) string {
	return systemId + ":" + name
}

// Check if the gamesdb exists on disk.
func DbExists() bool {
	_, err := os.Stat(config.GamesDb)
	return err == nil
}

// Open the gamesdb with the given options. If the database does not exist it
// will be created and the buckets will be initialized.
func open(options *bolt.Options) (*bolt.DB, error) {
	err := os.MkdirAll(filepath.Base(config.GamesDb), 0755)
	if err != nil {
		return nil, err
	}

	db, err := bolt.Open(config.GamesDb, 0600, options)
	if err != nil {
		return nil, err
	}

	db.Update(func(txn *bolt.Tx) error {
		for _, bucket := range []string{BucketNames} {
			_, err := txn.CreateBucketIfNotExists([]byte(bucket))
			if err != nil {
				return err
			}
		}

		return nil
	})

	return db, nil
}

// Open the gamesdb with default options for generating names index.
func openNames() (*bolt.DB, error) {
	return open(&bolt.Options{
		NoSync:         true,
		NoFreelistSync: true,
	})
}

type fileInfo struct {
	SystemId string
	Path     string
}

// Update the names index with the given files.
func updateNames(db *bolt.DB, files []fileInfo) error {
	return db.Batch(func(tx *bolt.Tx) error {
		bns := tx.Bucket([]byte(BucketNames))

		for _, file := range files {
			base := filepath.Base(file.Path)
			name := strings.TrimSuffix(base, filepath.Ext(base))

			nk := NameKey(file.SystemId, name)
			err := bns.Put([]byte(nk), []byte(file.Path))
			if err != nil {
				return err
			}
		}

		return nil
	})
}

type IndexStatus struct {
	Total    int
	Step     int
	SystemId string
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
	systems []games.System,
	update func(IndexStatus),
) (int, error) {
	status := IndexStatus{
		Total: len(systems) + 1,
		Step:  1,
	}

	db, err := openNames()
	if err != nil {
		return status.Files, fmt.Errorf("error opening gamesdb: %s", err)
	}
	defer db.Close()

	update(status)
	systemPaths := make(map[string][]string, 0)
	for _, v := range games.GetSystemPaths(&config.UserConfig{}, systems) {
		systemPaths[v.System.Id] = append(systemPaths[v.System.Id], v.Path)
	}

	g := new(errgroup.Group)

	for _, k := range utils.AlphaMapKeys(systemPaths) {
		status.SystemId = k
		status.Step++
		update(status)

		files := make([]fileInfo, 0)

		for _, path := range systemPaths[k] {
			pathFiles, err := games.GetFiles(k, path)
			if err != nil {
				return status.Files, fmt.Errorf("error getting files: %s", err)
			}

			if len(pathFiles) == 0 {
				continue
			}

			for pf := range pathFiles {
				files = append(files, fileInfo{SystemId: k, Path: pathFiles[pf]})
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
	status.SystemId = ""
	update(status)

	err = g.Wait()
	if err != nil {
		return status.Files, fmt.Errorf("error updating names index: %s", err)
	}

	err = db.Sync()
	if err != nil {
		return status.Files, fmt.Errorf("error syncing database: %s", err)
	}

	return status.Files, nil
}

type SearchResult struct {
	SystemId string
	Name     string
	Path     string
}

// Iterate all indexed names and return matches to test func against query.
func searchNamesGeneric(
	systems []games.System,
	query string,
	test func(string, string) bool,
) ([]SearchResult, error) {
	if !DbExists() {
		return nil, fmt.Errorf("gamesdb does not exist")
	}

	db, err := open(&bolt.Options{ReadOnly: true})
	if err != nil {
		return nil, err
	}

	var results []SearchResult

	err = db.View(func(tx *bolt.Tx) error {
		bn := tx.Bucket([]byte(BucketNames))

		for _, system := range systems {
			pre := []byte(system.Id + ":")
			nameIdx := bytes.Index(pre, []byte(":"))

			c := bn.Cursor()
			for k, v := c.Seek([]byte(pre)); k != nil && bytes.HasPrefix(k, pre); k, v = c.Next() {
				keyName := string(k[nameIdx+1:])

				if test(query, keyName) {
					results = append(results, SearchResult{
						SystemId: system.Id,
						Name:     keyName,
						Path:     string(v),
					})
				}
			}
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return results, nil
}

// Return indexed names matching exact query (case insensitive).
func SearchNamesExact(systems []games.System, query string) ([]SearchResult, error) {
	return searchNamesGeneric(systems, query, func(query, keyName string) bool {
		return strings.EqualFold(query, keyName)
	})
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
