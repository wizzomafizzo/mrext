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
	"fmt"
	"strings"
	"sync"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

type ListPickerOpts struct {
	Title         string
	Buttons       []string
	DefaultButton int
	ActionButton  int
	ShowTotal     bool
	Width         int
	Height        int
}

func Centered(width, height int, primitive tview.Primitive) tview.Primitive {
	row := tview.NewFlex().
		AddItem(nil, 0, 1, false).
		AddItem(primitive, width, 1, true).
		AddItem(nil, 0, 1, false)
	return tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(nil, 0, 1, false).
		AddItem(row, height, 1, true).
		AddItem(nil, 0, 1, false)
}

func ButtonBar(buttons []string, selected int) string {
	parts := make([]string, len(buttons))
	for i, button := range buttons {
		label := "< " + tview.Escape(button) + " >"
		if i == selected {
			parts[i] = "[black:white]" + label + "[-:-]"
		} else {
			parts[i] = label
		}
	}
	return strings.Join(parts, "    ")
}

func ListPicker(opts *ListPickerOpts, items []string) (button, item int, err error) {
	selectedButton := opts.DefaultButton
	selectedItem := 0
	resultButton := -1
	resultItem := -1

	builder := func() (*tview.Application, error) {
		app := tview.NewApplication()
		list := tview.NewList().ShowSecondaryText(false)
		list.SetHighlightFullLine(true)
		for _, item := range items {
			list.AddItem(item, "", 0, nil)
		}
		if len(items) > 0 {
			list.SetCurrentItem(selectedItem)
		}

		footer := tview.NewTextView().SetDynamicColors(true).SetTextAlign(tview.AlignCenter)
		drawFooter := func() {
			text := ButtonBar(opts.Buttons, selectedButton)
			if opts.ShowTotal && len(items) > 0 {
				text = fmt.Sprintf("%d/%d    %s", list.GetCurrentItem()+1, len(items), text)
			}
			footer.SetText(text)
		}
		drawFooter()

		content := tview.NewFlex().SetDirection(tview.FlexRow).
			AddItem(list, 0, 1, true).
			AddItem(footer, 1, 0, false)
		content.SetBorder(true).SetTitle(" " + opts.Title + " ")

		app.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
			switch event.Key() {
			case tcell.KeyEscape:
				app.Stop()
				return nil
			case tcell.KeyLeft:
				if len(opts.Buttons) > 0 {
					selectedButton = (selectedButton - 1 + len(opts.Buttons)) % len(opts.Buttons)
					drawFooter()
				}
				return nil
			case tcell.KeyRight:
				if len(opts.Buttons) > 0 {
					selectedButton = (selectedButton + 1) % len(opts.Buttons)
					drawFooter()
				}
				return nil
			case tcell.KeyPgUp:
				selectedItem = list.GetCurrentItem() - max(opts.Height-5, 1)
				if selectedItem < 0 {
					selectedItem = 0
				}
				list.SetCurrentItem(selectedItem)
				drawFooter()
				return nil
			case tcell.KeyPgDn:
				selectedItem = list.GetCurrentItem() + max(opts.Height-5, 1)
				if selectedItem >= len(items) {
					selectedItem = len(items) - 1
				}
				if selectedItem >= 0 {
					list.SetCurrentItem(selectedItem)
				}
				drawFooter()
				return nil
			case tcell.KeyEnter:
				if len(opts.Buttons) == 0 {
					return nil
				}
				buttonLabel := opts.Buttons[selectedButton]
				switch buttonLabel {
				case "PgUp":
					selectedItem = max(list.GetCurrentItem()-max(opts.Height-5, 1), 0)
					list.SetCurrentItem(selectedItem)
					drawFooter()
					return nil
				case "PgDn":
					selectedItem = min(list.GetCurrentItem()+max(opts.Height-5, 1), len(items)-1)
					if selectedItem >= 0 {
						list.SetCurrentItem(selectedItem)
					}
					drawFooter()
					return nil
				default:
					resultButton = selectedButton
					if selectedButton == opts.ActionButton && len(items) > 0 {
						resultItem = list.GetCurrentItem()
					}
					app.Stop()
					return nil
				}
			case tcell.KeyUp, tcell.KeyDown:
				defer drawFooter()
				return event
			default:
				return event
			}
		})
		return app.SetRoot(Centered(opts.Width, opts.Height, content), true).SetFocus(list), nil
	}

	if err := BuildAndRetry(builder); err != nil {
		return -1, -1, err
	}
	return resultButton, resultItem, nil
}

func InfoBox(title, text string) error {
	builder := func() (*tview.Application, error) {
		app := tview.NewApplication()
		modal := tview.NewModal().
			SetText(text).
			AddButtons([]string{"OK"}).
			SetDoneFunc(func(_ int, _ string) { app.Stop() })
		modal.SetTitle(" " + title + " ").SetBorder(true)
		return app.SetRoot(modal, true).SetFocus(modal), nil
	}
	return BuildAndRetry(builder)
}

type ProgressUpdate struct {
	Text    string
	Current int
	Total   int
}

func RunProgress(title string, initial ProgressUpdate, task func(update func(ProgressUpdate)) error) error {
	var once sync.Once
	result := make(chan error, 1)

	builder := func() (*tview.Application, error) {
		app := tview.NewApplication()
		text := tview.NewTextView().SetTextAlign(tview.AlignCenter)
		text.SetBorder(true).SetTitle(" " + title + " ")
		render := func(update ProgressUpdate) {
			message := update.Text
			if update.Total > 0 {
				current := min(max(update.Current, 0), update.Total)
				width := 50
				filled := current * width / update.Total
				message += "\n\n[" + strings.Repeat("#", filled) + strings.Repeat("-", width-filled) + "]"
			}
			text.SetText(message)
		}
		render(initial)

		app.SetAfterDrawFunc(func(_ tcell.Screen) {
			once.Do(func() {
				go func() {
					result <- task(func(update ProgressUpdate) {
						app.QueueUpdateDraw(func() { render(update) })
					})
					app.Stop()
				}()
			})
		})
		return app.SetRoot(Centered(60, 7, text), true), nil
	}

	if err := BuildAndRetry(builder); err != nil {
		return err
	}
	return <-result
}
