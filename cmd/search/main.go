package main

import (
	"flag"
	"fmt"
	"log"
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
	return tui.RunProgress("", tui.ProgressUpdate{
		Text:    "Finding games folders...",
		Current: 1,
		Total:   100,
	}, func(update func(tui.ProgressUpdate)) error {
		_, err := gamesdb.NewNamesIndex(cfg, games.AllSystems(), func(status gamesdb.IndexStatus) {
			systemName := status.SystemId
			if system, systemErr := games.GetSystem(status.SystemId); systemErr == nil {
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
		return err
	})
}

func mainOptionsWindow(cfg *config.UserConfig) error {
	button, selected, err := tui.ListPicker(tui.ListPickerOpts{
		Title:         "Options",
		Buttons:       []string{"Select", "Back"},
		DefaultButton: 0,
		ActionButton:  0,
		Width:         70,
		Height:        18,
	}, []string{"Update games database..."})
	if err != nil {
		return err
	}
	if button == 0 && selected == 0 {
		return generateIndexWindow(cfg)
	}
	return nil
}

func searchWindow(cfg *config.UserConfig, query string, launchGame bool) error {
	button, text, err := tui.OnScreenKeyboard("Search", []string{"Options", "Search", "Exit"}, query)
	if err != nil {
		return err
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
		err := tui.RunProgress("", tui.ProgressUpdate{Text: "Searching..."}, func(_ func(tui.ProgressUpdate)) error {
			var searchErr error
			results, searchErr = gamesdb.SearchNamesWords(games.AllSystems(), text)
			return searchErr
		})
		if err != nil {
			return err
		}
		if len(results) == 0 {
			if err := tui.InfoBox("", "No results found."); err != nil {
				return err
			}
			return searchWindow(cfg, text, launchGame)
		}

		names := make([]string, 0, len(results))
		items := make([]gamesdb.SearchResult, 0, len(results))
		for _, result := range results {
			systemName := result.SystemId
			if system, systemErr := games.GetSystem(result.SystemId); systemErr == nil {
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
		button, selected, err := tui.ListPicker(tui.ListPickerOpts{
			Title:         titleLabel,
			Buttons:       []string{"PgUp", "PgDn", launchLabel, "Cancel"},
			DefaultButton: 2,
			ActionButton:  2,
			ShowTotal:     true,
			Width:         70,
			Height:        18,
		}, names)
		if err != nil {
			return err
		}
		if button != 2 || selected < 0 {
			return searchWindow(cfg, text, launchGame)
		}

		game := items[selected]
		if !launchGame {
			fmt.Fprintln(os.Stderr, game.Path)
			return nil
		}
		system, err := games.GetSystem(game.SystemId)
		if err != nil {
			return err
		}
		return mister.LaunchGame(cfg, *system, game.Path)
	default:
		return nil
	}
}

func main() {
	printPath := flag.Bool("print", false, "Print game path to stderr instead of launching the game")
	flag.Parse()

	cfg, err := config.LoadUserConfig(appName, &config.UserConfig{})
	if err != nil {
		log.Fatal(err)
	}
	if !gamesdb.DbExists() {
		if err := generateIndexWindow(cfg); err != nil {
			log.Fatal(err)
		}
	}
	if err := searchWindow(cfg, "", !*printPath); err != nil {
		log.Fatal(err)
	}
}
