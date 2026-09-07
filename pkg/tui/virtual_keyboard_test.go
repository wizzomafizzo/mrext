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
	"testing"

	"github.com/gdamore/tcell/v2"
)

func pressKeyboard(keyboard *VirtualKeyboard, key tcell.Key) {
	keyboard.InputHandler()(tcell.NewEventKey(key, 0, tcell.ModNone), nil)
}

func selectedKey(keyboard *VirtualKeyboard) string {
	return keyboard.layout()[keyboard.cursorRow][keyboard.cursorCol]
}

// selectedSpan renders the keyboard and reports where the highlighted key is
// actually drawn, so the tests measure the layout instead of recomputing it.
func selectedSpan(t *testing.T, keyboard *VirtualKeyboard) (row, start, end int) {
	t.Helper()
	screen := tcell.NewSimulationScreen("UTF-8")
	if err := screen.Init(); err != nil {
		t.Fatal(err)
	}
	defer screen.Fini()
	screen.SetSize(keyboardModalWidth, keyboardModalHeight)
	keyboard.SetRect(0, 0, keyboardModalWidth, keyboardModalHeight)
	keyboard.Draw(screen)
	highlight := tcell.GetColor(CurrentTheme().HighlightBackgroundName)
	row, start, end = -1, -1, -1
	for y := range keyboardModalHeight {
		for x := range keyboardModalWidth {
			_, style, _ := screen.Get(x, y)
			_, background, _ := style.Decompose()
			if background != highlight {
				continue
			}
			if row < 0 {
				row, start = y, x
			}
			end = x
		}
	}
	if row < 0 {
		t.Fatalf("no key is highlighted for %q", selectedKey(keyboard))
	}
	return row, start, end
}

// The character rows are drawn on a four-cell pitch and the action row on a
// five-cell one, so a plain column index does not survive the move between
// them: "v" lands on DEL, seven cells away. Pressing down must land on the key
// drawn underneath the one being left.
func TestVirtualKeyboardActionKeysSitUnderTheCharactersAbove(t *testing.T) {
	if !SetCurrentTheme("default") {
		t.Fatal("failed to select default theme")
	}
	characters := keyboardLower[len(keyboardLower)-2]
	for column, key := range characters {
		keyboard := NewVirtualKeyboard("", nil, nil)
		keyboard.cursorRow = len(keyboardLower) - 2
		keyboard.cursorCol = column
		characterRow, characterStart, characterEnd := selectedSpan(t, keyboard)

		pressKeyboard(keyboard, tcell.KeyDown)
		actionRow, actionStart, actionEnd := selectedSpan(t, keyboard)

		if actionRow != characterRow+1 {
			t.Fatalf("%q moved from row %d to row %d", key, characterRow, actionRow)
		}
		if actionEnd < characterStart || actionStart > characterEnd {
			t.Fatalf("down from %q at %d-%d landed on %q at %d-%d, which is not drawn under it",
				key, characterStart, characterEnd, selectedKey(keyboard), actionStart, actionEnd)
		}
	}
}

func TestVirtualKeyboardVerticalNavigationReturnsToTheSameKey(t *testing.T) {
	actions := keyboardLower[len(keyboardLower)-1]
	for column, key := range actions {
		keyboard := NewVirtualKeyboard("", nil, nil)
		keyboard.cursorRow = len(keyboardLower) - 1
		keyboard.cursorCol = column
		pressKeyboard(keyboard, tcell.KeyUp)
		above := selectedKey(keyboard)
		pressKeyboard(keyboard, tcell.KeyDown)
		if got := selectedKey(keyboard); got != key {
			t.Fatalf("%q went up to %q and came back to %q", key, above, got)
		}
	}
}

func TestVirtualKeyboardSymbolToggleKeepsTheCursorWhereItWas(t *testing.T) {
	if !SetCurrentTheme("default") {
		t.Fatal("failed to select default theme")
	}
	keyboard := NewVirtualKeyboard("", nil, nil)
	keyboard.cursorRow = len(keyboardLower) - 1
	keyboard.cursorCol = 1
	if selectedKey(keyboard) != keyboardSymbols {
		t.Fatalf("expected to start on SYM, got %q", selectedKey(keyboard))
	}
	symbolRow, symbolStart, symbolEnd := selectedSpan(t, keyboard)

	pressKeyboard(keyboard, tcell.KeyEnter)
	if selectedKey(keyboard) != "ABC" {
		t.Fatalf("SYM left the cursor on %q", selectedKey(keyboard))
	}
	row, start, end := selectedSpan(t, keyboard)
	if row != symbolRow || start != symbolStart || end != symbolEnd {
		t.Fatalf("ABC is drawn at %d:%d-%d but SYM was at %d:%d-%d",
			row, start, end, symbolRow, symbolStart, symbolEnd)
	}

	pressKeyboard(keyboard, tcell.KeyEnter)
	if selectedKey(keyboard) != keyboardSymbols {
		t.Fatalf("ABC left the cursor on %q", selectedKey(keyboard))
	}
}

// The symbols action row carries a blank in SHFT's column so ABC lines up with
// SYM. It is drawn, but it is not a key, and the cursor must never stop there.
func TestVirtualKeyboardNeverSelectsTheSpacer(t *testing.T) {
	keyboard := NewVirtualKeyboard("", nil, nil)
	keyboard.cursorRow = len(keyboardLower) - 1
	keyboard.cursorCol = 1
	pressKeyboard(keyboard, tcell.KeyEnter)
	if !keyboard.symbols {
		t.Fatal("SYM did not switch to the symbols layout")
	}

	characters := keyboardSymbolKeys[len(keyboardSymbolKeys)-3]
	for column := range characters {
		keyboard.cursorRow = len(keyboardSymbolKeys) - 3
		keyboard.cursorCol = column
		pressKeyboard(keyboard, tcell.KeyDown)
		if selectedKey(keyboard) == "" {
			t.Fatalf("down from column %d rested on the spacer", column)
		}
	}

	for _, direction := range []tcell.Key{tcell.KeyLeft, tcell.KeyRight} {
		keyboard.cursorRow = len(keyboardSymbolKeys) - 1
		keyboard.cursorCol = 1
		for step := range len(keyboardSymbolKeys[len(keyboardSymbolKeys)-1]) * 2 {
			pressKeyboard(keyboard, direction)
			if selectedKey(keyboard) == "" {
				t.Fatalf("step %d of horizontal navigation rested on the spacer", step)
			}
		}
	}
}
