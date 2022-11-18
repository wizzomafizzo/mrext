package main

import (
	"fmt"
	"io"
	"io/fs"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"gopkg.in/ini.v1"

	"github.com/wizzomafizzo/mrext/pkg/games"
	"github.com/wizzomafizzo/mrext/pkg/mister"
	"github.com/wizzomafizzo/mrext/pkg/txtindex"
	"github.com/wizzomafizzo/mrext/pkg/utils"
)

type syncFileGame struct {
	id      string
	name    string
	folder  string
	system  *games.System
	matches []string
}

type syncFile struct {
	name    string
	author  string
	url     string
	updated time.Time
	folder  string
	path    string
	games   []syncFileGame
}

func getSyncFiles(paths []string) []string {
	var syncFiles []string
	for _, path := range paths {
		files, _ := filepath.Glob(filepath.Join(path, "*.sync"))
		if len(files) > 0 {
			syncFiles = append(syncFiles, files...)
		}
	}
	return syncFiles
}

// Parse a section name and return a cleaned and formatted filename and relative folder path.
func readSectionName(sectionName string) (name string, path string, err error) {
	parts := strings.Split(sectionName, "/")

	if len(parts) < 1 {
		return "", "", fmt.Errorf("invalid section name: %s", sectionName)
	} else if len(parts) == 1 {
		// root level file
		return utils.StripBadFileChars(parts[0]), "", nil
	}

	name = utils.StripBadFileChars(parts[len(parts)-1])

	var folders []string

	for i := 0; i < len(parts)-1; i++ {
		fn := utils.StripBadFileChars(parts[i])

		if fn == "" || fn == "." || fn == ".." || fn == "_" {
			break
		}

		if fn[0] != '_' {
			fn = "_" + fn
		}

		folders = append(folders, fn)
	}

	path = filepath.Join(folders...)

	return name, path, nil
}

func readSyncFile(path string) (*syncFile, error) {
	var sf syncFile

	cfg, err := ini.ShadowLoad(path)
	if err != nil {
		return nil, err
	}

	sf.path = path

	sf.name = cfg.Section("DEFAULT").Key("name").String()
	if sf.name == "" {
		return nil, fmt.Errorf("missing name field")
	}

	sf.folder = filepath.Join(filepath.Dir(path), "_"+utils.StripBadFileChars(sf.name))

	sf.author = cfg.Section("DEFAULT").Key("author").String()
	if sf.author == "" {
		return nil, fmt.Errorf("missing author field")
	}

	sf.url = cfg.Section("DEFAULT").Key("url").String()

	if cfg.Section("DEFAULT").HasKey("updated") {
		updated := cfg.Section("DEFAULT").Key("updated")
		sf.updated, err = updated.TimeFormat("2006-01-02")
		if err != nil {
			sf.updated, err = updated.TimeFormat("2006-01-02 15:04")
			if err != nil {
				return nil, fmt.Errorf("invalid updated date/time: %s", err)
			}
		}
	} else if sf.url != "" {
		return nil, fmt.Errorf("updated field is required with a url")
	}

	for _, section := range cfg.Sections() {
		if section.Name() == "DEFAULT" {
			continue
		}

		var game syncFileGame

		game.id = section.Name()

		game.name, game.folder, err = readSectionName(game.id)
		if err != nil {
			return nil, err
		}

		if game.name == "" {
			return nil, fmt.Errorf("missing name in %s", game.id)
		}

		systemName := section.Key("system").String()
		system, err := games.LookupSystem(systemName)
		if err != nil {
			return nil, fmt.Errorf("invalid system in %s: %s", game.id, err)
		} else {
			game.system = system
		}

		matches := section.Key("match").ValueWithShadows()
		game.matches = append(game.matches, matches...)

		if len(game.matches) == 0 {
			return nil, fmt.Errorf("missing matches in %s", game.id)
		}

		sf.games = append(sf.games, game)
	}

	return &sf, nil
}

// Update a sync file in place if it has been updated online.
func updateSyncFile(sync *syncFile) (*syncFile, bool, error) {
	if sync.url == "" {
		return sync, false, nil
	}

	resp, err := http.Get(sync.url)
	if err != nil {
		return nil, false, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, false, fmt.Errorf("failed to download %s: %s", sync.url, resp.Status)
	}

	fp, err := os.CreateTemp("", "launchsync-")
	if err != nil {
		return nil, false, err
	}
	defer fp.Close()
	defer os.Remove(fp.Name())

	_, err = io.Copy(fp, resp.Body)
	if err != nil {
		return nil, false, err
	}
	fp.Close()

	newSync, err := readSyncFile(fp.Name())
	if err != nil {
		return nil, false, err
	}

	if newSync.updated.After(sync.updated) {
		newSync.path = sync.path
		newSync.folder = sync.folder

		err := utils.MoveFile(fp.Name(), sync.path)
		if err != nil {
			return nil, false, err
		}

		return newSync, true, nil
	} else {
		return sync, false, nil
	}
}

