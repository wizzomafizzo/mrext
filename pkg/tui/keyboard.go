// mrext
// Copyright (c) 2026 mrext contributors.
// SPDX-License-Identifier: GPL-3.0-or-later
//
// This file is part of mrext.
//
// mrext is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// mrext is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with mrext. If not, see <http://www.gnu.org/licenses/>.

package tui

import (
	"strings"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

type onScreenKeyboard struct {
	*tview.Box
	done           func(button int, text string)
	text           []rune
	buttons        []string
	selectedKey    [2]int
	selectedButton int
	section        int
	cursor         int
}

var keyboardKeys = [][]rune{
	{'1', '2', '3', '4', '5', '6', '7', '8', '9', '0'},
	{'Q', 'W', 'E', 'R', 'T', 'Y', 'U', 'I', 'O', 'P'},
	{'A', 'S', 'D', 'F', 'G', 'H', 'J', 'K', 'L', '-'},
	{'Z', 'X', 'C', 'V', 'B', 'N', 'M', '_', '<', '>'},
}

func newOnScreenKeyboard(title string, buttons []string, defaultText string, done func(int, string)) *onScreenKeyboard {
	keyboard := &onScreenKeyboard{
		Box:            tview.NewBox(),
		done:           done,
		text:           []rune(defaultText),
		buttons:        append([]string(nil), buttons...),
		selectedButton: min(1, len(buttons)-1),
		section:        2,
		cursor:         len([]rune(defaultText)),
	}
	keyboard.SetBorder(true).SetTitle(" " + title + " ")
	return keyboard
}

func drawText(screen tcell.Screen, x, y, width int, text string, style tcell.Style) {
	runes := []rune(text)
	for i := range width {
		value := ' '
		if i < len(runes) {
			value = runes[i]
		}
		screen.SetContent(x+i, y, value, nil, style)
	}
}

func (k *onScreenKeyboard) Draw(screen tcell.Screen) {
	k.DrawForSubclass(screen, k)
	x, y, width, _ := k.GetInnerRect()
	if width < 4 {
		return
	}
	fieldStyle := tcell.StyleDefault.Foreground(tcell.ColorWhite).Background(tcell.ColorBlack)
	selectedStyle := tcell.StyleDefault.Foreground(tcell.ColorBlack).Background(tcell.ColorWhite)
	drawText(screen, x+1, y+1, width-2, string(k.text), fieldStyle)
	if k.section == 0 && k.cursor < width-2 {
		screen.ShowCursor(x+1+k.cursor, y+1)
	} else {
		screen.HideCursor()
	}

	for row, keys := range keyboardKeys {
		for column, key := range keys {
			label := "[ " + string(key) + " ]"
			switch key {
			case '_':
				label = "[SPC]"
			case '-':
				label = "[DEL]"
			case '<':
				label = "[ ← ]"
			case '>':
				label = "[ → ]"
			}
			style := fieldStyle
			if k.section == 1 && k.selectedKey[0] == row && k.selectedKey[1] == column {
				style = selectedStyle
			}
			keyX := x + 1 + column*6
			keyY := y + 4 + row*2
			drawText(screen, keyX, keyY, 5, label, style)
		}
	}

	buttonY := y + 12
	buttonWidth := 0
	for _, button := range k.buttons {
		buttonWidth += len([]rune(button)) + 4
	}
	buttonWidth += max(len(k.buttons)-1, 0) * 4
	buttonX := x + max((width-buttonWidth)/2, 0)
	for index, button := range k.buttons {
		label := "< " + button + " >"
		style := fieldStyle
		if k.section == 2 && k.selectedButton == index {
			style = selectedStyle
		}
		drawText(screen, buttonX, buttonY, len([]rune(label)), label, style)
		buttonX += len([]rune(label)) + 4
	}
}

func (k *onScreenKeyboard) Focus(delegate func(tview.Primitive)) {
	delegate(k)
}

func (*onScreenKeyboard) HasFocus() bool {
	return true
}

func (k *onScreenKeyboard) addText(input string) {
	value := []rune(strings.ToLower(input))
	_, _, width, _ := k.GetInnerRect()
	if len(k.text)+len(value) >= width-2 {
		return
	}
	updated := make([]rune, 0, len(k.text)+len(value))
	updated = append(updated, k.text[:k.cursor]...)
	updated = append(updated, value...)
	updated = append(updated, k.text[k.cursor:]...)
	k.text = updated
	k.cursor += len(value)
}

func (k *onScreenKeyboard) activateKey() {
	key := keyboardKeys[k.selectedKey[0]][k.selectedKey[1]]
	switch key {
	case '-':
		if k.cursor > 0 {
			k.text = append(k.text[:k.cursor-1], k.text[k.cursor:]...)
			k.cursor--
		}
	case '_':
		k.addText(" ")
	case '<':
		if k.cursor > 0 {
			k.cursor--
		}
	case '>':
		if k.cursor < len(k.text) {
			k.cursor++
		}
	default:
		k.addText(string(key))
	}
}

func (k *onScreenKeyboard) InputHandler() func(*tcell.EventKey, func(tview.Primitive)) {
	return k.WrapInputHandler(func(event *tcell.EventKey, _ func(tview.Primitive)) {
		switch event.Key() {
		case tcell.KeyEscape:
			k.done(-1, string(k.text))
		case tcell.KeyDown:
			switch k.section {
			case 0:
				k.section = 1
			case 1:
				if k.selectedKey[0] < len(keyboardKeys)-1 {
					k.selectedKey[0]++
				} else {
					k.section = 2
					k.selectedButton = min(k.selectedKey[1]*len(k.buttons)/len(keyboardKeys[0]), len(k.buttons)-1)
				}
			case 2:
				k.section = 0
				k.selectedKey[0] = 0
			}
		case tcell.KeyUp:
			switch k.section {
			case 0:
				k.section = 2
			case 1:
				if k.selectedKey[0] > 0 {
					k.selectedKey[0]--
				} else {
					k.section = 0
				}
			case 2:
				k.section = 1
				k.selectedKey[0] = len(keyboardKeys) - 1
				if len(k.buttons) > 0 {
					keyColumn := k.selectedButton*len(keyboardKeys[0])/len(k.buttons) + 1
					k.selectedKey[1] = min(keyColumn, len(keyboardKeys[0])-1)
				}
			}
		case tcell.KeyLeft:
			switch k.section {
			case 0:
				if k.cursor > 0 {
					k.cursor--
				}
			case 1:
				k.selectedKey[1] = (k.selectedKey[1] - 1 + len(keyboardKeys[0])) % len(keyboardKeys[0])
			case 2:
				if len(k.buttons) > 0 {
					k.selectedButton = (k.selectedButton - 1 + len(k.buttons)) % len(k.buttons)
				}
			}
		case tcell.KeyRight:
			switch k.section {
			case 0:
				if k.cursor < len(k.text) {
					k.cursor++
				}
			case 1:
				k.selectedKey[1] = (k.selectedKey[1] + 1) % len(keyboardKeys[0])
			case 2:
				if len(k.buttons) > 0 {
					k.selectedButton = (k.selectedButton + 1) % len(k.buttons)
				}
			}
		case tcell.KeyEnter:
			if k.section == 1 {
				k.activateKey()
			} else if k.section == 2 && len(k.buttons) > 0 {
				k.done(k.selectedButton, string(k.text))
			}
		case tcell.KeyBackspace, tcell.KeyBackspace2, tcell.KeyDelete:
			if k.cursor > 0 {
				k.text = append(k.text[:k.cursor-1], k.text[k.cursor:]...)
				k.cursor--
			}
		case tcell.KeyRune:
			if event.Rune() >= 32 && event.Rune() <= 126 {
				k.addText(string(event.Rune()))
			}
		default:
			return
		}
	})
}

func OnScreenKeyboard(title string, buttons []string, defaultText string) (button int, text string, err error) {
	button = -1
	text = defaultText
	builder := func() (*tview.Application, error) {
		app := tview.NewApplication()
		keyboard := newOnScreenKeyboard(title, buttons, text, func(selected int, value string) {
			button = selected
			text = value
			app.Stop()
		})
		return app.SetRoot(Centered(63, 16, keyboard), true).SetFocus(keyboard), nil
	}
	if runErr := BuildAndRetry(builder); runErr != nil {
		return -1, text, runErr
	}
	return button, text, nil
}
