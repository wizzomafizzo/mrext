package main

import (
	"fmt"
	"log"
	"strings"

	gc "github.com/rthornton128/goncurses"

	"github.com/wizzomafizzo/mrext/pkg/config"
	"github.com/wizzomafizzo/mrext/pkg/curses"
	"github.com/wizzomafizzo/mrext/pkg/games"
	"github.com/wizzomafizzo/mrext/pkg/mister"
	"github.com/wizzomafizzo/mrext/pkg/txtindex"
	"github.com/wizzomafizzo/mrext/pkg/utils"
)

// TODO: list display window showing 2 values per row (left and right aligned)
// TODO: list display window with selected/deselected status per item
// TODO: small popup selection menu dialog

// Create a channel that will be used to pass the index around. This is so
// the index file can be loaded in the background on startup.
func newIndexChannel() chan txtindex.Index {
	ic := make(chan txtindex.Index, 1)
	go func() {
		index, err := txtindex.Open(config.SearchDbFile)
		if err != nil {
			log.Fatal(err)
		}
		ic <- index
	}()
	return ic
}

func getIndex(ic chan txtindex.Index) (txtindex.Index, chan txtindex.Index) {
	index := <-ic
	ic <- index
	return index, ic
}

func generateIndexWindow(stdscr *gc.Window) error {
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

	win.MovePrint(1, 2, "Finding games folders...")
	drawProgressBar(1, 100)
	win.NoutRefresh()
	gc.Update()

	systemPaths := make(map[string][]string)

	for _, path := range games.GetSystemPaths(games.AllSystems()) {
		systemPaths[path.System.Id] = append(systemPaths[path.System.Id], path.Path)
	}

	totalSteps := 0
	for _, systems := range systemPaths {
		totalSteps += len(systems)
	}
	totalSteps += 3
	currentStep := 2

	files, _ := games.GetAllFiles(systemPaths, func(system string, path string) {
		clearText()
		win.MovePrint(1, 2, fmt.Sprintf("Scanning %s: %s", system, path))
		drawProgressBar(currentStep, totalSteps)
		currentStep++
		win.NoutRefresh()
		gc.Update()
	})

	clearText()
	win.MovePrint(1, 2, "Generating index files...")
	drawProgressBar(currentStep, totalSteps)
	win.NoutRefresh()
	gc.Update()

	if err := txtindex.Generate(files, config.SearchDbFile); err != nil {
		log.Fatal(err)
	}

	return nil
}

func mainOptionsWindow(stdscr *gc.Window) error {
	options := [][2]string{
		{"Rescan games...", ""},
	}

	button, selected, err := curses.KeyValueListPicker(stdscr, curses.ListPickerOpts{
		Title:         "Options",
		Buttons:       []string{"Select", "Back"},
		DefaultButton: 0,
		ShowTotal:     false,
		Width:         70,
		Height:        18,
	}, options)

	if err != nil {
		return err
	}

	if button == 0 {
		switch selected {
		case 0:
			generateIndexWindow(stdscr)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func searchWindow(stdscr *gc.Window, ic chan txtindex.Index, query string) (err error) {
	stdscr.Erase()
	stdscr.NoutRefresh()
	gc.Update()

	searchTitle := "Search"
	searchButtons := []string{"Options", "Search", "Exit"}
	button, text, err := curses.OnScreenKeyboard(stdscr, searchTitle, searchButtons, query)
	if err != nil {
		log.Fatal(err)
	}

	if button == 0 {
		err = mainOptionsWindow(stdscr)
		if err != nil {
			return err
		}

		return searchWindow(stdscr, ic, text)
	} else if button == 1 {
		if len(text) == 0 {
			return searchWindow(stdscr, ic, "")
		}

		index, ic := getIndex(ic)
		if err := curses.InfoBox(stdscr, "", "Searching...", false, false); err != nil {
			log.Fatal(err)
		}

		results := index.SearchAllByWords(text)

		if len(results) == 0 {
			if err := curses.InfoBox(stdscr, "", "No results found.", false, true); err != nil {
				log.Fatal(err)
			}
			return searchWindow(stdscr, ic, text)
		}

		var names []string
		var items []txtindex.SearchResult
		for _, result := range results {
			display := fmt.Sprintf("[%s] %s", result.System, result.Name)
			if !utils.Contains(names, display) {
				names = append(names, display)
				items = append(items, result)
			}
		}

		stdscr.Erase()
		stdscr.NoutRefresh()
		gc.Update()

		button, selected, err := curses.ListPicker(stdscr, curses.ListPickerOpts{
			Title:         "Launch Game",
			Buttons:       []string{"PgUp", "PgDn", "Launch", "Options", "Cancel"},
			DefaultButton: 2,
			ShowTotal:     true,
			Width:         70,
			Height:        18,
		}, names)
		if err != nil {
			log.Fatal(err)
		}

		if button == 2 {
			game := items[selected]

			system, err := games.GetSystem(game.System)
			if err != nil {
				log.Fatal(err)
			}

			err = mister.LaunchGame(system, game.Path)
			if err != nil {
				log.Fatal(err)
			} else {
				return nil
			}
		}

		return searchWindow(stdscr, ic, text)
	} else {
		return nil
	}
}

func main() {
	stdscr, err := curses.Setup()
	if err != nil {
		log.Fatal(err)
	}
	defer gc.End()

	if !txtindex.Exists() {
		generateIndexWindow(stdscr)
		if err != nil {
			log.Fatal(err)
		}
	}

	ic := newIndexChannel()
	err = searchWindow(stdscr, ic, "")
	if err != nil {
		log.Fatal(err)
	}
}
