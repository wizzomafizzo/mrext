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

// The grid is 39 cells wide. Character rows sit on a four-cell pitch, ten to a
// row. The action row is [SHFT SYM] [SPC] [DEL OK CANC] on a five-cell pitch,
// with the space bar stretched across whatever is left in the middle.
const (
	keyboardGridWidth   = 39
	keyboardKeyPitch    = 4
	keyboardActionPitch = 5
	keyboardLeftKeys    = 2
	keyboardSpaceCol    = 2
	keyboardRightCol    = 3
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
		// The blank holds ABC in the column SYM occupies in the other
		// layouts, so toggling between them does not move the cursor.
		{"", "ABC", keyboardSpace, keyboardDelete, keyboardSubmit, keyboardCancel},
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
	if row != len(k.layout())-1 {
		for column, key := range keys {
			k.drawKey(screen, x+column*keyboardKeyPitch, y, key, keyboardKeyPitch-1, row, column, theme)
		}
		return
	}
	leftEnd := keyboardLeftKeys * keyboardActionPitch
	rightStart := keyboardGridWidth - (len(keys)-keyboardRightCol)*keyboardActionPitch
	for column, key := range keys {
		keyX, keyWidth := x+column*keyboardActionPitch, keyboardActionPitch-1
		switch {
		case column == keyboardSpaceCol:
			keyX, keyWidth = x+leftEnd, rightStart-leftEnd-1
		case column > keyboardSpaceCol:
			keyX = x + rightStart + (column-keyboardRightCol)*keyboardActionPitch
		}
		k.drawKey(screen, keyX, y, key, keyWidth, row, column, theme)
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
			k.cursorCol = k.moveHorizontal(layout, -1)
		case tcell.KeyRight:
			k.cursorCol = k.moveHorizontal(layout, 1)
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

// moveHorizontal wraps along the current row, stepping over the blank that
// holds the symbols action row in line with the other layouts.
func (k *VirtualKeyboard) moveHorizontal(layout [][]string, direction int) int {
	keys := layout[k.cursorRow]
	column := k.cursorCol
	for range keys {
		column = (column + direction + len(keys)) % len(keys)
		if keys[column] != "" {
			return column
		}
	}
	return k.cursorCol
}

func (k *VirtualKeyboard) moveVertical(layout [][]string, direction int) {
	from := k.cursorRow
	row := k.cursorRow
	for range layout {
		row = (row + direction + len(layout)) % len(layout)
		if len(layout[row]) > 0 {
			break
		}
	}
	k.cursorRow = row
	k.cursorCol = mapKeyboardColumn(layout, from, row, k.cursorCol)
}

// mapKeyboardColumn keeps the cursor under the key it left when crossing
// between the character rows and the action row, which are drawn on different
// pitches: z<->SHFT, x/c<->SYM, v/b/n<->SPC, m<->DEL, ,<->OK, .<->CANC. This
// is Zaparoo Core's mapping, and without it a plain column index sends "." to
// CANC on the way down and lands on "n" on the way back up, with everything
// from "v" rightwards jumping across the row.
func mapKeyboardColumn(layout [][]string, from, to, column int) int {
	bottom := len(layout) - 1
	keys := layout[to]
	switch {
	case from == bottom && to != bottom:
		return skipBlankKey(keys, actionToCharacterColumn(column, len(keys)))
	case from != bottom && to == bottom:
		return skipBlankKey(keys, characterToActionColumn(column))
	default:
		return min(column, len(keys)-1)
	}
}

func actionToCharacterColumn(column, width int) int {
	switch column {
	case 0:
		return 0
	case 1:
		return min(1, width-1)
	case 2:
		return min(4, width-1)
	case 3:
		return min(6, width-1)
	case 4:
		return min(7, width-1)
	default:
		return width - 1
	}
}

func characterToActionColumn(column int) int {
	switch column {
	case 0:
		return 0
	case 1, 2:
		return 1
	case 3, 4, 5:
		return keyboardSpaceCol
	case 6:
		return keyboardRightCol
	case 7:
		return 4
	default:
		return 5
	}
}

// skipBlankKey shifts off the symbols row's spacer, which is drawn but is not
// a key. Zaparoo lets the cursor rest on it; there is nothing to press there.
func skipBlankKey(keys []string, column int) int {
	for column < len(keys) && keys[column] == "" {
		column++
	}
	return min(column, len(keys)-1)
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
		// ABC sits where SYM was, so the cursor does not appear to move.
		k.cursorCol = 1
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
