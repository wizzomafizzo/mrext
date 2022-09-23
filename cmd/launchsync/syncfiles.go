package main

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"gopkg.in/ini.v1"

	"github.com/wizzomafizzo/mrext/pkg/games"
	"github.com/wizzomafizzo/mrext/pkg/mister"
	"github.com/wizzomafizzo/mrext/pkg/txtindex"
	"github.com/wizzomafizzo/mrext/pkg/utils"
)

type syncFileGame struct {
	name    string
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

	if cfg.Section("DEFAULT").HasKey("url") {
		sf.updated, err = cfg.Section("DEFAULT").Key("updated").TimeFormat("2006-01-02")
		if err != nil {
			sf.updated, err = cfg.Section("DEFAULT").Key("updated").TimeFormat("2006-01-02 15:04")
			if err != nil {
				return nil, fmt.Errorf("invalid updated date/time: %s", err)
			}
		}
	}

	for _, section := range cfg.Sections() {
		if section.Name() == "DEFAULT" {
			continue
		}

		var game syncFileGame

		// TODO: support subfolders
		game.name = utils.StripBadFileChars(section.Name())

		if game.name == "" {
			return nil, fmt.Errorf("missing name in %s", section.Name())
		}

		systemName := section.Key("system").String()
		system, err := games.LookupSystem(systemName)
		if err != nil {
			return nil, fmt.Errorf("invalid system in %s: %s", section.Name(), err)
		} else {
			game.system = system
		}

		matches := section.Key("match").ValueWithShadows()
		game.matches = append(game.matches, matches...)

		if len(game.matches) == 0 {
			return nil, fmt.Errorf("missing matches in %s", section.Name())
		}

		sf.games = append(sf.games, game)
	}

	if len(sf.games) == 0 {
		return nil, fmt.Errorf("no games found")
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
	var systems []*games.System
	for _, sync := range syncs {
		for _, game := range sync.games {
			systems = append(systems, game.system)
		}
	}

	systemPaths := make(map[string][]string)
	for systemId, path := range games.GetSystemPaths() {
		for _, system := range systems {
			if system.Id == systemId {
				systemPaths[systemId] = path
				break
			}
		}
	}

	systemFiles := make([][2]string, 0)
	for systemId, paths := range systemPaths {
		for _, path := range paths {
			files, err := games.GetFiles(systemId, path)
			if err != nil {
				return index, err
			}

			for _, file := range files {
				systemFiles = append(systemFiles, [2]string{systemId, file})
			}
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

func checkForUpdate(sync *syncFile) (*syncFile, bool, error) {
	// TODO: diff sync/removals could work without a url
	newSync, updated, err := updateSyncFile(sync)
	if err != nil {
		return sync, false, err
	}

	if updated {
		var newNames []string
		for _, game := range newSync.games {
			newNames = append(newNames, game.name)
		}

		// delete removed games
		if _, ok := os.Stat(sync.folder); ok == nil {
			for _, game := range sync.games {
				if !utils.Contains(newNames, game.name) {
					mister.DeleteLauncher(mister.GetLauncherFilename(game.system, sync.folder, game.name))
					os.Remove(notFoundFilename(sync.folder, game.name))
				}
			}
		}

		return newSync, true, nil
	} else {
		return sync, false, nil
	}
}

func notFoundFilename(folder string, name string) string {
	return filepath.Join(folder, name+" [NOT FOUND].mgl")
}

func tryLinkGame(sync *syncFile, game syncFileGame, index txtindex.Index) (string, bool, error) {
	var match txtindex.SearchResult

	for _, m := range game.matches {
		var results []txtindex.SearchResult

		if m == "" {
			continue
		}

		// TODO: include extension in regex search?
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

	if _, ok := os.Stat(sync.folder); ok != nil {
		err := os.Mkdir(sync.folder, 0755)
		if err != nil {
			return "", false, err
		}
	}

	if match.Name != "" {
		// TODO: don't write if it's the same file
		_, err := mister.CreateLauncher(game.system, match.Path, sync.folder, game.name)
		if err != nil {
			return "", false, err
		}

		if _, err := os.Stat(notFoundFilename(sync.folder, game.name)); err == nil {
			os.Remove(notFoundFilename(sync.folder, game.name))
		}

		return filepath.Base(match.Path), true, nil
	} else {
		fp, err := os.Create(notFoundFilename(sync.folder, game.name))
		if err != nil {
			return "", false, err
		}
		defer fp.Close()

		return "", false, nil
	}
}
