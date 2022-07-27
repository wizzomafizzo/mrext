package main

import (
	"fmt"
	"log"
	"os"
	s "strings"
	"time"

	gc "github.com/rthornton128/goncurses"

	"github.com/wizzomafizzo/mrext/pkg/games"
	"github.com/wizzomafizzo/mrext/pkg/index"
)

func SetupCurses() (*gc.Window, error) {
	stdscr, err := gc.Init()
	if err != nil {
		return nil, err
	}

	gc.Echo(false)
	gc.CBreak(true)
	gc.Cursor(0)

	return stdscr, nil
}

func NewWindow(stdscr *gc.Window, height int, width int, title string, timeout int) (*gc.Window, error) {
	rows, cols := stdscr.MaxYX()
	y, x := (rows-height)/2, (cols-width)/2

	var win *gc.Window
	win, err := gc.NewWindow(height, width, y, x)
	if err != nil {
		return nil, err
	}
	win.Keypad(true)
	win.Timeout(timeout)

	win.Erase()
	win.NoutRefresh()

	win.Box(gc.ACS_VLINE, gc.ACS_HLINE)
	if len(title) > 0 {
		titleX := (width - len(title)) / 2
		win.MovePrint(0, titleX, title)
	}
	win.NoutRefresh()

	return win, nil
}

func newGenerateIndexWindow(stdscr *gc.Window) error {
	win, err := NewWindow(stdscr, 5, 75, "", -1)
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

	files := games.GetSystemFiles(func(system string) {
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

func DrawBox(win *gc.Window, y int, x int, height int, width int) {
	win.HLine(y, x+1, gc.ACS_HLINE, width-1)
	win.HLine(height, x+1, gc.ACS_HLINE, width-1)
	win.VLine(y+1, x, gc.ACS_VLINE, height-1)
	win.VLine(y+1, x+width-1, gc.ACS_VLINE, height-1)
	win.MoveAddChar(y, x, gc.ACS_ULCORNER)
	win.MoveAddChar(y, x+width-1, gc.ACS_URCORNER)
	win.MoveAddChar(height, x, gc.ACS_LLCORNER)
	win.MoveAddChar(height, x+width-1, gc.ACS_LRCORNER)
	win.NoutRefresh()
}

func OnScreenKeyboard(stdscr *gc.Window) {
	win, err := NewWindow(stdscr, 16, 63, "Search", -1)
	if err != nil {
		log.Fatal(err)
	}
	defer win.Delete()

	height, width := win.MaxYX()
	DrawBox(win, 1, 1, 3, width-2)
	gc.Cursor(1)

	win.HLine(height-3, 1, gc.ACS_HLINE, width-2)
	win.MoveAddChar(height-3, 0, gc.ACS_LTEE)
	win.MoveAddChar(height-3, width-1, gc.ACS_RTEE)

	keys := [4][10]gc.Char{
		{'1', '2', '3', '4', '5', '6', '7', '8', '9', '0'},
		{'Q', 'W', 'E', 'R', 'T', 'Y', 'U', 'I', 'O', 'P'},
		{'A', 'S', 'D', 'F', 'G', 'H', 'J', 'K', 'L', '<'},
		{' ', 'Z', 'X', 'C', 'V', 'B', 'N', 'M', '_', ' '},
	}

	for y, row := range keys {
		for x, key := range row {
			win.Move(5+(y*2), 2+(x*6))
			switch key {
			case ' ':
				continue
			case '<':
				win.AddChar('[')
				win.AddChar(' ')
				win.AddChar(gc.ACS_LARROW)
				win.AddChar(' ')
				win.AddChar(']')
			case '_':
				win.AddChar('[')
				win.AddChar('S')
				win.AddChar('P')
				win.AddChar('C')
				win.AddChar(']')
			default:
				win.AddChar('[')
				win.AddChar(' ')
				win.AddChar(key)
				win.AddChar(' ')
				win.AddChar(']')
			}
		}
	}

	win.Move(2, 2)

	win.NoutRefresh()
	gc.Update()

	ch := win.GetChar()
	for ch != 'q' {
		switch ch {
		case gc.KEY_DOWN, gc.KEY_TAB:
		case gc.KEY_UP:
		case gc.KEY_BACKSPACE:
		default:
		}
		ch = win.GetChar()

		gc.Update()
	}
}

func main() {
	stdscr, err := SetupCurses()
	if err != nil {
		log.Fatal(err)
	}
	defer gc.End()

	if _, err := os.Stat(index.GetDbPath()); os.IsNotExist(err) {
		if err := newGenerateIndexWindow(stdscr); err != nil {
			log.Fatal(err)
		}
	} else if err != nil {
		log.Fatal(err)
	}

	OnScreenKeyboard(stdscr)
}
