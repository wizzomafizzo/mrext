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

	"github.com/wizzomafizzo/mrext/pkg/config"
	"github.com/wizzomafizzo/mrext/pkg/games"
	"github.com/wizzomafizzo/mrext/pkg/gamesdb"
	"github.com/wizzomafizzo/mrext/pkg/mister"
	"github.com/wizzomafizzo/mrext/pkg/tui"
	"github.com/wizzomafizzo/mrext/pkg/utils"
)

const appName = "search"

func generateIndexWindow(cfg *config.UserConfig) error {
	err := tui.RunProgress("", tui.ProgressUpdate{
		Text:    "Finding games folders...",
		Current: 1,
		Total:   100,
	}, func(update func(tui.ProgressUpdate)) error {
		_, err := gamesdb.NewNamesIndex(cfg, games.AllSystems(), func(status gamesdb.IndexStatus) {
			systemName := status.SystemID
			if system, systemErr := games.GetSystem(status.SystemID); systemErr == nil {
				systemName = system.Name
			}
			text := fmt.Sprintf("Indexing %s...", systemName)
			switch status.Step {
			case 1:
				text = "Finding games folders..."
			case status.Total:
				text = "Writing database to disk..."
			}
			update(tui.ProgressUpdate{Text: text, Current: status.Step, Total: status.Total})
		})
		if err != nil {
			return fmt.Errorf("build game-name index: %w", err)
		}
		return nil
	})
	if err != nil {
		return fmt.Errorf("show index progress: %w", err)
	}
	return nil
}

func mainOptionsWindow(cfg *config.UserConfig) error {
	button, selected, err := tui.ListPicker(&tui.ListPickerOpts{
		Title:         "Options",
		Buttons:       []string{"Select", "Back"},
		DefaultButton: 0,
		ActionButton:  0,
		Width:         70,
		Height:        18,
	}, []string{"Update games database..."})
	if err != nil {
		return fmt.Errorf("show options: %w", err)
	}
	if button == 0 && selected == 0 {
		return generateIndexWindow(cfg)
	}
	return nil
}

func searchWindow(cfg *config.UserConfig, query string, launchGame bool) error {
	button, text, err := tui.OnScreenKeyboard("Search", []string{"Options", "Search", "Exit"}, query)
	if err != nil {
		return fmt.Errorf("show search keyboard: %w", err)
	}

	switch button {
	case 0:
		if err := mainOptionsWindow(cfg); err != nil {
			return err
		}
		return searchWindow(cfg, text, launchGame)
	case 1:
		if text == "" {
			return searchWindow(cfg, "", launchGame)
		}

		var results []gamesdb.SearchResult
		progressErr := tui.RunProgress(
			"",
			tui.ProgressUpdate{Text: "Searching..."},
			func(_ func(tui.ProgressUpdate)) error {
				var searchErr error
				results, searchErr = gamesdb.SearchNamesWords(games.AllSystems(), text)
				if searchErr != nil {
					return fmt.Errorf("search game names: %w", searchErr)
				}
				return nil
			},
		)
		if progressErr != nil {
			return fmt.Errorf("show search progress: %w", progressErr)
		}
		if len(results) == 0 {
			if infoErr := tui.InfoBox("", "No results found."); infoErr != nil {
				return fmt.Errorf("show no-results message: %w", infoErr)
			}
			return searchWindow(cfg, text, launchGame)
		}

		names := make([]string, 0, len(results))
		items := make([]gamesdb.SearchResult, 0, len(results))
		for _, result := range results {
			systemName := result.SystemID
			if system, systemErr := games.GetSystem(result.SystemID); systemErr == nil {
				systemName = system.Name
			}
			display := fmt.Sprintf("[%s] %s", systemName, result.Name)
			if !utils.Contains(names, display) {
				names = append(names, display)
				items = append(items, result)
			}
		}

		titleLabel := "Launch Game"
		launchLabel := "Launch"
		if !launchGame {
			titleLabel = "Pick Game"
			launchLabel = "Select"
		}
		button, selected, err := tui.ListPicker(&tui.ListPickerOpts{
			Title:         titleLabel,
			Buttons:       []string{"PgUp", "PgDn", launchLabel, "Cancel"},
			DefaultButton: 2,
			ActionButton:  2,
			ShowTotal:     true,
			Width:         70,
			Height:        18,
		}, names)
		if err != nil {
			return fmt.Errorf("show search results: %w", err)
		}
		if button != 2 || selected < 0 {
			return searchWindow(cfg, text, launchGame)
		}

		game := items[selected]
		if !launchGame {
			_, _ = fmt.Fprintln(os.Stderr, game.Path)
			return nil
		}
		system, err := games.GetSystem(game.SystemID)
		if err != nil {
			return fmt.Errorf("get selected system: %w", err)
		}
		if err := mister.LaunchGame(cfg, system, game.Path); err != nil {
			return fmt.Errorf("launch selected game: %w", err)
		}
		return nil
	default:
		return nil
	}
}

func fatal(err error) {
	_, _ = fmt.Fprintln(os.Stderr, err)
	os.Exit(1)
}

func main() {
	printPath := flag.Bool("print", false, "Print game path to stderr instead of launching the game")
	flag.Parse()

	cfg, err := config.LoadUserConfig(appName, &config.UserConfig{})
	if err != nil {
		fatal(err)
	}
	if !gamesdb.DBExists() {
		if indexErr := generateIndexWindow(cfg); indexErr != nil {
			fatal(indexErr)
		}
	}
	if searchErr := searchWindow(cfg, "", !*printPath); searchErr != nil {
		fatal(searchErr)
	}
}
