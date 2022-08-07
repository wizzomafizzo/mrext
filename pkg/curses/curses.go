package curses

import (
	"fmt"
	m "math"
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

func DrawActionButtons(win *gc.Window, buttons []string, selected int) {
	height, width := win.MaxYX()

	win.HLine(height-3, 1, gc.ACS_HLINE, width-2)
	win.MoveAddChar(height-3, 0, gc.ACS_LTEE)
	win.MoveAddChar(height-3, width-1, gc.ACS_RTEE)

	maxButtonWidth := 0
	for _, button := range buttons {
		if len(button) > maxButtonWidth {
			maxButtonWidth = len(button)
		}
	}

	buttonSection := width / (len(buttons) + 1)
	for i, button := range buttons {
		if i == selected {
			win.ColorOn(1)
		}
		textPadding := (maxButtonWidth - len(button)) / 2
		buttonText := "<" + s.Repeat(" ", textPadding) + button + s.Repeat(" ", textPadding) + ">"
		win.MovePrint(height-2, (i+1)*buttonSection-(maxButtonWidth/2), buttonText)
		win.ColorOff(1)
	}

	win.NoutRefresh()
}

func OnScreenKeyboard(stdscr *gc.Window, title string, buttons []string, defaultText string) (int, string, error) {
	win, err := NewWindow(stdscr, 16, 63, title, -1)
	if err != nil {
		return 0, "", err
	}
	defer win.Delete()

	_, width := win.MaxYX()

	selected := 2
	selectedKey := Coords{0, 0}
	selectedButton := 1
	cursor := len(defaultText)
	text := defaultText

	keys := [4][10]gc.Char{
		{'1', '2', '3', '4', '5', '6', '7', '8', '9', '0'},
		{'Q', 'W', 'E', 'R', 'T', 'Y', 'U', 'I', 'O', 'P'},
		{'A', 'S', 'D', 'F', 'G', 'H', 'J', 'K', 'L', '-'},
		{'Z', 'X', 'C', 'V', 'B', 'N', 'M', '_', '<', '>'},
	}

	var ch gc.Key

	addText := func(input string) {
		if len(text)+len(input) < width-4 {
			text = fmt.Sprintf("%s%s%s", text[:cursor], s.ToLower(input), text[cursor:])
			cursor += len(input)
		}
	}

	for ch != gc.KEY_ESC {
		if selected == 0 {
			gc.Cursor(2)
		} else {
			gc.Cursor(1)
		}
		DrawBox(win, 1, 1, 3, width-2)
		win.MovePrint(2, 2, s.Repeat(" ", width-4))
		win.MovePrint(2, 2, text)

		for y, row := range keys {
			for x, key := range row {
				win.Move(5+(y*2), 2+(x*6))
				if selected == 1 && selectedKey.Y == y && selectedKey.X == x {
					win.ColorOn(1)
				}
				switch key {
				case ' ':
					continue
				case '<':
					win.AddChar('[')
					win.AddChar(' ')
					win.AddChar(gc.ACS_LARROW)
					win.AddChar(' ')
					win.AddChar(']')
				case '>':
					win.AddChar('[')
					win.AddChar(' ')
					win.AddChar(gc.ACS_RARROW)
					win.AddChar(' ')
					win.AddChar(']')
				case '_':
					win.Print("[SPC]")
				case '-':
					win.Print("[DEL]")
				default:
					win.AddChar('[')
					win.AddChar(' ')
					win.AddChar(key)
					win.AddChar(' ')
					win.AddChar(']')
				}
				win.ColorOff(1)
			}
		}

		var button int
		if selected == 2 {
			button = selectedButton
		} else {
			button = -1
		}
		DrawActionButtons(win, buttons, button)

		win.Move(2, cursor+2)

		win.NoutRefresh()
		gc.Update()

		ch = win.GetChar()

		switch ch {
		case gc.KEY_DOWN:
			if selected == 0 {
				selected = 1
			} else if selected == 1 {
				if selectedKey.Y < 3 {
					selectedKey.Y++
				} else {
					selected = 2
					if selectedKey.X < 3 {
						selectedButton = 0
					} else if selectedKey.X > 6 {
						selectedButton = 2
					} else {
						selectedButton = 1
					}
				}
			} else if selected == 2 {
				selected = 0
				selectedKey.Y = 0
			}
		case gc.KEY_UP:
			if selected == 0 {
				selected = 2
				selectedKey.Y = 3
				selectedButton = 1
			} else if selected == 1 {
				if selectedKey.Y > 0 {
					selectedKey.Y--
				} else {
					selected = 0
				}
			} else if selected == 2 {
				selected = 1
				selectedKey.Y = 3
				// FIXME: this only works well for 3 buttons
				if selectedButton == 0 {
					selectedKey.X = 2
				} else if selectedButton == 1 {
					selectedKey.X = 4
				} else {
					selectedKey.X = 7
				}
			}
		case gc.KEY_LEFT:
			if selected == 0 {
				if cursor > 0 {
					cursor--
				}
			} else if selected == 1 {
				if selectedKey.X > 0 {
					selectedKey.X--
				} else {
					selectedKey.X = 9
				}
			} else if selected == 2 {
				if selectedButton > 0 {
					selectedButton--
				} else {
					selectedButton = 2
				}
			}
		case gc.KEY_RIGHT:
			if selected == 0 {
				if cursor < len(text) {
					cursor++
				}
			} else if selected == 1 {
				if selectedKey.X < 9 {
					selectedKey.X++
				} else {
					selectedKey.X = 0
				}
			} else if selected == 2 {
				if selectedButton < 2 {
					selectedButton++
				} else {
					selectedButton = 0
				}
			}
		case gc.KEY_ENTER, 10, 13:
			if selected == 1 {
				c := string(rune(keys[selectedKey.Y][selectedKey.X]))
				if c == "-" {
					if cursor > 0 {
						text = text[:cursor-1] + text[cursor:]
						cursor--
					}
					break
				} else if c == "_" {
					c = " "
				} else if c == ">" {
					if cursor < len(text) {
						cursor++
					}
					break
				} else if c == "<" {
					if cursor > 0 {
						cursor--
					}
					break
				}
				addText(c)
			} else if selected == 2 {
				return selectedButton, text, nil
			}
		case gc.KEY_BACKSPACE, gc.KEY_DC, 127:
			if cursor > 0 {
				text = text[:cursor-1] + text[cursor:]
				cursor--
			}
		default:
			if ch >= 32 && ch <= 126 {
				addText(string(rune(ch)))
			}
		}

		gc.Update()
	}

	return -1, "", nil
}

func ListPicker(stdscr *gc.Window, title string, items []string, buttons []string, defaultButton int) (int, int, error) {
	selectedItem := 0
	selectedButton := defaultButton

	height := 18
	width := 70

	viewStart := 0
	viewHeight := height - 4

	win, err := NewWindow(stdscr, height, width, title, -1)
	if err != nil {
		return -1, -1, err
	}
	defer win.Delete()

	var ch gc.Key

	for ch != gc.KEY_ESC {
		// list items
		max := utils.MinInt([]int{len(items), viewHeight})

		for i := 0; i < max; i++ {
			var display string

			item := items[viewStart+i]

			if len(item) > width-5 {
				display = item[:width-8] + "..."
			} else {
				display = item
			}

			if viewStart+i == selectedItem {
				win.ColorOn(1)
			}

			win.MovePrint(i+1, 2, s.Repeat(" ", width-5))
			win.MovePrint(i+1, 2, display)
			win.ColorOff(1)
		}

		// scroll bar
		var gripHeight int
		var gripOffset int
		scrollHeight := viewHeight - 2

		// FIXME: not quite working
		if len(items) <= scrollHeight {
			gripHeight = scrollHeight
		} else {
			gripHeight = int(m.Ceil((float64(scrollHeight) / float64(len(items))) * float64(scrollHeight)))
		}

		if gripHeight >= scrollHeight {
			gripOffset = 0
		} else {
			gripOffset = int(m.Floor(float64(viewStart) * float64(scrollHeight) / float64(len(items))))
		}

		for i := 0; i < scrollHeight; i++ {
			if i >= gripOffset && i < gripOffset+gripHeight {
				win.ColorOn(1)
				win.MoveAddChar(i+2, width-2, ' ')
			} else {
				win.MoveAddChar(i+2, width-2, ' ')
			}
			win.ColorOff(1)
		}

		win.MoveAddChar(1, width-2, ' ')
		if viewStart > 0 {
			win.ColorOn(1)
			win.MoveAddChar(1, width-2, gc.ACS_UARROW)
			win.ColorOff(1)
		}

		win.MoveAddChar(height-4, width-2, ' ')
		if viewStart+viewHeight < len(items) {
			win.ColorOn(1)
			win.MoveAddChar(height-4, width-2, gc.ACS_DARROW)
			win.ColorOff(1)
		}

		// buttons
		DrawActionButtons(win, buttons, selectedButton)

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
