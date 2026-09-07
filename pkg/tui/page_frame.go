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

type PageFrame struct {
	*tview.Box
	app       *tview.Application
	content   tview.Primitive
	focus     tview.Primitive
	helpText  *tview.TextView
	buttonBar *ButtonBar
	onEscape  func()
	version   string
}

func NewPageFrame(app *tview.Application) *PageFrame {
	theme := CurrentTheme()
	frame := &PageFrame{
		Box: tview.NewBox(),
		app: app,
		helpText: tview.NewTextView().
			SetDynamicColors(true).
			SetTextAlign(tview.AlignCenter).
			SetTextColor(theme.SecondaryTextColor),
	}
	frame.SetBorder(true).
		SetBorderColor(theme.BorderColor).
		SetTitleColor(theme.BorderColor).
		SetTitleAlign(tview.AlignCenter)
	return frame
}

func (pf *PageFrame) SetTitle(path ...string) *PageFrame {
	pf.Box.SetTitle(" " + strings.Join(path, " > ") + " ")
	return pf
}

// SetVersion prints a build identifier in the bottom border, opposite the key
// hints, so a user can say which build they are running without guessing from
// a file date. Ignored when it does not fit.
func (pf *PageFrame) SetVersion(version string) *PageFrame {
	pf.version = version
	return pf
}

func (pf *PageFrame) SetContent(content tview.Primitive) *PageFrame {
	pf.content = content
	pf.focus = content
	return pf
}

func (pf *PageFrame) SetFocusTarget(focus tview.Primitive) *PageFrame {
	pf.focus = focus
	return pf
}

// SetHelpText sets the footer line. The text is escaped for the same reason
// dialog text is: the view has dynamic colours enabled, and callers put paths
// and file names here. The Favorites browser shows the current folder and
// GamesMenu shows a row's source paths, so a bracketed name lost its marker.
func (pf *PageFrame) SetHelpText(text string) *PageFrame {
	pf.helpText.SetText(tview.Escape(text))
	return pf
}

func (pf *PageFrame) SetButtonBar(bar *ButtonBar) *PageFrame {
	pf.buttonBar = bar
	return pf
}

func (pf *PageFrame) SetOnEscape(callback func()) *PageFrame {
	pf.onEscape = callback
	return pf
}

func (pf *PageFrame) Draw(screen tcell.Screen) {
	pf.DrawForSubclass(screen, pf)
	x, y, width, height := pf.GetInnerRect()
	if width <= 0 || height <= 0 {
		return
	}

	buttonHeight := 0
	if pf.buttonBar != nil {
		buttonHeight = 1
	}
	contentHeight := max(height-1-buttonHeight, 1)
	if pf.content != nil {
		pf.content.SetRect(x, y, width, contentHeight)
		pf.content.Draw(screen)
	}
	pf.helpText.SetRect(x, y+contentHeight, width, 1)
	pf.helpText.Draw(screen)
	if pf.buttonBar != nil {
		pf.buttonBar.SetRect(x, y+contentHeight+1, width, buttonHeight)
		pf.buttonBar.Draw(screen)
	}
	pf.drawHints(screen)
}

func hintsRunes() []rune {
	return []rune("↑↓: Rows │ ←→: Actions │ Enter: Confirm │ ESC: Back")
}

func (pf *PageFrame) drawHints(screen tcell.Screen) {
	x, y, width, height := pf.GetRect()
	if width <= 4 || height <= 2 {
		return
	}
	hints := hintsRunes()
	if len(hints) > width-4 {
		hints = hints[:width-4]
	}
	start := x + (width-len(hints))/2
	style := tcell.StyleDefault.
		Foreground(CurrentTheme().BorderColor).
		Background(CurrentTheme().PrimitiveBackgroundColor)
	for column := start - 1; column < start+len(hints)+1; column++ {
		screen.SetContent(column, y+height-1, ' ', nil, style)
	}
	for index, value := range hints {
		screen.SetContent(start+index, y+height-1, value, nil, style)
	}
	pf.drawVersion(screen, x, y, width, height, start)
}

