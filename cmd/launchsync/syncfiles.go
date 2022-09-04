package main

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"gopkg.in/ini.v1"

	"github.com/wizzomafizzo/mrext/pkg/games"
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

	sf.folder = filepath.Dir(path)
	sf.path = path

	sf.name = cfg.Section("DEFAULT").Key("name").String()
	if sf.name == "" {
		return nil, fmt.Errorf("missing name field")
	}

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
		strippedName := section.Name()
		for _, char := range []string{"<", ">", ":", "\"", "/", "\\", "|", "?", "*"} {
			strippedName = strings.ReplaceAll(strippedName, char, "")
		}

		game.name = strippedName

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
		for _, match := range matches {
			// escape these regex characters for convenience matching rom filenames
			escapedMatch := match
			for _, char := range []string{"(", ")", "[", "]"} {
				escapedMatch = strings.ReplaceAll(escapedMatch, char, "\\"+char)
			}
			game.matches = append(game.matches, "(?i)"+escapedMatch)
		}

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
