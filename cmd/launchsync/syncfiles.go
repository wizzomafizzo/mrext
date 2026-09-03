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

package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/wizzomafizzo/mrext/pkg/config"
	"github.com/wizzomafizzo/mrext/pkg/games"
	"github.com/wizzomafizzo/mrext/pkg/gamesdb"
	"github.com/wizzomafizzo/mrext/pkg/mister"
	"github.com/wizzomafizzo/mrext/pkg/utils"
	"gopkg.in/ini.v1"
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
func readSectionName(sectionName string) (name, path string, err error) {
	parts := strings.Split(sectionName, "/")

	if len(parts) == 1 {
		// root level file
		return utils.StripBadFileChars(parts[0]), "", nil
	}

	name = utils.StripBadFileChars(parts[len(parts)-1])

	var folders []string

	for i := range len(parts) - 1 {
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

func readSyncFile(path string) (syncFile, error) {
	var sf syncFile

	cfg, err := ini.ShadowLoad(path)
	if err != nil {
		return sf, fmt.Errorf("load sync file: %w", err)
	}

	sf.path = path

	sf.name = cfg.Section("DEFAULT").Key("name").String()
	if sf.name == "" {
		return sf, errors.New("missing name field")
	}

	sf.folder = filepath.Join(filepath.Dir(path), "_"+utils.StripBadFileChars(sf.name))

	sf.author = cfg.Section("DEFAULT").Key("author").String()
	if sf.author == "" {
		return sf, errors.New("missing author field")
	}

	sf.url = cfg.Section("DEFAULT").Key("url").String()

	if cfg.Section("DEFAULT").HasKey("updated") {
		updated := cfg.Section("DEFAULT").Key("updated")
		sf.updated, err = updated.TimeFormat("2006-01-02")
		if err != nil {
			sf.updated, err = updated.TimeFormat("2006-01-02 15:04")
			if err != nil {
				return sf, fmt.Errorf("invalid updated date/time: %w", err)
			}
		}
	} else if sf.url != "" {
		return sf, errors.New("updated field is required with a url")
	}

	for _, section := range cfg.Sections() {
		if section.Name() == "DEFAULT" {
			continue
		}

		var game syncFileGame

		game.id = section.Name()

		game.name, game.folder, err = readSectionName(game.id)
		if err != nil {
			return sf, err
		}

		if game.name == "" {
			return sf, fmt.Errorf("missing name in %s", game.id)
		}

		systemName := section.Key("system").String()
		system, err := games.LookupSystem(systemName)
		if err != nil {
			return sf, fmt.Errorf("invalid system in %s: %w", game.id, err)
		}
		game.system = system

		matches := section.Key("match").ValueWithShadows()
		game.matches = append(game.matches, matches...)

		if len(game.matches) == 0 {
			return sf, fmt.Errorf("missing matches in %s", game.id)
		}

		sf.games = append(sf.games, game)
	}

	return sf, nil
}

// Update a sync file in place if it has been updated online.
func updateSyncFile(sync *syncFile) (syncFile, bool, error) {
	if sync.url == "" {
		return *sync, false, nil
	}

	req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, sync.url, http.NoBody)
	if err != nil {
		return *sync, false, fmt.Errorf("create sync request: %w", err)
	}
	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return *sync, false, fmt.Errorf("download sync file: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return *sync, false, fmt.Errorf("failed to download %s: %s", sync.url, resp.Status)
	}

	fp, err := os.CreateTemp("", "launchsync-")
	if err != nil {
		return *sync, false, fmt.Errorf("create temporary sync file: %w", err)
	}
	defer func() { _ = fp.Close() }()
	defer func() { _ = os.Remove(fp.Name()) }()

	if _, err = io.Copy(fp, resp.Body); err != nil {
		return *sync, false, fmt.Errorf("save downloaded sync file: %w", err)
	}
	if err = fp.Close(); err != nil {
		return *sync, false, fmt.Errorf("close downloaded sync file: %w", err)
	}

	newSync, err := readSyncFile(fp.Name())
	if err != nil {
		return *sync, false, err
	}

	if !newSync.updated.After(sync.updated) {
		return *sync, false, nil
	}

	newSync.path = sync.path
	newSync.folder = sync.folder
	if err := utils.MoveFile(fp.Name(), sync.path); err != nil {
		return *sync, false, fmt.Errorf("replace sync file: %w", err)
	}

	return newSync, true, nil
}

func makeIndex(cfg *config.UserConfig, syncs []syncFile) error {
	// restrict index to necessary systems
	var systems []games.System
	for syncIndex := range syncs {
		sync := &syncs[syncIndex]
		for gameIndex := range sync.games {
			game := &sync.games[gameIndex]
			systems = append(systems, *game.system)
		}
	}

	if len(systems) == 0 {
		return nil
	}

	_, err := gamesdb.NewNamesIndex(cfg, systems, func(gamesdb.IndexStatus) {})
	if err != nil {
		return fmt.Errorf("build game-name index: %w", err)
	}

	return nil
}

func checkForChanges(sync *syncFile) (syncFile, bool, error) {
	newSync, updated, err := updateSyncFile(sync)
	if err != nil {
		return *sync, false, err
	}

	if updated || sync.url == "" {
		var newPaths []string
		for gameIndex := range newSync.games {
			game := &newSync.games[gameIndex]
			path := filepath.Join(sync.folder, game.folder)
			newPaths = append(
				newPaths,
				mister.GetLauncherFilename(game.system, path, game.name),
				notFoundFilename(sync.folder, game),
			)
		}

		// delete removed games
		if _, statErr := os.Stat(sync.folder); statErr == nil {
			root, rootErr := os.OpenRoot(sync.folder)
			if rootErr != nil {
				return newSync, true, fmt.Errorf("open sync output root: %w", rootErr)
			}
			defer func() { _ = root.Close() }()

			walkErr := fs.WalkDir(root.FS(), ".", func(path string, info fs.DirEntry, walkErr error) error {
				if walkErr != nil {
					return walkErr
				}
				if info.IsDir() || utils.Contains(newPaths, filepath.Join(sync.folder, path)) {
					return nil
				}
				if err := root.Remove(path); err != nil {
					return fmt.Errorf("remove stale launcher %s: %w", path, err)
				}
				return nil
			})
			if walkErr != nil {
				return newSync, true, fmt.Errorf("remove stale launchers: %w", walkErr)
			}

			files, readErr := fs.ReadDir(root.FS(), ".")
			if readErr != nil {
				return newSync, true, fmt.Errorf("read sync output directory: %w", readErr)
			}
			for _, file := range files {
				if file.IsDir() {
					path := filepath.Join(sync.folder, file.Name())
					if err := utils.RemoveEmptyDirs(path); err != nil {
						return newSync, true, fmt.Errorf("remove empty sync directories: %w", err)
					}
				}
			}
		}

		return newSync, true, nil
	}
	return *sync, false, nil
}

func notFoundFilename(folder string, game *syncFileGame) string {
	return filepath.Join(folder, game.folder, game.name+" [NOT FOUND].mgl")
}

func tryLinkGame(
	cfg *config.UserConfig,
	sync *syncFile,
	game *syncFileGame,
) (filename string, found bool, err error) {
	var match gamesdb.SearchResult

	for _, m := range game.matches {
		var results []gamesdb.SearchResult
		var searchErr error

		if m == "" {
			continue
		}

		if m[0] == '~' {
			// regex match
			if m[1:] == "" {
				continue
			}
			results, searchErr = gamesdb.SearchNamesRegexp([]games.System{*game.system}, "(?i)"+m[1:])
			if searchErr != nil {
				return "", false, fmt.Errorf("search games by regular expression: %w", searchErr)
			}
		} else {
			// partial match
			results, searchErr = gamesdb.SearchNamesPartial([]games.System{*game.system}, m)
			if searchErr != nil {
				return "", false, fmt.Errorf("search games by partial name: %w", searchErr)
			}
		}

		if len(results) > 0 {
			match = results[0]
			break
		}
	}

	// top level folder creation
	if _, statErr := os.Stat(sync.folder); statErr != nil {
		// #nosec G301,G703 -- generated launcher directories must remain world-readable.
		mkdirErr := os.Mkdir(sync.folder, 0o755)
		if mkdirErr != nil {
			return "", false, fmt.Errorf("create sync output directory: %w", mkdirErr)
		}
	}

	// optional subfolder creation
	if game.folder != "" {
		// #nosec G301,G703 -- generated launcher directories must remain world-readable.
		mkdirErr := os.MkdirAll(filepath.Join(sync.folder, game.folder), 0o755)
		if mkdirErr != nil {
			return "", false, fmt.Errorf("create sync launcher subdirectory: %w", mkdirErr)
		}
	}

	launcherFolder := filepath.Join(sync.folder, game.folder)
	launcherFn := mister.GetLauncherFilename(game.system, launcherFolder, game.name)
	notFoundFn := notFoundFilename(sync.folder, game)

	if match.Name != "" {
		// found a match
		// TODO: don't write if it's the same file
		_, launchErr := mister.CreateLauncher(cfg, game.system, match.Path, launcherFolder, game.name)
		if launchErr != nil {
			return "", false, fmt.Errorf("create game launcher: %w", launchErr)
		}

		_ = os.Remove(notFoundFn)

		return filepath.Base(match.Path), true, nil
	}

	// no match
	// #nosec G306,G703 -- MiSTer menu placeholder must remain world-readable.
	if writeErr := os.WriteFile(notFoundFn, nil, 0o644); writeErr != nil {
		return "", false, fmt.Errorf("create not-found launcher: %w", writeErr)
	}
	_ = os.Remove(launcherFn)
	return "", false, nil
}
