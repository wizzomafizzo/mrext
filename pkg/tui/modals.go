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

const (
	infoModalPage    = "tui_info_modal"
	errorModalPage   = "tui_error_modal"
	confirmModalPage = "tui_confirm_modal"
	inputModalPage   = "tui_input_modal"

	keyboardModalWidth  = 41
	keyboardModalHeight = 8
)

func ShowInfoModal(
	pages *tview.Pages,
	app *tview.Application,
	title, message string,
	onDismiss func(),
) {
	showMessageModal(pages, app, infoModalPage, title, message, CurrentTheme().BorderColor, onDismiss)
}

func ShowErrorModal(
	pages *tview.Pages,
	app *tview.Application,
	message string,
	onDismiss func(),
) {
	showMessageModal(pages, app, errorModalPage, "Error", message, CurrentTheme().ErrorColor, onDismiss)
}

func showMessageModal(
	pages *tview.Pages,
	app *tview.Application,
	page, title, message string,
	accent tcell.Color,
	onDismiss func(),
) {
	dialog := NewDialog().
		SetTitle(title).
		SetAccentColor(accent).
		SetText(message).
		AddButtons([]string{"OK"}).
		SetDoneFunc(func(int) {
			pages.RemovePage(page)
			if onDismiss != nil {
				onDismiss()
			}
		})
	pages.AddPage(page, dialog, true, true)
	app.SetFocus(dialog)
}

func ShowConfirmModal(
	pages *tview.Pages,
	app *tview.Application,
	title, message string,
	onYes, onNo func(),
) {
	dialog := NewDialog().
		SetTitle(title).
		SetText(message).
		AddButtons([]string{"Yes", "No"}).
		SetDoneFunc(func(button int) {
			pages.RemovePage(confirmModalPage)
			if button == 0 {
				if onYes != nil {
					onYes()
				}
				return
			}
			if onNo != nil {
				onNo()
			}
		})
	pages.AddPage(confirmModalPage, dialog, true, true)
	app.SetFocus(dialog)
}

// InputOptions configures ShowInputModal. Prompt is drawn above the box, so it
// should carry a rule or an example the Title does not already give: the page
// behind the modal keeps drawing, which means the row being edited and its
// footer help are still on screen next to the Title. A prompt that restates
// any of those is the fourth way of saying one thing.
//
//nolint:govet // Field order keeps textual options readable at call sites.
type InputOptions struct {
	Title            string
	Prompt           string
	InitialValue     string
	OnScreenKeyboard bool
	Validate         func(string) error
}

func ShowInputModal(
	pages *tview.Pages,
	app *tview.Application,
	options InputOptions,
	onSubmit func(string),
	onCancel func(),
) {
	cleanup := func() {
		pages.RemovePage(inputModalPage)
	}
	submit := func(value string) {
		if options.Validate != nil {
			if err := options.Validate(value); err != nil {
				options.InitialValue = value
				ShowErrorModal(pages, app, err.Error(), func() {
					ShowInputModal(pages, app, options, onSubmit, onCancel)
				})
				return
			}
		}
		cleanup()
		if onSubmit != nil {
			onSubmit(value)
		}
	}
	cancel := func() {
		cleanup()
		if onCancel != nil {
			onCancel()
		}
	}

	if options.OnScreenKeyboard {
		keyboard := NewVirtualKeyboard(options.InitialValue, submit, cancel)
		keyboard.SetTitle(" " + options.Title + " ")
		// The prompt used to be drawn only on the physical-keyboard path, so
		// the one audience that cannot see the rest of the app -- controller
		// users, which is the default on MiSTer -- never saw the hint that
		// explains what to type.
		if options.Prompt == "" {
			pages.AddPage(inputModalPage, Centered(41, 8, keyboard), true, true)
			app.SetFocus(keyboard)
			return
		}
		prompt := tview.NewTextView().
			SetText(tview.Escape(options.Prompt)).
			SetWrap(true).
			SetWordWrap(true).
			SetTextAlign(tview.AlignCenter)
		prompt.SetTextColor(CurrentTheme().SecondaryTextColor)
		promptHeight := len(tview.WordWrap(options.Prompt, keyboardModalWidth))
		content := tview.NewFlex().SetDirection(tview.FlexRow).
			AddItem(prompt, promptHeight, 0, false).
			AddItem(keyboard, keyboardModalHeight, 0, true)
		pages.AddPage(inputModalPage,
			Centered(keyboardModalWidth, keyboardModalHeight+promptHeight, content), true, true)
		app.SetFocus(keyboard)
		return
	}

	input := tview.NewInputField().
		SetText(options.InitialValue).
		SetFieldBackgroundColor(CurrentTheme().FieldFocusedBackground)
	form := tview.NewForm().
		AddFormItem(input).
		AddButton("Save", func() { submit(input.GetText()) }).
		AddButton("Cancel", cancel)
	form.SetBorder(true).SetTitle(" " + options.Title + " ")
	form.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyEscape {
			cancel()
			return nil
		}
		return event
	})
	content := tview.NewFlex().SetDirection(tview.FlexRow)
	if options.Prompt != "" {
		content.AddItem(tview.NewTextView().SetText(options.Prompt).SetWordWrap(true), 2, 0, false)
	}
	content.AddItem(form, 4, 0, true)
	pages.AddPage(inputModalPage, Centered(61, 7, content), true, true)
	app.SetFocus(input)
}
