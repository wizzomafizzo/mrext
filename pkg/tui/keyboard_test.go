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

func TestOnScreenKeyboardAcceptsTypingAndButton(t *testing.T) {
	selected := -1
	text := ""
	keyboard := newOnScreenKeyboard(
		"Search",
		[]string{"Options", "Search", "Exit"},
		"nes",
		func(button int, value string) {
			selected = button
			text = value
		},
	)
	keyboard.SetRect(0, 0, 63, 16)
	handle := keyboard.InputHandler()

	keyboard.section = 0
	handle(tcell.NewEventKey(tcell.KeyRune, 'X', tcell.ModNone), nil)
	keyboard.section = 2
	keyboard.selectedButton = 1
	handle(tcell.NewEventKey(tcell.KeyEnter, 0, tcell.ModNone), nil)

	if selected != 1 || text != "nesx" {
		t.Fatalf("selected=%d text=%q", selected, text)
	}
}

func TestOnScreenKeyboardEditsAtCursor(t *testing.T) {
	keyboard := newOnScreenKeyboard("Search", []string{"Search"}, "ab", func(int, string) {})
	keyboard.SetRect(0, 0, 63, 16)
	keyboard.section = 0
	keyboard.cursor = 1
	handle := keyboard.InputHandler()
	handle(tcell.NewEventKey(tcell.KeyRune, 'Z', tcell.ModNone), nil)
	handle(tcell.NewEventKey(tcell.KeyBackspace2, 0, tcell.ModNone), nil)

	if string(keyboard.text) != "ab" || keyboard.cursor != 1 {
		t.Fatalf("text=%q cursor=%d", keyboard.text, keyboard.cursor)
	}
}
