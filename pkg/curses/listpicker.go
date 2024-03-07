package curses

import (
	"fmt"
	s "strings"

	gc "github.com/rthornton128/goncurses"
	"github.com/wizzomafizzo/mrext/pkg/utils"
)

type ListPickerOpts struct {
	Title         string
	Buttons       []string
	DefaultButton int
	ActionButton  int
	ShowTotal     bool
	Width         int
	Height        int
}

func ListPicker(stdscr *gc.Window, opts ListPickerOpts, items []string) (int, int, error) {
	selectedItem := 0
	selectedButton := opts.DefaultButton

	pgUpName := "PgUp"
	pgDownName := "PgDn"

	height := opts.Height
	width := opts.Width

	viewStart := 0
	viewHeight := height - 4
	viewWidth := width - 4
	pgAmount := viewHeight - 1

	pageUp := func() {
		if viewStart == 0 {
			selectedItem = 0
		} else if (viewStart - pgAmount) < 0 {
			viewStart = 0
			selectedItem = 0
		} else {
			viewStart -= pgAmount
			selectedItem = viewStart + pgAmount
		}
	}

	pageDown := func() {
		if viewStart == len(items)-viewHeight {
			selectedItem = len(items) - 1
		} else if (viewStart + pgAmount) > len(items)-viewHeight {
			viewStart = len(items) - viewHeight
			selectedItem = len(items) - 1
		} else {
			viewStart += pgAmount
			selectedItem = viewStart
		}
	}

	win, err := NewWindow(stdscr, height, width, opts.Title, -1)
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
		// FIXME: not quite working
		// var gripHeight int
		// var gripOffset int
		// scrollHeight := viewHeight - 2

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
		DrawActionButtons(win, opts.Buttons, selectedButton, 4)
		win.NoutRefresh()

		// location indicators
		if opts.ShowTotal {
			totalStatus := fmt.Sprintf("%*d/%d", len(fmt.Sprint(len(items))), selectedItem+1, len(items))
			win.MovePrint(0, 2, totalStatus)
		}

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

		win.Move(viewStart+selectedItem+1, width-3)

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
				selectedButton = len(opts.Buttons) - 1
			}
		case gc.KEY_RIGHT:
			if selectedButton < len(opts.Buttons)-1 {
				selectedButton++
			} else {
				selectedButton = 0
			}
		case gc.KEY_PAGEUP:
			pageUp()
		case gc.KEY_PAGEDOWN:
			pageDown()
		case gc.KEY_ENTER, 10, 13:
			if selectedButton == opts.ActionButton {
				return selectedButton, selectedItem, nil
			} else if opts.Buttons[selectedButton] == pgUpName {
				pageUp()
			} else if opts.Buttons[selectedButton] == pgDownName {
				pageDown()
			} else {
				return selectedButton, -1, nil
			}
		}
	}

	return -1, -1, nil
}

func KeyValueListPicker(stdscr *gc.Window, opts ListPickerOpts, items [][2]string) (int, int, error) {
	strItems := make([]string, len(items))
	for i, item := range items {
		strItems[i] = item[0] + " " + item[1]
	}

	return ListPicker(stdscr, opts, strItems)
}
