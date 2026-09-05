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

// Portions adapted from Zaparoo Core, Copyright (c) 2026 The Zaparoo Project Contributors.

package tui

import (
	"strings"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

const (
	keyboardDelete  = "DEL"
	keyboardSubmit  = "OK"
	keyboardShift   = "SHFT"
	keyboardSymbols = "SYM"
	keyboardSpace   = "SPC"
	keyboardCancel  = "CANC"
)

var (
	keyboardLower = [][]string{
		{"1", "2", "3", "4", "5", "6", "7", "8", "9", "0"},
		{"q", "w", "e", "r", "t", "y", "u", "i", "o", "p"},
		{"a", "s", "d", "f", "g", "h", "j", "k", "l"},
		{"z", "x", "c", "v", "b", "n", "m", ",", "."},
		{keyboardShift, keyboardSymbols, keyboardSpace, keyboardDelete, keyboardSubmit, keyboardCancel},
	}
	keyboardUpper = [][]string{
		{"1", "2", "3", "4", "5", "6", "7", "8", "9", "0"},
		{"Q", "W", "E", "R", "T", "Y", "U", "I", "O", "P"},
		{"A", "S", "D", "F", "G", "H", "J", "K", "L"},
		{"Z", "X", "C", "V", "B", "N", "M", ",", "."},
		{keyboardShift, keyboardSymbols, keyboardSpace, keyboardDelete, keyboardSubmit, keyboardCancel},
	}
	keyboardSymbolKeys = [][]string{
		{"!", "@", "#", "$", "%", "^", "&", "*", "(", ")"},
		{"-", "_", "=", "+", "[", "]", "{", "}", "\\", "|"},
		{";", ":", "'", "\"", "`", "~", "/", "?", "<", ">"},
		{},
		{"ABC", keyboardSpace, keyboardDelete, keyboardSubmit, keyboardCancel},
	}
)

type VirtualKeyboard struct {
	*tview.Box
	onSubmit  func(string)
	onCancel  func()
	text      []rune
	cursorRow int
	cursorCol int
	shift     bool
	symbols   bool
}

func NewVirtualKeyboard(initialText string, onSubmit func(string), onCancel func()) *VirtualKeyboard {
	keyboard := &VirtualKeyboard{
		Box:       tview.NewBox(),
		onSubmit:  onSubmit,
		onCancel:  onCancel,
		text:      []rune(initialText),
		cursorRow: len(keyboardLower) - 1,
		cursorCol: 4,
	}
	keyboard.SetBorder(true).SetTitle(" Keyboard ")
	return keyboard
}

func (k *VirtualKeyboard) Text() string {
	return string(k.text)
}

func (k *VirtualKeyboard) layout() [][]string {
	switch {
	case k.symbols:
		return keyboardSymbolKeys
	case k.shift:
		return keyboardUpper
	default:
		return keyboardLower
	}
}

func (k *VirtualKeyboard) Draw(screen tcell.Screen) {
	k.DrawForSubclass(screen, k)
	x, y, width, height := k.GetInnerRect()
	if width <= 0 || height <= 0 {
		return
	}
	theme := CurrentTheme()
	inputStyle := tcell.StyleDefault.
		Foreground(theme.PrimaryTextColor).
		Background(theme.FieldFocusedBackground)
	for column := x; column < x+width; column++ {
		screen.SetContent(column, y, ' ', nil, inputStyle)
	}
	display := k.text
	if len(display) > width-2 {
		display = display[len(display)-(width-2):]
	}
	for index, value := range display {
		screen.SetContent(x+1+index, y, value, nil, inputStyle)
	}
	cursor := x + 1 + len(display)
	if cursor < x+width-1 {
		screen.SetContent(cursor, y, '_', nil, inputStyle.Blink(true))
	}

	for row, keys := range k.layout() {
		if y+1+row >= y+height || len(keys) == 0 {
			continue
		}
		k.drawRow(screen, x, y+1+row, row, keys, theme)
	}
}

func (k *VirtualKeyboard) drawRow(
	screen tcell.Screen,
	x, y, row int,
	keys []string,
	theme *Theme,
) {
	if row == len(k.layout())-1 {
		leftCount := len(keys) - 4
		rightStart := 39 - 3*5
		for column, key := range keys {
			keyX, keyWidth := x+column*5, 4
			switch {
			case column == leftCount:
				keyWidth = rightStart - leftCount*5 - 1
			case column > leftCount:
				keyX = x + rightStart + (column-leftCount-1)*5
			}
			k.drawKey(screen, keyX, y, key, keyWidth, row, column, theme)
		}
		return
	}
	for column, key := range keys {
		k.drawKey(screen, x+column*4, y, key, 3, row, column, theme)
	}
}

func (k *VirtualKeyboard) drawKey(
	screen tcell.Screen,
	x, y int,
	label string,
	width, row, column int,
	theme *Theme,
) {
	style := tcell.StyleDefault.
		Foreground(theme.PrimaryTextColor).
		Background(theme.PrimitiveBackgroundColor)
	if row == k.cursorRow && column == k.cursorCol {
		style = tcell.StyleDefault.
			Foreground(tcell.GetColor(theme.HighlightForegroundName)).
			Background(tcell.GetColor(theme.HighlightBackgroundName))
	} else if isKeyboardAction(label) {
		style = style.Foreground(theme.LabelColor)
	}
	padding := max((width-len(label))/2, 0)
	for index := range width {
		value := ' '
		labelIndex := index - padding
		if labelIndex >= 0 && labelIndex < len(label) {
			value = rune(label[labelIndex])
		}
		screen.SetContent(x+index, y, value, nil, style)
	}
}

func isKeyboardAction(key string) bool {
	switch key {
	case keyboardDelete, keyboardSubmit, keyboardShift, keyboardSymbols,
		keyboardSpace, keyboardCancel, "ABC":
		return true
	default:
		return false
	}
}

func (k *VirtualKeyboard) InputHandler() func(*tcell.EventKey, func(tview.Primitive)) {
	return k.WrapInputHandler(func(event *tcell.EventKey, _ func(tview.Primitive)) {
		layout := k.layout()
		switch event.Key() {
		case tcell.KeyUp:
			k.moveVertical(layout, -1)
		case tcell.KeyDown:
			k.moveVertical(layout, 1)
		case tcell.KeyLeft:
			k.cursorCol = (k.cursorCol - 1 + len(layout[k.cursorRow])) % len(layout[k.cursorRow])
		case tcell.KeyRight:
			k.cursorCol = (k.cursorCol + 1) % len(layout[k.cursorRow])
		case tcell.KeyEnter:
			k.activate()
		case tcell.KeyEscape:
			if k.onCancel != nil {
				k.onCancel()
			}
		case tcell.KeyBackspace, tcell.KeyBackspace2, tcell.KeyDelete:
			k.deleteLast()
		case tcell.KeyRune:
			k.text = append(k.text, event.Rune())
		default:
		}
	})
}

func (k *VirtualKeyboard) moveVertical(layout [][]string, direction int) {
	column := k.cursorCol
	for {
		k.cursorRow = (k.cursorRow + direction + len(layout)) % len(layout)
		if len(layout[k.cursorRow]) > 0 {
			k.cursorCol = min(column, len(layout[k.cursorRow])-1)
			return
		}
	}
}

func (k *VirtualKeyboard) deleteLast() {
	if len(k.text) > 0 {
		k.text = k.text[:len(k.text)-1]
	}
}

func (k *VirtualKeyboard) activate() {
	layout := k.layout()
	if k.cursorRow >= len(layout) || k.cursorCol >= len(layout[k.cursorRow]) {
		return
	}
	key := layout[k.cursorRow][k.cursorCol]
	switch key {
	case keyboardDelete:
		k.deleteLast()
	case keyboardSubmit:
		if k.onSubmit != nil {
			k.onSubmit(string(k.text))
		}
	case keyboardShift:
		k.shift = !k.shift
		k.symbols = false
	case keyboardSymbols:
		k.symbols = true
		k.shift = false
		k.cursorCol = 0
	case "ABC":
		k.symbols = false
		k.cursorCol = 1
	case keyboardSpace:
		k.text = append(k.text, ' ')
	case keyboardCancel:
		if k.onCancel != nil {
			k.onCancel()
		}
	default:
		if strings.TrimSpace(key) != "" {
			k.text = append(k.text, []rune(key)...)
			if k.shift && !k.symbols {
				k.shift = false
			}
		}
	}
}

func (k *VirtualKeyboard) Focus(delegate func(tview.Primitive)) {
	k.Box.Focus(delegate)
}

func (k *VirtualKeyboard) HasFocus() bool {
	return k.Box.HasFocus()
}
