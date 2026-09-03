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
	"flag"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/wizzomafizzo/mrext/pkg/config"
	"github.com/wizzomafizzo/mrext/pkg/games"
	"github.com/wizzomafizzo/mrext/pkg/gamesdb"
	"github.com/wizzomafizzo/mrext/pkg/mister"
)

// TODO: handle filename being too long (255 chars)
// TODO: add system id to mgl name if many systems, and config option
// TODO: mention about shortcuts ordering in mister menu

const appName = "launchsync"

func testSyncFile(cfg *config.UserConfig, path string) {
	sf, err := readSyncFile(path)
	if err != nil {
		_, _ = fmt.Printf("Error reading %s: %s\n", path, err)
		os.Exit(1)
	}

	_, _ = fmt.Printf("Name:    %s\n", sf.name)
	_, _ = fmt.Printf("Author:  %s\n", sf.author)
	_, _ = fmt.Printf("URL:     %s\n", sf.url)
	_, _ = fmt.Printf("Updated: %s\n", sf.updated)
	_, _ = fmt.Printf("Folder:  %s\n", sf.folder)
	_, _ = fmt.Printf("Games:   %d\n", len(sf.games))
	_, _ = fmt.Println("---")

	if sf.url != "" {
		_, _ = fmt.Print("Testing URL... ")

		req, requestErr := http.NewRequestWithContext(context.Background(), http.MethodGet, sf.url, http.NoBody)
		if requestErr != nil {
			_, _ = fmt.Printf("error: %s\n", requestErr)
		} else {
			client := &http.Client{Timeout: 30 * time.Second}
			resp, responseErr := client.Do(req)
			switch {
			case responseErr != nil:
				_, _ = fmt.Printf("error: %s\n", responseErr)
			case resp.StatusCode != http.StatusOK:
				_ = resp.Body.Close()
				_, _ = fmt.Printf("bad response: %s\n", resp.Status)
			default:
				_ = resp.Body.Close()
				_, _ = fmt.Println("tested OK")
			}
		}
	}

	if len(sf.games) == 0 {
		_, _ = fmt.Println("---")
		_, _ = fmt.Println("No games")
		return
	}

	_, _ = fmt.Print("Building games index... ")
	err = makeIndex(cfg, []syncFile{sf})
	if err != nil {
		_, _ = fmt.Printf("error generating index: %s\n", err)
		os.Exit(1)
	}
	_, _ = fmt.Println("done")

	for _, game := range sf.games {
		_, _ = fmt.Println("---")
		_, _ = fmt.Printf("Game:    %s\n", game.name)
		_, _ = fmt.Printf("System:  %s\n", game.system.Id)

		var fn string
		if game.system.Id == "Arcade" {
			fn = game.name + ".mra"
		} else {
			fn = game.name + ".mgl"
		}
		_, _ = fmt.Printf("Path:    %s\n", filepath.Join(sf.folder, game.folder, fn))

		_, _ = fmt.Printf("Matches: %d\n", len(game.matches))

		for _, match := range game.matches {
			_, _ = fmt.Printf("- %s\n", match[4:])
			results, searchErr := gamesdb.SearchNamesRegexp([]games.System{*game.system}, match)
			if searchErr != nil {
				_, _ = fmt.Printf("  error: %s\n", searchErr)
				continue
			}
			for i := 0; i < 5 && i < len(results); i++ {
				if i == 0 {
					_, _ = fmt.Printf(" *%s\n", results[i].Path)
				} else {
					_, _ = fmt.Printf("  %s\n", results[i].Path)
				}
			}
		}
	}
}

func findSyncFiles(verbose, update *bool) []syncFile {
	menuFolders := mister.GetMenuFolders(config.SdFolder)
	menuFolders = append(menuFolders, config.SdFolder)
	syncFiles := getSyncFiles(menuFolders)
	var syncs []syncFile

	for _, path := range syncFiles {
		sf, err := readSyncFile(path)
		if err != nil {
			if *verbose || !*update {
				_, _ = fmt.Printf("Error reading %s: %s\n", path, err)
			}
			continue
		}
		syncs = append(syncs, sf)
	}

	return syncs
}

func main() {
	update := flag.Bool("update", false, "find, update and link all sync files on system")
	verbose := flag.Bool("verbose", false, "print status information during update")
	test := flag.String("test", "", "report if specified sync file is valid and display match results")
	flag.Parse()

	cfg, err := config.LoadUserConfig(appName, &config.UserConfig{})
	if err != nil {
		_, _ = fmt.Println("Error loading config file:", err)
		os.Exit(1)
	}

	if *test != "" {
		testSyncFile(cfg, *test)
		return
	}

	if *verbose || !*update {
		_, _ = fmt.Print("Searching for sync files... ")
	}
	syncs := findSyncFiles(verbose, update)

	if len(syncs) == 0 {
		if *verbose || !*update {
			_, _ = fmt.Println("no sync files found")
		}
		os.Exit(1)
	}
	if *verbose || !*update {
		_, _ = fmt.Printf("found %d\n", len(syncs))
	}

	if *verbose || !*update {
		_, _ = fmt.Println("Checking for updates...")
	}
	for i := range syncs {
		sync := &syncs[i]
		if *verbose || !*update {
			_, _ = fmt.Printf("%d/%d: %s... ", i+1, len(syncs), sync.name)
		}
		newSync, updated, changeErr := checkForChanges(sync)
		switch {
		case changeErr != nil:
			if *verbose || !*update {
				_, _ = fmt.Printf("error: %s\n", changeErr)
			}
		case updated:
			syncs[i] = newSync
			if *verbose || !*update {
				_, _ = fmt.Println("updated")
			}
		case *verbose || !*update:
			_, _ = fmt.Println("no update")
		}
	}

	if *verbose || !*update {
		_, _ = fmt.Print("Building games index... ")
	}
	err = makeIndex(cfg, syncs)
	if err != nil {
		if *verbose || !*update {
			_, _ = fmt.Printf("error generating index: %s\n", err)
		}
		os.Exit(1)
	}
	if *verbose || !*update {
		_, _ = fmt.Println("done")
	}

	for syncIndex := range syncs {
		sync := &syncs[syncIndex]
		if *verbose || !*update {
			_, _ = fmt.Println("---")
			_, _ = fmt.Printf("Name:    %s\n", sync.name)
			_, _ = fmt.Printf("Author:  %s\n", sync.author)
			_, _ = fmt.Printf("URL:     %s\n", sync.url)
			_, _ = fmt.Printf("Updated: %s\n", sync.updated)
			_, _ = fmt.Printf("Folder:  %s\n", sync.folder)
			_, _ = fmt.Println("Games:")
		}

		// #nosec G301 -- generated launcher directory must remain world-readable.
		err := os.MkdirAll(sync.folder, 0o755)
		if err != nil {
			if *verbose || !*update {
				_, _ = fmt.Printf("error creating folder: %s\n", err)
			}
			os.Exit(1)
		}

		for gameIndex := range sync.games {
			game := &sync.games[gameIndex]
			if *verbose || !*update {
				_, _ = fmt.Print("- " + game.name + "... ")
			}
			file, found, err := tryLinkGame(cfg, sync, game)
			if *verbose || !*update {
				switch {
				case err != nil:
					_, _ = fmt.Printf("error: %s\n", err)
				case found:
					_, _ = fmt.Printf("found %s\n", file)
				default:
					_, _ = fmt.Println("not found")
				}
			}
		}
	}
}