func makeIndex(syncs []*syncFile) (txtindex.Index, error) {
	var index txtindex.Index
	indexFile := filepath.Join(os.TempDir(), "launchsync-index.tar")

	// Restrict index to necessary systems
	var systems []games.System
	for _, sync := range syncs {
		for _, game := range sync.games {
			systems = append(systems, *game.system)
		}
	}

	systemPaths := games.GetSystemPaths(systems)

	systemFiles := make([][2]string, 0)
	for _, path := range systemPaths {
		files, err := games.GetFiles(path.System.Id, path.Path)
		if err != nil {
			return index, err
		}

		for _, file := range files {
			systemFiles = append(systemFiles, [2]string{path.System.Id, file})
		}
	}

	err := txtindex.Generate(systemFiles, indexFile)
	if err != nil {
		return index, err
	}

	index, err = txtindex.Open(indexFile)
	if err != nil {
		return index, err
	}
	os.Remove(indexFile)

	return index, nil
}

func checkForChanges(sync *syncFile) (*syncFile, bool, error) {
	newSync, updated, err := updateSyncFile(sync)
	if err != nil {
		return sync, false, err
	}

	if updated || sync.url == "" {
		var newPaths []string
		for _, game := range newSync.games {
			path := filepath.Join(sync.folder, game.folder)
			newPaths = append(newPaths, mister.GetLauncherFilename(game.system, path, game.name))
			newPaths = append(newPaths, notFoundFilename(sync.folder, game))
		}

		// delete removed games
		if _, ok := os.Stat(sync.folder); ok == nil {
			err := filepath.WalkDir(sync.folder, func(path string, info fs.DirEntry, err error) error {
				if err != nil {
					return err
				}

				if !info.IsDir() {
					if !utils.Contains(newPaths, path) {
						return os.Remove(path)
					}
				}

				return nil
			})

			if err != nil {
				return newSync, true, err
			}

			// delete empty folders
			files, err := os.ReadDir(sync.folder)
			if err != nil {
				return newSync, true, err
			}

			for _, file := range files {
				if file.IsDir() {
					path := filepath.Join(sync.folder, file.Name())
					err = utils.RemoveEmptyDirs(path)
					if err != nil {
						return newSync, true, err
					}
				}
			}
		}

		return newSync, true, nil
	} else {
		return sync, false, nil
	}
}

func notFoundFilename(folder string, game syncFileGame) string {
	return filepath.Join(folder, game.folder, game.name+" [NOT FOUND].mgl")
}

func tryLinkGame(sync *syncFile, game syncFileGame, index txtindex.Index) (string, bool, error) {
	var match txtindex.SearchResult

	for _, m := range game.matches {
		var results []txtindex.SearchResult

		if m == "" {
			continue
		}

		if m[0] == '~' {
			// regex match
			if m[1:] == "" {
				continue
			}
			results = index.SearchSystemByNameRe(game.system.Id, "(?i)"+m[1:])
		} else {
			// partial match
			results = index.SearchSystemByName(game.system.Id, m)
		}

		if len(results) > 0 {
			match = results[0]
			break
		}
	}

	// top level folder creation
	if _, ok := os.Stat(sync.folder); ok != nil {
		err := os.Mkdir(sync.folder, 0755)
		if err != nil {
			return "", false, err
		}
	}

	// optional subfolder creation
	if game.folder != "" {
		err := os.MkdirAll(filepath.Join(sync.folder, game.folder), 0755)
		if err != nil {
			return "", false, err
		}
	}

	launcherFolder := filepath.Join(sync.folder, game.folder)
	launcherFn := mister.GetLauncherFilename(game.system, launcherFolder, game.name)
	notFoundFn := notFoundFilename(sync.folder, game)

	if match.Name != "" {
		// found a match
		// TODO: don't write if it's the same file
		_, err := mister.CreateLauncher(game.system, match.Path, launcherFolder, game.name)
		if err != nil {
			return "", false, err
		}

		os.Remove(notFoundFn)

		return filepath.Base(match.Path), true, nil
	} else {
		// no match
		fp, err := os.Create(notFoundFn)
		if err != nil {
			return "", false, err
		}
		defer fp.Close()

		os.Remove(launcherFn)

		return "", false, nil
	}
}
