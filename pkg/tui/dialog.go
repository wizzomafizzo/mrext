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
	dialogChromeWidth = 4
	dialogMargin      = 2
)

type Dialog struct {
	*tview.Box
	frame    *tview.Box
	textView *tview.TextView
	form     *tview.Form
	done     func(int)
	text     string
	buttons  []string
	overflow bool
}

func NewDialog() *Dialog {
	theme := CurrentTheme()
	dialog := &Dialog{
		Box: tview.NewBox(),
		frame: tview.NewBox().
			SetBackgroundColor(tview.Styles.ContrastBackgroundColor),
		textView: tview.NewTextView().
			SetDynamicColors(true).
			SetWrap(true).
			SetWordWrap(true).
			SetTextAlign(tview.AlignCenter),
		form: tview.NewForm().
			SetButtonsAlign(tview.AlignCenter).
			SetButtonBackgroundColor(tview.Styles.PrimitiveBackgroundColor).
			SetButtonTextColor(tview.Styles.PrimaryTextColor),
	}
	dialog.frame.SetBorder(true).
		SetBorderColor(theme.BorderColor).
		SetTitleColor(theme.BorderColor).
		SetTitleAlign(tview.AlignCenter)
	dialog.textView.SetBackgroundColor(tview.Styles.ContrastBackgroundColor)
	dialog.textView.SetTextColor(theme.PrimaryTextColor)
	dialog.form.SetBackgroundColor(tview.Styles.ContrastBackgroundColor)
	dialog.form.SetBorderPadding(0, 0, 0, 0)
	dialog.form.SetCancelFunc(func() {
		if dialog.done != nil {
			dialog.done(-1)
		}
	})
	return dialog
}

func (d *Dialog) SetText(text string) *Dialog {
	d.text = text
	d.textView.SetText(text).ScrollToBeginning()
	return d
}

func (d *Dialog) SetTextAlign(align int) *Dialog {
	d.textView.SetTextAlign(align)
	return d
}

func (d *Dialog) SetTitle(title string) *Dialog {
	d.frame.SetTitle(" " + title + " ")
	return d
}

func (d *Dialog) SetAccentColor(color tcell.Color) *Dialog {
	d.frame.SetBorderColor(color).SetTitleColor(color)
	return d
}

func (d *Dialog) AddButtons(labels []string) *Dialog {
	for index, label := range labels {
		buttonIndex := index
		d.buttons = append(d.buttons, label)
		d.form.AddButton(label, func() {
			if d.done != nil {
				d.done(buttonIndex)
			}
		})
		button := d.form.GetButton(d.form.GetButtonCount() - 1)
		button.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
			// Keep scrolling separate from button selection and activation.
			if d.overflow {
				switch event.Key() {
				case tcell.KeyUp, tcell.KeyDown, tcell.KeyPgUp, tcell.KeyPgDn, tcell.KeyHome, tcell.KeyEnd:
					d.textView.InputHandler()(event, func(tview.Primitive) {})
					return nil
				default:
					// Confirmation keys must still reach the focused button.
				}
			}
			switch event.Key() {
			case tcell.KeyDown, tcell.KeyRight:
				return tcell.NewEventKey(tcell.KeyTab, 0, tcell.ModNone)
			case tcell.KeyUp, tcell.KeyLeft:
				return tcell.NewEventKey(tcell.KeyBacktab, 0, tcell.ModNone)
			default:
				return event
			}
		})
	}
	return d
}

func (d *Dialog) SetDoneFunc(handler func(int)) *Dialog {
	d.done = handler
	return d
}

func (d *Dialog) contentWidth() int {
	width := 0
	for _, label := range d.buttons {
		width += tview.TaggedStringWidth(label) + 6
	}
	if width > 0 {
		width -= 2
	}
	for _, line := range strings.Split(d.text, "\n") {
		width = max(width, tview.TaggedStringWidth(line))
	}
	return max(width, tview.TaggedStringWidth(d.frame.GetTitle()))
}

func (d *Dialog) Draw(screen tcell.Screen) {
	x, y, width, height := d.GetInnerRect()
	if width < 1 || height < 1 {
		return
	}

	dialogWidth := min(d.contentWidth()+dialogChromeWidth, width-dialogMargin)
	textWidth := max(dialogWidth-dialogChromeWidth, 1)
	dialogWidth = textWidth + dialogChromeWidth
	buttonRows := 0
	if len(d.buttons) > 0 {
		buttonRows = 2
	}
	textLines := len(tview.WordWrap(d.text, textWidth))
	d.overflow = textLines+buttonRows+2 > height
	hintRows := 0
	if d.overflow && len(d.buttons) > 0 {
		hintRows = 1
	}
	dialogHeight := min(textLines+buttonRows+hintRows+2, height)

	d.frame.SetRect(x+(width-dialogWidth)/2, y+(height-dialogHeight)/2, dialogWidth, dialogHeight)
	d.frame.Draw(screen)

	innerX, innerY, innerWidth, innerHeight := d.frame.GetInnerRect()
	if textHeight := innerHeight - buttonRows - hintRows; textHeight > 0 {
		d.textView.SetRect(innerX+1, innerY, max(innerWidth-2, 1), textHeight)
		d.textView.Draw(screen)
	}
	if hintRows > 0 && innerHeight > buttonRows {
		tview.Print(screen, "Up/Down: scroll  Left/Right: buttons", innerX, innerY+innerHeight-buttonRows-1,
			innerWidth, tview.AlignCenter, CurrentTheme().SecondaryTextColor)
	}
	if buttonRows > 0 && innerHeight > 0 {
		d.form.SetRect(innerX, innerY+innerHeight-1, innerWidth, 1)
		d.form.Draw(screen)
	}
}

func (d *Dialog) Focus(delegate func(tview.Primitive)) {
	if len(d.buttons) > 0 {
		delegate(d.form)
		return
	}
	d.Box.Focus(delegate)
}

func (d *Dialog) HasFocus() bool {
	return d.form.HasFocus() || d.Box.HasFocus()
}

func (d *Dialog) InputHandler() func(*tcell.EventKey, func(tview.Primitive)) {
	return d.WrapInputHandler(func(event *tcell.EventKey, setFocus func(tview.Primitive)) {
		if d.form.HasFocus() {
			if handler := d.form.InputHandler(); handler != nil {
				handler(event, setFocus)
			}
		}
	})
}

func (d *Dialog) MouseHandler() func(
	tview.MouseAction,
	*tcell.EventMouse,
	func(tview.Primitive),
) (bool, tview.Primitive) {
	return d.WrapMouseHandler(func(
		action tview.MouseAction,
		event *tcell.EventMouse,
		setFocus func(tview.Primitive),
	) (bool, tview.Primitive) {
		consumed, capture := d.form.MouseHandler()(action, event, setFocus)
		if !consumed && action == tview.MouseLeftDown && d.InRect(event.Position()) {
			setFocus(d)
			consumed = true
		}
		return consumed, capture
	})
}
