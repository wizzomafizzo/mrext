package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/wizzomafizzo/mrext/pkg/mister"

	gc "github.com/rthornton128/goncurses"

	"github.com/wizzomafizzo/mrext/pkg/config"
	"github.com/wizzomafizzo/mrext/pkg/curses"
	"github.com/wizzomafizzo/mrext/pkg/games"
	"github.com/wizzomafizzo/mrext/pkg/gamesdb"
	"github.com/wizzomafizzo/mrext/pkg/utils"
)

// TODO: list display window showing 2 values per row (left and right aligned)
// TODO: list display window with selected/deselected status per item
// TODO: small popup selection menu dialog

const appName = "search"

func generateIndexWindow(cfg *config.UserConfig, stdscr *gc.Window) error {
	win, err := curses.NewWindow(stdscr, 4, 75, "", -1)
	if err != nil {
		return err
	}
	defer win.Delete()

	_, width := win.MaxYX()

	drawProgressBar := func(current int, total int) {
		pct := int(float64(current) / float64(total) * 100)
		progressWidth := width - 4
		progressPct := int(float64(pct) / float64(100) * float64(progressWidth))
		if progressPct < 1 {
			progressPct = 1
		}
		for i := 0; i < progressPct; i++ {
			win.MoveAddChar(2, 2+i, gc.ACS_BLOCK)
		}
		win.NoutRefresh()
	}

	clearText := func() {
		win.MovePrint(1, 2, strings.Repeat(" ", width-4))
	}

	status := struct {
		Step        int
		Total       int
		SystemName  string
		DisplayText string
		Complete    bool
		Error       error
	}{
		Step:        1,
		Total:       100,
		DisplayText: "Finding games folders...",
	}

	go func() {
		_, err = gamesdb.NewNamesIndex(cfg, games.AllSystems(), func(is gamesdb.IndexStatus) {
			systemName := is.SystemId
			system, err := games.GetSystem(is.SystemId)
			if err == nil {
				systemName = system.Name
			}

			text := fmt.Sprintf("Indexing %s...", systemName)
			if is.Step == 1 {
				text = "Finding games folders..."
			} else if is.Step == is.Total {
				text = "Writing database to disk..."
			}

			status.Step = is.Step
			status.Total = is.Total
			status.SystemName = systemName
			status.DisplayText = text
		})

		status.Error = err
		status.Complete = true
	}()

	spinnerSeq := []string{"|", "/", "-", "\\"}
	spinnerCount := 0

	for {
		if status.Complete || status.Error != nil {
			break
		}

		clearText()

		spinnerCount++
		if spinnerCount == len(spinnerSeq) {
			spinnerCount = 0
		}

		win.MovePrint(1, width-3, spinnerSeq[spinnerCount])

		win.MovePrint(1, 2, status.DisplayText)
		drawProgressBar(status.Step, status.Total)

		win.NoutRefresh()
		_ = gc.Update()
		gc.Nap(100)
	}

	return status.Error
}

func mainOptionsWindow(cfg *config.UserConfig, stdscr *gc.Window) error {
	button, selected, err := curses.ListPicker(stdscr, curses.ListPickerOpts{
		Title:         "Options",
		Buttons:       []string{"Select", "Back"},
		DefaultButton: 0,
		ActionButton:  0,
		ShowTotal:     false,
		Width:         70,
		Height:        18,
	}, []string{"Update games database..."})

	if err != nil {
		return err
	}

	if button == 0 {
		switch selected {
		case 0:
			err := generateIndexWindow(cfg, stdscr)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func searchWindow(cfg *config.UserConfig, stdscr *gc.Window, query string, launchGame bool) (err error) {
	stdscr.Erase()
	stdscr.NoutRefresh()
	_ = gc.Update()

	searchTitle := "Search"
	searchButtons := []string{"Options", "Search", "Exit"}
	button, text, err := curses.OnScreenKeyboard(stdscr, searchTitle, searchButtons, query)
	if err != nil {
		log.Fatal(err)
	}

	if button == 0 {
		err = mainOptionsWindow(cfg, stdscr)
		if err != nil {
			return err
		}

		return searchWindow(cfg, stdscr, text, launchGame)
	} else if button == 1 {
		if len(text) == 0 {
			return searchWindow(cfg, stdscr, "", launchGame)
		}

		if err := curses.InfoBox(stdscr, "", "Searching...", false, false); err != nil {
			log.Fatal(err)
		}

		results, err := gamesdb.SearchNamesWords(games.AllSystems(), text)
		if err != nil {
			return err
		}

		if len(results) == 0 {
			if err := curses.InfoBox(stdscr, "", "No results found.", false, true); err != nil {
				log.Fatal(err)
			}
			return searchWindow(cfg, stdscr, text, launchGame)
		}

		var names []string
		var items []gamesdb.SearchResult
		for _, result := range results {
			systemName := result.SystemId
			system, err := games.GetSystem(result.SystemId)
			if err == nil {
				systemName = system.Name
			}

			display := fmt.Sprintf("[%s] %s", systemName, result.Name)
			if !utils.Contains(names, display) {
				names = append(names, display)
				items = append(items, result)
			}
		}

		stdscr.Erase()
		stdscr.NoutRefresh()
		_ = gc.Update()

		var titleLabel, launchLabel string

		if launchGame {
			titleLabel = "Launch Game"
			launchLabel = "Launch"
		} else {
			titleLabel = "Pick Game"
			launchLabel = "Select"
		}
		button, selected, err := curses.ListPicker(stdscr, curses.ListPickerOpts{
			Title:         titleLabel,
			Buttons:       []string{"PgUp", "PgDn", launchLabel, "Cancel"},
			DefaultButton: 2,
			ActionButton:  2,
			ShowTotal:     true,
			Width:         70,
			Height:        18,
		}, names)
		if err != nil {
			log.Fatal(err)
		}

		if button == 2 {
			game := items[selected]

			if launchGame {
				system, err := games.GetSystem(game.SystemId)
				if err != nil {
					log.Fatal(err)
				}

				err = mister.LaunchGame(cfg, *system, game.Path)
				if err != nil {
					log.Fatal(err)
				} else {
					return nil
				}
			} else {
				gc.End()
				fmt.Fprintln(os.Stderr, game.Path)
				os.Exit(0)
			}
		}

		return searchWindow(cfg, stdscr, text, launchGame)
	} else {
		return nil
	}
}

func main() {
	printPtr := flag.Bool("print", false, "Print game path to stderr instead of launching the game")
	flag.Parse()
	var launchGame bool = !*printPtr

	cfg, err := config.LoadUserConfig(appName, &config.UserConfig{})
	if err != nil {
		fmt.Println("Error loading config file:", err)
		os.Exit(1)
	}

	stdscr, err := curses.Setup()
	if err != nil {
		log.Fatal(err)
	}
	defer gc.End()

	if !gamesdb.DbExists() {
		err := generateIndexWindow(cfg, stdscr)
		if err != nil {
			log.Fatal(err)
		}
	}

	err = searchWindow(cfg, stdscr, "", launchGame)
	if err != nil {
		log.Fatal(err)
	}
}