// drawVersion puts the build in the right of the bottom border, but only when
// there is clear space between it and the key hints.
func (pf *PageFrame) drawVersion(screen tcell.Screen, x, y, width, height, hintsStart int) {
	if pf.version == "" {
		return
	}
	label := []rune(pf.version)
	start := x + width - 2 - len(label)
	if start < hintsStart+len(hintsRunes())+2 {
		return
	}
	style := tcell.StyleDefault.
		Foreground(CurrentTheme().BorderColor).
		Background(CurrentTheme().PrimitiveBackgroundColor)
	for column := start - 1; column < start+len(label)+1; column++ {
		screen.SetContent(column, y+height-1, ' ', nil, style)
	}
	for index, value := range label {
		screen.SetContent(start+index, y+height-1, value, nil, style)
	}
}

func (pf *PageFrame) Focus(delegate func(tview.Primitive)) {
	if pf.focus != nil {
		delegate(pf.focus)
	} else if pf.buttonBar != nil {
		delegate(pf.buttonBar)
	}
}

func (pf *PageFrame) HasFocus() bool {
	return pf.focus != nil && pf.focus.HasFocus() || pf.buttonBar != nil && pf.buttonBar.HasFocus()
}

func (pf *PageFrame) InputHandler() func(*tcell.EventKey, func(tview.Primitive)) {
	return pf.WrapInputHandler(func(event *tcell.EventKey, setFocus func(tview.Primitive)) {
		if event.Key() == tcell.KeyEscape && pf.onEscape != nil {
			pf.onEscape()
			return
		}
		if pf.focus != nil && pf.focus.HasFocus() {
			if handler := pf.focus.InputHandler(); handler != nil {
				handler(event, setFocus)
			}
			return
		}
		if pf.buttonBar != nil && pf.buttonBar.HasFocus() {
			if handler := pf.buttonBar.InputHandler(); handler != nil {
				handler(event, setFocus)
			}
		}
	})
}

func (pf *PageFrame) MouseHandler() func(
	tview.MouseAction,
	*tcell.EventMouse,
	func(tview.Primitive),
) (bool, tview.Primitive) {
	return pf.WrapMouseHandler(func(
		action tview.MouseAction,
		event *tcell.EventMouse,
		setFocus func(tview.Primitive),
	) (bool, tview.Primitive) {
		if pf.buttonBar != nil && pf.buttonBar.InRect(event.Position()) {
			return pf.buttonBar.MouseHandler()(action, event, setFocus)
		}
		if pf.content != nil {
			if handler := pf.content.MouseHandler(); handler != nil {
				return handler(action, event, setFocus)
			}
		}
		return false, nil
	})
}

func (pf *PageFrame) FocusContent() {
	if pf.focus != nil {
		pf.app.SetFocus(pf.focus)
	}
}

func (pf *PageFrame) FocusButtonBar() {
	if pf.buttonBar != nil {
		pf.app.SetFocus(pf.buttonBar)
	}
}

func (pf *PageFrame) SetupContentToButtonNavigation() {
	var list *tview.List
	switch focus := pf.focus.(type) {
	case *tview.List:
		list = focus
	case interface{ NavigationList() *tview.List }:
		list = focus.NavigationList()
	}
	if list == nil || pf.buttonBar == nil {
		return
	}
	pf.buttonBar.SetPersistentFocus(true)
	capture := list.GetInputCapture()
	list.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		// Left, Right and Tab move between actions; Enter presses the
		// highlighted one, which is what the "Enter: Confirm" hint promises.
		//
		// Routing Enter here means the list's own SetSelectedFunc is never
		// reached while this navigation is installed. Pages still behave
		// correctly because every one of them puts the row action first, so
		// Enter on a row activates that row. Keep that invariant: a page whose
		// first button is not its row action would make Enter do something
		// else entirely.
		case tcell.KeyLeft, tcell.KeyRight, tcell.KeyTab, tcell.KeyBacktab, tcell.KeyEnter:
			if handler := pf.buttonBar.InputHandler(); handler != nil {
				handler(event, func(tview.Primitive) {})
			}
			return nil
		case tcell.KeyEscape:
			if pf.onEscape != nil {
				pf.onEscape()
				return nil
			}
		default:
		}
		if capture != nil {
			return capture(event)
		}
		return event
	})
}
