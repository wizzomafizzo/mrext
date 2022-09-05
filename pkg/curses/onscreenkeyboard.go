package curses

import (
	"fmt"
	s "strings"

	gc "github.com/rthornton128/goncurses"
)

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
		DrawActionButtons(win, buttons, button, 6)

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
