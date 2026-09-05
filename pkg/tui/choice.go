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

import "github.com/rivo/tview"

type Choice struct {
	Label string
	Help  string
}

// ShowChoiceModal selects a value without applying changes on cancellation.
func ShowChoiceModal(
	pages *tview.Pages,
	app *tview.Application,
	title string,
	choices []Choice,
	selected int,
	onSelect func(int),
	onCancel func(),
) {
	const page = "tui_choice_modal"
	list := NewMenuList()
	for _, choice := range choices {
		list.AddRow(choice.Label, MenuRowItem)
	}
	list.SetCurrentItem(selected)
	cancel := func() {
		pages.RemovePage(page)
		if onCancel != nil {
			onCancel()
		}
	}
	selectChoice := func() {
		index := list.GetCurrentItem()
		if index < 0 || index >= len(choices) {
			return
		}
		pages.RemovePage(page)
		onSelect(index)
	}
	bar := NewButtonBar(app).
		AddButton("Select", selectChoice).
		AddButton("Cancel", cancel).
		SetupNavigation(cancel)
	frame := NewPageFrame(app).
		SetTitle(title).
		SetContent(list).
		SetButtonBar(bar).
		SetOnEscape(cancel)
	help := func(index int) {
		if index >= 0 && index < len(choices) {
			frame.SetHelpText(choices[index].Help)
		}
	}
	list.SetRowChangedFunc(help)
	help(list.GetCurrentItem())
	list.SetSelectedFunc(func(int, string, string, rune) { selectChoice() })
	frame.SetupContentToButtonNavigation()
	pages.AddPage(page, Centered(69, min(len(choices)+4, 13), frame), true, true)
	app.SetFocus(list)
}
