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
	"strings"

	"github.com/wizzomafizzo/mrext/pkg/config"
	"github.com/wizzomafizzo/mrext/pkg/games"
	"github.com/wizzomafizzo/mrext/pkg/mister"
	"github.com/wizzomafizzo/mrext/pkg/utils"
)

const (
	appName         = "random"
	maxPickAttempts = 100
)

func main() {
	// TODO: support an ini file for default values

	filter := flag.String("filter", "", "list of systems to filter (ex. gba,psx,nes)")
	ignore := flag.String("ignore", "", "list of systems to ignore (ex. tgfx16-cd)")
	noscan := flag.Bool("noscan", false, "don't index entire system (faster, but less random)")
	flag.Parse()

	cfg, err := config.LoadUserConfig(appName, &config.UserConfig{})
	if err != nil {
		_, _ = fmt.Println("Error loading config file:", err)
		os.Exit(1)
	}

	filteredIDs := strings.Split(*filter, ",")
	var filteredSystems []games.System
	for _, id := range filteredIDs {
		system, _ := games.LookupSystem(id)
		if system != nil {
			filteredSystems = append(filteredSystems, *system)
		}
	}

	ignoredIDs := strings.Split(*ignore, ",")
	var ignoredSystems []games.System
	for _, id := range ignoredIDs {
		found, _ := games.LookupSystem(id)
		if found != nil {
			ignoredSystems = append(ignoredSystems, *found)
		}
	}

	systems := games.AllSystems()

	// filter systems
	if len(filteredSystems) > 0 {
		systems = filteredSystems
	}

	// ignore systems
	if len(ignoredSystems) > 0 {
		var filtered []games.System
		for systemIndex := range systems {
			system := &systems[systemIndex]
			ignore := false
			for ignoredIndex := range ignoredSystems {
				ignored := &ignoredSystems[ignoredIndex]
				if system.Id == ignored.Id {
					ignore = true
					break
				}
			}
			if !ignore {
				filtered = append(filtered, *system)
			}
		}
		systems = filtered
	}

	results := games.GetSystemPaths(cfg, systems)
	if len(results) == 0 {
		_, _ = fmt.Println("No games folders found.")
		os.Exit(1)
	}

	// pick out the folders that actually have stuff in them
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

	if len(populated) == 0 {
		_, _ = fmt.Println("No games found.")
		return
	}

	if *noscan {
		for range maxPickAttempts {
			systemID, randomErr := utils.RandomElem(utils.MapKeys(populated))
			if randomErr != nil {
				continue
			}

			folder, randomErr := utils.RandomElem(populated[systemID])
			if randomErr != nil {
				continue
			}
			system, systemErr := games.GetSystem(systemID)
			if systemErr != nil {
				continue
			}
			game, gameErr := mister.TryPickRandomGame(system, folder)
			if gameErr != nil || game == "" {
				continue
			}

			_, _ = fmt.Printf("Launching %s: %s\n", system.Id, game)
			if launchErr := mister.LaunchGame(cfg, system, game); launchErr != nil {
				_, _ = fmt.Println(launchErr)
			}
			return
		}
	} else {
		for range maxPickAttempts {
			systemID, randomErr := utils.RandomElem(utils.MapKeys(populated))
			if randomErr != nil {
				continue
			}

			var files []string
			for _, path := range populated[systemID] {
				results, filesErr := games.GetFiles(systemID, path)
				if filesErr != nil {
					continue
				}
				files = append(files, results...)
			}
			if len(files) == 0 {
				continue
			}

			system, systemErr := games.GetSystem(systemID)
			if systemErr != nil {
				continue
			}
			game, gameErr := utils.RandomElem(files)
			if gameErr != nil {
				continue
			}

			_, _ = fmt.Printf("Launching %s: %s\n", system.Id, game)
			if launchErr := mister.LaunchGame(cfg, system, game); launchErr != nil {
				_, _ = fmt.Println(launchErr)
			}
			return
		}
	}

	_, _ = fmt.Println("No games found.")
}
