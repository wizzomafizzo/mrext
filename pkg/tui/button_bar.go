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
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

type ButtonBar struct {
	*tview.Box
	app             *tview.Application
	onEscape        func()
	onUp            func()
	onDown          func()
	onWrap          func()
	onLeft          func()
	onRight         func()
	helpCallback    func(string)
	buttons         []*tview.Button
	helpTexts       []string
	focusedIndex    int
	persistentFocus bool
}

func NewButtonBar(app *tview.Application) *ButtonBar {
	return &ButtonBar{Box: tview.NewBox(), app: app}
}

func (bb *ButtonBar) AddButton(label string, action func()) *ButtonBar {
	return bb.AddButtonWithHelp(label, "", action)
}

func (bb *ButtonBar) AddButtonWithHelp(label, helpText string, action func()) *ButtonBar {
	theme := CurrentTheme()
	button := tview.NewButton(label).
		SetStyle(tcell.StyleDefault.
			Foreground(theme.PrimaryTextColor).
			Background(theme.ContrastBackgroundColor)).
		SetActivatedStyle(tcell.StyleDefault.
			Foreground(tcell.ColorBlack).
			Background(theme.BorderColor)).
		SetSelectedFunc(action)
	bb.buttons = append(bb.buttons, button)
	bb.helpTexts = append(bb.helpTexts, helpText)
	return bb
}

func (bb *ButtonBar) SetHelpCallback(callback func(string)) *ButtonBar {
	bb.helpCallback = callback
	return bb
}

func (bb *ButtonBar) SetPersistentFocus(enabled bool) *ButtonBar {
	bb.persistentFocus = enabled
	return bb
}

func (bb *ButtonBar) SetupNavigation(onEscape func()) *ButtonBar {
	bb.onEscape = onEscape
	return bb
}

func (bb *ButtonBar) SetOnUp(callback func()) *ButtonBar {
	bb.onUp = callback
	return bb
}

func (bb *ButtonBar) SetOnDown(callback func()) *ButtonBar {
	bb.onDown = callback
	return bb
}

func (bb *ButtonBar) SetOnWrap(callback func()) *ButtonBar {
	bb.onWrap = callback
	return bb
}

func (bb *ButtonBar) SetOnLeft(callback func()) *ButtonBar {
	bb.onLeft = callback
	return bb
}

func (bb *ButtonBar) SetOnRight(callback func()) *ButtonBar {
	bb.onRight = callback
	return bb
}

// FocusedIndex returns the highlighted button.
func (bb *ButtonBar) FocusedIndex() int {
	return bb.focusedIndex
}

// SetFocusedIndex highlights a button without activating it, so a rebuilt
// page can keep the button the user had moved to.
func (bb *ButtonBar) SetFocusedIndex(index int) *ButtonBar {
	if index >= 0 && index < len(bb.buttons) {
		bb.focusedIndex = index
		bb.triggerHelp()
	}
	return bb
}

func (bb *ButtonBar) triggerHelp() {
	if bb.helpCallback != nil && bb.focusedIndex < len(bb.helpTexts) {
		bb.helpCallback(bb.helpTexts[bb.focusedIndex])
	}
}

func (bb *ButtonBar) Draw(screen tcell.Screen) {
	bb.DrawForSubclass(screen, bb)
	x, y, width, _ := bb.GetInnerRect()
	if len(bb.buttons) == 0 || width <= 0 {
		return
	}

	spacing := 2
	buttonWidth := max((width-spacing*(len(bb.buttons)-1))/len(bb.buttons), 6)
	focused := bb.HasFocus() || bb.persistentFocus
	currentX := x
	for index, button := range bb.buttons {
		widthRemaining := x + width - currentX
		if widthRemaining <= 0 {
			break
		}
		itemWidth := min(buttonWidth, widthRemaining)
		button.SetRect(currentX, y, itemWidth, 1)
		if focused && index == bb.focusedIndex {
			button.Focus(func(tview.Primitive) {})
		} else {
			button.Blur()
		}
		button.Draw(screen)
		currentX += itemWidth + spacing
	}
}

