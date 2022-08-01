package main

import (
	"fmt"
	"log"
	"os"
	s "strings"
	"time"

	gc "github.com/rthornton128/goncurses"

	"github.com/wizzomafizzo/mrext/pkg/curses"
	"github.com/wizzomafizzo/mrext/pkg/games"
	index "github.com/wizzomafizzo/mrext/pkg/sqlindex"
)

func generateIndexWindow(stdscr *gc.Window) error {
	win, err := curses.NewWindow(stdscr, 5, 75, "", -1)
	if err != nil {
		return err
	}
	defer win.Delete()

	scanningGamesText := "Scanning games folders..."
	_, width := win.MaxYX()
	scanningGamesClear := s.Repeat(" ", width-len(scanningGamesText)-4)
	win.MovePrint(1, 2, scanningGamesText)
	win.NoutRefresh()
	gc.Update()

	systemPaths := games.GetSystemPaths()

	files := games.GetSystemFiles(systemPaths, func(system string, path string) {
		win.MovePrint(1, 3+len(scanningGamesText), scanningGamesClear)
		win.MovePrint(1, 3+len(scanningGamesText), system)
		win.NoutRefresh()
		gc.Update()
	})

	win.MovePrint(1, 3+len(scanningGamesText), scanningGamesClear)
	win.MovePrint(1, 3+len(scanningGamesText), fmt.Sprintf("Done! (%d games)", len(files)))

	generatingIndexText := "Generating index..."
	win.MovePrint(2, 2, generatingIndexText)
	win.NoutRefresh()
	gc.Update()

	start := time.Now()
	addedPct := 0
	win.MovePrint(2, 3+len(generatingIndexText), fmt.Sprintf("%d%%", addedPct))
	win.NoutRefresh()
	gc.Update()

	if err := index.Generate(files, func(count int) {
		nextAddedPct := int(float64(count) / float64(len(files)) * 100)
		if nextAddedPct != addedPct {
			addedPct = nextAddedPct
			win.MovePrint(2, 3+len(generatingIndexText), fmt.Sprintf("%d%%", addedPct))
			progressWidth := width - 4
			progressPct := int(float64(addedPct) / float64(100) * float64(progressWidth))
			for i := 0; i <= progressPct; i++ {
				win.MoveAddChar(3, 2+i, gc.ACS_BLOCK)
			}
			win.NoutRefresh()
			gc.Update()
		}
	}); err != nil {
		log.Fatal(err)
	}

	win.MovePrint(2, 3+len(generatingIndexText), fmt.Sprintf("Done! (took %d seconds)", int(time.Since(start).Seconds())))
	win.NoutRefresh()
	gc.Update()

	win.GetChar()

	return nil
}

func searchWindow(stdscr *gc.Window, query string) (err error) {
	searchTitle := "Search"
	searchButtons := []string{"Advanced", "Search", "Exit"}
	button, text, err := curses.OnScreenKeyboard(stdscr, searchTitle, searchButtons, query)
	if err != nil {
		panic(err)
	}

	if button == 0 {
		return searchWindow(stdscr, text)
	} else if button == 1 {
		results, err := index.SearchGames(text)
		if err != nil {
			return err
		}

		return searchWindow(stdscr, fmt.Sprintf("%d", len(results)))
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

	if _, err := os.Stat(index.GetDbPath()); os.IsNotExist(err) {
		if err := generateIndexWindow(stdscr); err != nil {
			log.Fatal(err)
		}
		stdscr.Clear()
	} else if err != nil {
		log.Fatal(err)
	}

	err = searchWindow(stdscr, "")
	if err != nil {
		log.Fatal(err)
	}
}
