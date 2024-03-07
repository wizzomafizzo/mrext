package curses

import (
	s "strings"

	gc "github.com/rthornton128/goncurses"
)

type Coords struct {
	Y int
	X int
}

type SetupWindowError struct {
	Ctx error
}

func (e *SetupWindowError) Error() string {
	return e.Ctx.Error()
}

func Setup() (*gc.Window, error) {
	stdscr, err := gc.Init()
	if err != nil {
		return nil, err
	}

	gc.Echo(false)
	gc.CBreak(true)
	gc.Cursor(0)

	gc.StartColor()
	gc.InitPair(1, gc.C_BLACK, gc.C_WHITE)

	return stdscr, nil
}

func NewWindow(stdscr *gc.Window, height int, width int, title string, timeout int) (*gc.Window, error) {
	rows, cols := stdscr.MaxYX()
	y, x := (rows-height)/2, (cols-width)/2

	var win *gc.Window
	win, err := gc.NewWindow(height, width, y, x)
	if err != nil {
		return nil, &SetupWindowError{Ctx: err}
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

func DrawActionButtons(win *gc.Window, buttons []string, selected int, padding int) {
	height, width := win.MaxYX()

	win.HLine(height-3, 1, gc.ACS_HLINE, width-2)
	win.MoveAddChar(height-3, 0, gc.ACS_LTEE)
	win.MoveAddChar(height-3, width-1, gc.ACS_RTEE)

	maxButtonTextLen := 0
	for _, button := range buttons {
		if len(button) > maxButtonTextLen {
			maxButtonTextLen = len(button)
		}
	}

	buttonPadding := padding
	buttonWidth := maxButtonTextLen + 2
	totalButtonWidth := buttonWidth*len(buttons) + ((len(buttons) - 1) * buttonPadding)
	buttonsLeftMargin := (width - totalButtonWidth) / 2

	// buttonSection := width / (len(buttons) + 1)
	for i, button := range buttons {
		if i == selected {
			win.ColorOn(1)
		}
		textPadding := (maxButtonTextLen - len(button)) / 2
		buttonText := "<" + s.Repeat(" ", textPadding) + button + s.Repeat(" ", textPadding) + ">"
		win.MovePrint(height-2, buttonsLeftMargin+(i*(buttonWidth+buttonPadding)), buttonText)
		win.ColorOff(1)
	}

	win.NoutRefresh()
}

func InfoBox(stdscr *gc.Window, title string, text string, clear bool, ok bool) error {
	if clear {
		stdscr.Erase()
		stdscr.NoutRefresh()
		gc.Update()
	}

	height := 3
	// if ok {
	// 	height = 5
	// }

	win, err := NewWindow(stdscr, height, len(text)+4, title, -1)
	if err != nil {
		return err
	}
	defer win.Delete()

	gc.Cursor(0)

	win.MovePrint(1, 2, text)

	// if ok {
	// 	DrawActionButtons(win, []string{"OK"}, 0)
	// }

	win.NoutRefresh()
	gc.Update()

	if ok {
		win.GetChar()
	}

	return nil
}