func (bb *ButtonBar) InputHandler() func(*tcell.EventKey, func(tview.Primitive)) {
	return bb.WrapInputHandler(func(event *tcell.EventKey, setFocus func(tview.Primitive)) {
		if len(bb.buttons) == 0 {
			return
		}
		switch event.Key() {
		case tcell.KeyLeft:
			if bb.focusedIndex == 0 && bb.onLeft != nil {
				bb.onLeft()
			} else {
				bb.focusedIndex = (bb.focusedIndex - 1 + len(bb.buttons)) % len(bb.buttons)
				bb.triggerHelp()
			}
		case tcell.KeyBacktab:
			if bb.focusedIndex == 0 && bb.onWrap != nil {
				bb.onWrap()
			} else {
				bb.focusedIndex = (bb.focusedIndex - 1 + len(bb.buttons)) % len(bb.buttons)
				bb.triggerHelp()
			}
		case tcell.KeyRight:
			if bb.focusedIndex == len(bb.buttons)-1 && bb.onRight != nil {
				bb.onRight()
			} else {
				bb.focusedIndex = (bb.focusedIndex + 1) % len(bb.buttons)
				bb.triggerHelp()
			}
		case tcell.KeyTab:
			if bb.focusedIndex == len(bb.buttons)-1 && bb.onWrap != nil {
				bb.onWrap()
			} else {
				bb.focusedIndex = (bb.focusedIndex + 1) % len(bb.buttons)
				bb.triggerHelp()
			}
		case tcell.KeyUp:
			if bb.onUp != nil {
				bb.onUp()
			}
		case tcell.KeyDown:
			if bb.onDown != nil {
				bb.onDown()
			} else if bb.onUp != nil {
				bb.onUp()
			}
		case tcell.KeyEnter:
			if handler := bb.buttons[bb.focusedIndex].InputHandler(); handler != nil {
				handler(event, setFocus)
			}
		case tcell.KeyEscape:
			if bb.onEscape != nil {
				bb.onEscape()
			}
		default:
		}
	})
}

func (bb *ButtonBar) MouseHandler() func(
	tview.MouseAction,
	*tcell.EventMouse,
	func(tview.Primitive),
) (bool, tview.Primitive) {
	return bb.WrapMouseHandler(func(
		action tview.MouseAction,
		event *tcell.EventMouse,
		setFocus func(tview.Primitive),
	) (bool, tview.Primitive) {
		for index, button := range bb.buttons {
			if !button.InRect(event.Position()) {
				continue
			}
			bb.focusedIndex = index
			bb.triggerHelp()
			if !bb.persistentFocus {
				setFocus(bb)
			}
			if action == tview.MouseLeftClick {
				if handler := button.MouseHandler(); handler != nil {
					buttonFocus := setFocus
					if bb.persistentFocus {
						buttonFocus = func(tview.Primitive) {}
					}
					return handler(action, event, buttonFocus)
				}
			}
			return action == tview.MouseLeftDown, nil
		}
		return false, nil
	})
}

func (bb *ButtonBar) Focus(delegate func(tview.Primitive)) {
	if len(bb.buttons) > 0 {
		bb.Box.Focus(delegate)
		bb.triggerHelp()
	}
}

func (bb *ButtonBar) HasFocus() bool {
	return bb.Box.HasFocus()
}

func (bb *ButtonBar) GetFirstButton() *tview.Button {
	if len(bb.buttons) == 0 {
		return nil
	}
	return bb.buttons[0]
}

func (bb *ButtonBar) UpdateButtonLabel(index int, label string) {
	if index >= 0 && index < len(bb.buttons) {
		bb.buttons[index].SetLabel(label)
	}
}
