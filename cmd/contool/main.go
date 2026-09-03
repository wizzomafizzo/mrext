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
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/wizzomafizzo/mrext/pkg/config"
	"github.com/wizzomafizzo/mrext/pkg/games"
	"github.com/wizzomafizzo/mrext/pkg/mister"
	"github.com/wizzomafizzo/mrext/pkg/utils"
)

// Alternate names for systems.
var idMap = map[string]string{
	// "TurboGrafx16":   "tgfx16",
}

// Only allow these extensions to be indexed.
// Any systems not listed will allow all extensions.
var extMap = map[string][]string{
	// "Atari5200":    {".a52", ".car"},
}

// Convert an internal system ID to a mistercon ID if possible.
func conID(id string) string {
	if mappedID, ok := idMap[id]; ok {
		return mappedID
	}

	return id
}

// Convert a mistercon system ID to an internal ID if possible.
func reverseID(id string) string {
	for k, v := range idMap {
		if strings.EqualFold(v, id) {
			return k
		}
	}

	return id
}

// Return the filename of the gamelist for a given system ID.
func gamelistFilename(systemID string) string {
	var prefix string
	if id, ok := idMap[systemID]; ok {
		prefix = id
	} else {
		prefix = systemID
	}

	return strings.ToLower(prefix) + ".txt"
}

// Generate a gamelist file for a system with given results.
func writeGamelist(gamelistDir, systemID string, files []string) error {
	gamelistPath := filepath.Join(gamelistDir, gamelistFilename(systemID))
	tmpPath, err := os.CreateTemp("", "gamelist-*.txt")
	if err != nil {
		return fmt.Errorf("create temporary game list: %w", err)
	}
	defer func() { _ = tmpPath.Close() }()

	for _, file := range files {
		if _, err := tmpPath.WriteString(file + "\n"); err != nil {
			return fmt.Errorf("write temporary game list: %w", err)
		}
	}
	if err := tmpPath.Sync(); err != nil {
		return fmt.Errorf("sync temporary game list: %w", err)
	}
	if err := tmpPath.Close(); err != nil {
		return fmt.Errorf("close temporary game list: %w", err)
	}
	if err := utils.MoveFile(tmpPath.Name(), gamelistPath); err != nil {
		return fmt.Errorf("install game list: %w", err)
	}
	return nil
}

// Generate gamelists for all systems. Main workflow of app.
func createGamelists(
	gamelistDir string,
	systemPaths map[string][]string,
	progress, quiet, filter bool,
) int {
	start := time.Now()

	if !quiet && !progress {
		_, _ = fmt.Println("Finding system folders...")
	}

	// prep calculating progress
	totalPaths := 0
	for _, v := range systemPaths {
		totalPaths += len(v)
	}
	totalSteps := totalPaths
	currentStep := 0

	// generate file list
	totalGames := 0
	for systemID, paths := range systemPaths {
		var systemFiles []string

		for _, path := range paths {
			if !quiet {
				if progress {
					_, _ = fmt.Println("XXX")
					_, _ = fmt.Println(int(float64(currentStep) / float64(totalSteps) * 100))
					_, _ = fmt.Printf("Scanning %s (%s)\n", systemID, path)
					_, _ = fmt.Println("XXX")
				} else {
					_, _ = fmt.Printf("Scanning %s: %s\n", systemID, path)
				}
			}

			files, err := games.GetFiles(systemID, path)
			if err != nil {
				_, _ = fmt.Fprintln(os.Stderr, err)
				continue
			}
			systemFiles = append(systemFiles, files...)

			currentStep++
		}

		if filter {
			systemFiles = games.FilterUniqueFilenames(systemFiles)
		}

		// filter out certain extensions
		var filteredFiles []string
		if filterExts, ok := extMap[systemID]; ok {
			for _, file := range systemFiles {
				path := strings.ToLower(file)
				for _, ext := range filterExts {
					if strings.HasSuffix(path, ext) {
						filteredFiles = append(filteredFiles, file)
						break
					}
				}
			}
			systemFiles = filteredFiles
		}

		if len(systemFiles) > 0 {
			totalGames += len(systemFiles)
			if err := writeGamelist(gamelistDir, systemID, systemFiles); err != nil {
				panic(err)
			}
		}
	}

	if !quiet {
		taken := int(time.Since(start).Seconds())
		if progress {
			_, _ = fmt.Println("XXX")
			_, _ = fmt.Println("100")
			_, _ = fmt.Printf("Indexing complete (%d games in %ds)\n", totalGames, taken)
			_, _ = fmt.Println("XXX")
		} else {
			_, _ = fmt.Printf("Indexing complete (%d games in %ds)\n", totalGames, taken)
		}
	}

	return totalGames
}

func tryLaunchGame(cfg *config.UserConfig, launchPath string) error {
	system, err := games.BestSystemMatch(&config.UserConfig{}, launchPath)
	if err != nil {
		return fmt.Errorf("error during launch: %w", err)
	}

	err = mister.LaunchGame(cfg, &system, launchPath)
	if err != nil {
		return fmt.Errorf("error during launch: %w", err)
	}

	return nil
}

func main() {
	gamelistDir := flag.String("out", ".", "gamelist files directory")
	filter := flag.String("filter", "all", "list of systems to index (comma separated)")
	progress := flag.Bool("progress", false, "print output for dialog gauge")
	quiet := flag.Bool("quiet", false, "suppress all status output")
	detect := flag.Bool("detect", false, "list active system folders")
	noDupes := flag.Bool("nodupes", false, "filter out duplicate games")
	launchPath := flag.String("launch", "", "launch game with given path")
	flag.Parse()

	// launch game
	if *launchPath != "" {
		err := tryLaunchGame(&config.UserConfig{}, *launchPath)
		if err != nil {
			_, _ = fmt.Println("Error launching game:", err)
			os.Exit(1)
		}
		os.Exit(0)
	}

	// filter systems
	var systems []games.System
	if *filter == "all" {
		systems = games.AllSystems()
	} else {
		for _, filterID := range strings.Split(*filter, ",") {
			systemID := reverseID(filterID)

			if system, ok := games.Systems[systemID]; ok {
				systems = append(systems, system)
				continue
			}

			system, err := games.LookupSystem(systemID)
			if err != nil {
				continue
			}

			systems = append(systems, *system)
		}
	}

	// find active system paths
	if *detect {
		results := games.GetActiveSystemPaths(&config.UserConfig{}, systems)
		for i := range results {
			r := &results[i]
			_, _ = fmt.Printf("%s:%s\n", strings.ToLower(conID(r.System.Id)), r.Path)
		}
		os.Exit(0)
	}

	systemPaths := games.GetSystemPaths(&config.UserConfig{}, systems)
	systemPathsMap := make(map[string][]string)

	for i := range systemPaths {
		path := &systemPaths[i]
		systemPathsMap[path.System.Id] = append(systemPathsMap[path.System.Id], path.Path)
	}

	total := createGamelists(*gamelistDir, systemPathsMap, *progress, *quiet, *noDupes)

	if total == 0 {
		os.Exit(8)
	}
	os.Exit(0)
}
