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
	"github.com/wizzomafizzo/mrext/pkg/version"
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

// findSyncFiles reports unreadable sync files through report rather than
// printing them, so the same code serves the console log and the progress
// modal without writing underneath a drawn screen.
func findSyncFiles(report func(string)) []syncFile {
	menuFolders := mister.GetMenuFolders(config.SdFolder)
	menuFolders = append(menuFolders, config.SdFolder)
	syncFiles := getSyncFiles(menuFolders)
	var syncs []syncFile

	for _, path := range syncFiles {
		sf, err := readSyncFile(path)
		if err != nil {
			report(fmt.Sprintf("Error reading %s: %s", path, err))
			continue
		}
		syncs = append(syncs, sf)
	}

	return syncs
}

// syncOutcome is what one sync file produced, for the summary page.
type syncOutcome struct {
	name    string
	found   int
	missing int
	failed  int
}

// runSync performs the whole sync, reporting each step as a complete line.
// The console path prints those lines exactly as it always did; the TUI path
// shows the latest one in a progress modal and the outcomes on a summary page.
func runSync(cfg *config.UserConfig, report func(string)) ([]syncOutcome, error) {
	report("Searching for sync files...")
	syncs := findSyncFiles(report)
	if len(syncs) == 0 {
		return nil, errors.New("no sync files found")
	}
	report(fmt.Sprintf("Found %d sync file(s).", len(syncs)))

	report("Checking for updates...")
	for i := range syncs {
		sync := &syncs[i]
		newSync, updated, changeErr := checkForChanges(sync)
		switch {
		case changeErr != nil:
			report(fmt.Sprintf("%d/%d: %s... error: %s", i+1, len(syncs), sync.name, changeErr))
		case updated:
			syncs[i] = newSync
			report(fmt.Sprintf("%d/%d: %s... updated", i+1, len(syncs), sync.name))
		default:
			report(fmt.Sprintf("%d/%d: %s... no update", i+1, len(syncs), sync.name))
		}
	}

	report("Building games index...")
	if err := makeIndex(cfg, syncs); err != nil {
		return nil, fmt.Errorf("error generating index: %w", err)
	}
	report("Building games index... done")

	outcomes := make([]syncOutcome, 0, len(syncs))
	for syncIndex := range syncs {
		sync := &syncs[syncIndex]
		report("---")
		report("Name:    " + sync.name)
		report("Author:  " + sync.author)
		report("URL:     " + sync.url)
		report(fmt.Sprintf("Updated: %s", sync.updated))
		report("Folder:  " + sync.folder)
		report("Games:")

		// #nosec G301 -- generated launcher directory must remain world-readable.
		if err := os.MkdirAll(sync.folder, 0o755); err != nil {
			return outcomes, fmt.Errorf("error creating folder: %w", err)
		}

		outcome := syncOutcome{name: sync.name}
		for gameIndex := range sync.games {
			game := &sync.games[gameIndex]
			file, found, err := tryLinkGame(cfg, sync, game)
			switch {
			case err != nil:
				outcome.failed++
				report("- " + game.name + "... error: " + err.Error())
			case found:
				outcome.found++
				report("- " + game.name + "... found " + file)
			default:
				outcome.missing++
				report("- " + game.name + "... not found")
			}
		}
		outcomes = append(outcomes, outcome)
	}
	return outcomes, nil
}

func main() {
	update := flag.Bool("update", false, "find, update and link all sync files on system")
	verbose := flag.Bool("verbose", false, "print status information during update")
	test := flag.String("test", "", "report if specified sync file is valid and display match results")
	showVersion := flag.Bool("version", false, "print the version and exit")
	flag.Parse()
	if *showVersion {
		_, _ = fmt.Printf("%s %s\n", "launchsync", version.String())
		return
	}

	cfg, err := config.LoadUserConfig(appName, &config.UserConfig{
		TUI: config.TUIConfig{Theme: "default", Mouse: true, CRTMode: true},
	})
	if err != nil {
		_, _ = fmt.Println("Error loading config file:", err)
		os.Exit(1)
	}

	if *test != "" {
		testSyncFile(cfg, *test)
		return
	}

	// -update is how a script or a startup hook runs this, so it keeps the
	// console log and never draws a screen. A plain run from the Scripts menu
	// gets the shared interface.
	if *update {
		runConsole(cfg, *verbose)
		return
	}
	started, screenErr := showSyncScreen(cfg)
	if screenErr == nil {
		return
	}
	if started {
		// A screen came up and the sync ran inside it, so this is a real
		// failure. Re-running here would rewrite every shortcut again.
		_, _ = fmt.Fprintln(os.Stderr, screenErr)
		os.Exit(1)
	}
	// No screen was ever obtained and nothing ran, so this is a headless
	// invocation: cron, ssh without a tty, a wrapper script. Behave the way a
	// bare "launchsync" always did and print the log.
	runConsole(cfg, true)
}

// runConsole keeps the original non-interactive behaviour: the same lines, in
// the same order, and exit 1 on failure.
func runConsole(cfg *config.UserConfig, printLines bool) {
	report := func(line string) {
		if printLines {
			_, _ = fmt.Println(line)
		}
	}
	if _, err := runSync(cfg, report); err != nil {
		if printLines {
			_, _ = fmt.Println(err)
		}
		os.Exit(1)
	}
}
