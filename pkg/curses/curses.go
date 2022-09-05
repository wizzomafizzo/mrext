package curses

import (
	"fmt"
	s "strings"

	gc "github.com/rthornton128/goncurses"
	"github.com/wizzomafizzo/mrext/pkg/utils"
)

type Coords struct {
	Y int
	X int
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

func ListPicker(stdscr *gc.Window, title string, items []string, buttons []string, defaultButton int) (int, int, error) {
	selectedItem := 0
	selectedButton := defaultButton

	height := 18
	width := 70

	viewStart := 0
	viewHeight := height - 4
	viewWidth := width - 4

	win, err := NewWindow(stdscr, height, width, title, -1)
	if err != nil {
		return -1, -1, err
	}
	defer win.Delete()

	var ch gc.Key

	for ch != gc.KEY_ESC {
		// list items
		max := utils.Min([]int{len(items), viewHeight})

		for i := 0; i < max; i++ {
			var display string

			item := items[viewStart+i]

			if len(item) > viewWidth {
				display = item[:width-(width-viewWidth)-3] + "..."
			} else {
				display = item
			}

			if viewStart+i == selectedItem {
				win.ColorOn(1)
			}

			win.MovePrint(i+1, 2, s.Repeat(" ", viewWidth))
			win.MovePrint(i+1, 2, display)
			win.ColorOff(1)
		}

		// scroll bar
		// var gripHeight int
		// var gripOffset int
		// scrollHeight := viewHeight - 2

		// FIXME: not quite working
		// if len(items) <= scrollHeight {
		// 	gripHeight = scrollHeight
		// } else {
		// 	gripHeight = int(m.Ceil((float64(scrollHeight) / float64(len(items))) * float64(scrollHeight)))
		// }

		// if gripHeight >= scrollHeight {
		// 	gripOffset = 0
		// } else {
		// 	gripOffset = int(m.Floor(float64(viewStart) * float64(scrollHeight) / float64(len(items))))
		// }

		// for i := 0; i < scrollHeight; i++ {
		// 	if i >= gripOffset && i < gripOffset+gripHeight {
		// 		win.ColorOn(1)
		// 		win.MoveAddChar(i+2, width-2, ' ')
		// 	} else {
		// 		win.MoveAddChar(i+2, width-2, ' ')
		// 	}
		// 	win.ColorOff(1)
		// }

		// win.MoveAddChar(1, width-2, ' ')
		// if viewStart > 0 {
		// 	win.ColorOn(1)
		// 	win.MoveAddChar(1, width-2, gc.ACS_UARROW)
		// 	win.ColorOff(1)
		// }

		// win.MoveAddChar(height-4, width-2, ' ')
		// if viewStart+viewHeight < len(items) {
		// 	win.ColorOn(1)
		// 	win.MoveAddChar(height-4, width-2, gc.ACS_DARROW)
		// 	win.ColorOff(1)
		// }

		// buttons
		DrawActionButtons(win, buttons, selectedButton, 4)
		win.NoutRefresh()

		// location indicators
		totalStatus := fmt.Sprintf("%*d/%d", len(fmt.Sprint(len(items))), selectedItem+1, len(items))
		if err != nil {
			return -1, -1, err
		}
		win.MovePrint(0, 2, totalStatus)

		if viewStart > 0 {
			win.MoveAddChar(0, width-3, gc.ACS_UARROW)
		} else {
			win.MoveAddChar(0, width-3, gc.ACS_HLINE)
		}

		if viewStart+viewHeight < len(items) {
			win.MoveAddChar(height-3, width-3, gc.ACS_DARROW)
		} else {
			win.MoveAddChar(height-3, width-3, gc.ACS_HLINE)
		}

		win.NoutRefresh()
		gc.Update()

		ch = win.GetChar()

		switch ch {
		case gc.KEY_DOWN:
			if selectedItem < len(items)-1 {
				selectedItem++
				if selectedItem >= viewStart+viewHeight && viewStart+viewHeight < len(items) {
					viewStart++
				}
			}
		case gc.KEY_UP:
			if selectedItem > 0 {
				selectedItem--
				if selectedItem < viewStart && viewStart > 0 {
					viewStart--
				}
			}
		case gc.KEY_LEFT:
			if selectedButton > 0 {
				selectedButton--
			} else {
				selectedButton = len(buttons) - 1
			}
		case gc.KEY_RIGHT:
			if selectedButton < len(buttons)-1 {
				selectedButton++
			} else {
				selectedButton = 0
			}
		case gc.KEY_ENTER, 10, 13:
			if selectedButton == 1 {
				return selectedButton, selectedItem, nil
			} else {
				return selectedButton, -1, nil
			}
		}
	}

	return -1, -1, nil
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
