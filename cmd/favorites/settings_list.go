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

package main

import (
	"errors"
	"path/filepath"
	"slices"
	"strings"

	"github.com/wizzomafizzo/mrext/pkg/tui"
)

func (s *settingsPage) editList(
	label string,
	values []string,
	allowEmpty bool,
	apply func([]string),
	onReturn func(),
) {
	const page = "favorites_settings_list"
	list := tui.NewMenuList()
	for _, value := range values {
		list.AddRow(value, tui.MenuRowItem)
	}
	if len(values) == 0 {
		list.AddHeader("No entries - choose Add")
	}
	back := func() {
		s.ui.pages.RemovePage(page)
		onReturn()
	}
	rebuild := func() { s.editList(label, values, allowEmpty, apply, onReturn) }
	edit := func(add bool) {
		if !add && len(values) == 0 {
			return
		}
		index := list.GetCurrentItem()
		initial := ""
		if !add {
			initial = values[index]
		}
		s.ui.showNameInput(tui.InputOptions{
			Title:        "One " + listEntryLabel(label),
			InitialValue: initial,
			Validate: func(value string) error {
				value = strings.TrimSpace(value)
				if value == "" || strings.ContainsAny(value, ",\r\n") {
					return errors.New("enter one nonempty value without commas or newlines")
				}
				if allowEmpty && !filepath.IsAbs(value) {
					return errors.New("enter a full filesystem path, such as /media/network")
				}
				return nil
			},
		}, func(value string) {
			value = strings.TrimSpace(value)
			if add {
				values = append(values, value)
			} else {
				values[index] = value
			}
			rebuild()
		}, rebuild)
	}
	remove := func() {
		if len(values) == 0 {
			return
		}
		if len(values) == 1 && !allowEmpty {
			s.ui.showError(errors.New("keep at least one folder name match"), rebuild)
			return
		}
		index := list.GetCurrentItem()
		values = slices.Delete(values, index, index+1)
		rebuild()
	}
	// Every other page gives each button a help line and pipes it to the
	// footer. This one used AddButton, so its footer never responded.
	bar := tui.NewButtonBar(s.ui.app).
		AddButtonWithHelp("Edit", "Change the selected entry", func() { edit(false) }).
		AddButtonWithHelp("Add", "Add a new entry to the list", func() { edit(true) }).
		AddButtonWithHelp("Remove", "Remove the selected entry", remove).
		AddButtonWithHelp("Done", "Keep these entries and return to Settings",
			func() { apply(slices.Clone(values)); back() }).
		AddButtonWithHelp("Cancel", "Discard entry changes and return", back).
		SetupNavigation(back)
	frame := tui.NewPageFrame(s.ui.app).
		SetTitle(appTitle, "Settings", label).
		SetContent(list).
		SetHelpText("One entry per row. Done keeps edits; Save settings writes them.").
		SetButtonBar(bar).
		SetOnEscape(back)
	bar.SetHelpCallback(func(text string) { frame.SetHelpText(text) })
	list.SetSelectedFunc(func(int, string, string, rune) { edit(false) })
	frame.SetupContentToButtonNavigation()
	s.ui.pages.AddAndSwitchToPage(page, tui.Centered(73, 13, frame), true)
	s.ui.app.SetFocus(list)
}

func listEntryLabel(label string) string {
	if label == "Folder name matches" {
		return "name fragment"
	}
	return "game folder path"
}
