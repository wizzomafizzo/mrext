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

package games

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/wizzomafizzo/mrext/pkg/config"
)

func GetGamesFolders(cfg *config.UserConfig) []string {
	var folders []string
	for _, folder := range cfg.Systems.GamesFolder {
		folder = filepath.Clean(folder)
		if !strings.HasSuffix(folder, "/games") {
			folders = append(folders, filepath.Join(folder, "games"))
		}
		folders = append(folders, folder)
	}
	folders = append(folders, config.GamesFolders...)
	return folders
}

func FindFile(path string) (string, error) {
	if _, err := os.Stat(path); err == nil {
		return path, nil
	}

	parent := filepath.Dir(path)
	name := filepath.Base(path)

	files, err := os.ReadDir(parent)
	if err != nil {
		return "", fmt.Errorf("read parent directory: %w", err)
	}

	for _, file := range files {
		target := file.Name()

		if len(target) != len(name) {
			continue
		}
		if strings.EqualFold(target, name) {
			return filepath.Join(parent, target), nil
		}
	}

	return "", fmt.Errorf("file match not found: %s", path)
}

// FolderToSystems returns what systems a path could be for.
func FolderToSystems(cfg *config.UserConfig, path string) []System {
	path = strings.ToLower(path)
	validGamesFolder := false
	gamesFolder := ""

	for _, folder := range GetGamesFolders(cfg) {
		if strings.HasPrefix(path, strings.ToLower(folder)) {
			validGamesFolder = true
			gamesFolder = folder
			break
		}
	}

	if !validGamesFolder {
		return nil
	}

	var validSystems []System
	for id := range Systems {
		system := Systems[id]
		for _, folder := range system.Folder {
			systemPath := strings.ToLower(filepath.Join(gamesFolder, folder))
			if strings.HasPrefix(path, systemPath) {
				validSystems = append(validSystems, system)
				break
			}
		}
	}

	if strings.HasSuffix(path, "/") {
		return validSystems
	}

	var matchedExtensions []System
	for i := range validSystems {
		if MatchSystemFile(&validSystems[i], path) {
			matchedExtensions = append(matchedExtensions, validSystems[i])
		}
	}

	if len(matchedExtensions) == 0 {
		// fall back to just the folder match
		return validSystems
	}

	return matchedExtensions
}

func BestSystemMatch(cfg *config.UserConfig, path string) (System, error) {
	systems := FolderToSystems(cfg, path)

	if len(systems) == 0 {
		return System{}, fmt.Errorf("no systems found for %s", path)
	}

	if len(systems) == 1 {
		return systems[0], nil
	}

	// check for system matches by file extension if possible
	if filepath.Ext(path) != "" {
		filtered := []System{}
		for i := range systems {
			if MatchSystemFile(&systems[i], path) {
				filtered = append(filtered, systems[i])
			}
		}

		if len(filtered) > 0 {
			systems = filtered
		}
	}

	// prefer the system with a setname
	for i := range systems {
		if systems[i].SetName != "" {
			return systems[i], nil
		}
	}

	// otherwise just return the first one
	return systems[0], nil
}

//nolint:govet // Field order preserves legacy JSON output.
type PathResult struct {
	System System
	Path   string
}

// GetSystemPaths returns all possible paths for each system.
func GetSystemPaths(cfg *config.UserConfig, systems []System) []PathResult {
	var matches []PathResult

	gamesFolders := GetGamesFolders(cfg)
	for i := range systems {
		system := &systems[i]
		for _, gamesFolder := range gamesFolders {
			gf, err := FindFile(gamesFolder)
			if err != nil {
				continue
			}

			for _, folder := range system.Folder {
				systemFolder := filepath.Join(gf, folder)
				path, err := FindFile(systemFolder)
				if err != nil {
					continue
				}

				matches = append(matches, PathResult{Path: path, System: *system})
			}
		}
	}

	return matches
}

// GetActiveSystemPaths returns the active path for each system.
func GetActiveSystemPaths(cfg *config.UserConfig, systems []System) []PathResult {
	var matches []PathResult

	gamesFolders := GetGamesFolders(cfg)
	for i := range systems {
		system := &systems[i]
		for _, gamesFolder := range gamesFolders {
			gf, err := FindFile(gamesFolder)
			if err != nil {
				continue
			}

			found := false

			for _, folder := range system.Folder {
				systemFolder := filepath.Join(gf, folder)
				path, err := FindFile(systemFolder)
				if err != nil {
					continue
				}

				matches = append(matches, PathResult{Path: path, System: *system})
				found = true
				break
			}

			if found {
				break
			}
		}

		if len(matches) == len(systems) {
			break
		}
	}

	return matches
}

func GetPopulatedGamesFolders(cfg *config.UserConfig, systems []System) map[string][]string {
	results := GetSystemPaths(cfg, systems)
	if len(results) == 0 {
		return nil
	}

	populated := make(map[string][]string)

	for i := range results {
		folder := &results[i]
		files, err := os.ReadDir(folder.Path)
		if err != nil {
			continue
		}

		if len(files) > 0 {
			populated[folder.System.Id] = append(populated[folder.System.Id], folder.Path)
		}
	}

	return populated
}
