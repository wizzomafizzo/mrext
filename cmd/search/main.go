package main

import (
	"fmt"
	"log"

	s "strings"

	gc "github.com/rthornton128/goncurses"

	"github.com/wizzomafizzo/mrext/pkg/curses"
	"github.com/wizzomafizzo/mrext/pkg/games"
	"github.com/wizzomafizzo/mrext/pkg/mister"
	"github.com/wizzomafizzo/mrext/pkg/txtindex"
	"github.com/wizzomafizzo/mrext/pkg/utils"
)

func newIndexChannel() chan txtindex.Index {
	ic := make(chan txtindex.Index, 1)
	go func() {
		index, err := txtindex.Open(txtindex.GetIndexPath())
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
		win.MovePrint(1, 2, s.Repeat(" ", width-4))
	}

	win.MovePrint(1, 2, "Finding games folders...")
	drawProgressBar(1, 100)
	win.NoutRefresh()
	gc.Update()

	systemPaths := make(map[string][]string)
	for _, path := range games.GetSystemPaths(games.AllSystems()) {
		systemPaths[path.System.Id] = append(systemPaths[path.System.Id], path.Path)
	}
	totalSteps := len(systemPaths) + 3
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

	if err := txtindex.Generate(files, txtindex.GetIndexPath()); err != nil {
		log.Fatal(err)
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

		button, selected, err := curses.ListPicker(stdscr, "Launch Game", names, []string{"PgUp", "PgDn", "Launch", "Options", "Cancel"}, 2)
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
