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
	"github.com/wizzomafizzo/mrext/pkg/gamesmenu"
	"github.com/wizzomafizzo/mrext/pkg/tui"
)

const pageSettings = "gamesmenu_settings"

//nolint:govet // Groups staged values and widgets.
type settingsPage struct {
	ui               *ui
	original, staged gamesmenu.Settings
	list             *tui.MenuList
	actions          []func()
	help             []string
}

func (u *ui) startSettings() { newSettingsPage(u).show(1) }

func newSettingsPage(u *ui) *settingsPage {
	settings := gamesmenu.SettingsFromConfig(u.cfg)
	return &settingsPage{ui: u, original: settings, staged: settings}
}

func (s *settingsPage) show(selection int) {
	s.list = tui.NewMenuList()
	s.actions, s.help = nil, nil
	s.list.AddHeader("Interface")
	s.actions = append(s.actions, nil)
	s.help = append(s.help, "Controller-friendly interface and display options")
	// tui.ThemeNames, not sorted map keys: the picker has to offer the themes
	// in the same order here as it does in BGM and Favorites.
	s.addRow("Theme", tui.ThemeLabel(s.staged.Theme), func() {
		choices := make([]tui.Choice, len(tui.ThemeNames))
		selected := 0
		for index, name := range tui.ThemeNames {
			choices[index] = tui.Choice{Label: tui.ThemeLabel(name), Help: "Applied after saving settings."}
			if name == s.staged.Theme {
				selected = index
			}
		}
		tui.ShowChoiceModal(s.ui.pages, s.ui.app, "Theme", choices, selected, func(index int) {
			s.staged.Theme = tui.ThemeNames[index]
			s.show(1)
		}, func() { s.show(1) })
	})
	s.addRow("Mouse", tui.BoolLabel(s.staged.Mouse), func() { s.staged.Mouse = !s.staged.Mouse; s.show(2) })
	s.addRow("CRT mode", tui.BoolLabel(s.staged.CRTMode), func() { s.staged.CRTMode = !s.staged.CRTMode; s.show(3) })
	// GamesMenu has no text entry, so it has no on-screen keyboard setting;
	// sharing still carries the value it inherited so other apps keep theirs.
	s.addRow(tui.ShareInterfaceSettingsLabel, "", func() {
		row := s.list.GetCurrentItem()
		options := tui.ApplicationOptions{
			Theme:            s.staged.Theme,
			Mouse:            s.staged.Mouse,
			CRTMode:          s.staged.CRTMode,
			OnScreenKeyboard: s.ui.cfg.TUI.OnScreenKeyboard,
		}
		tui.ConfirmShareInterfaceSettings(s.ui.pages, s.ui.app, s.ui.cfg.IniPath, options,
			func(err error) {
				if err != nil {
					s.ui.showError(err, func() { s.show(row) })
					return
				}
				s.show(row)
			},
			func() { s.show(row) })
	})
	activate := func() {
		if action := s.actions[s.list.GetCurrentItem()]; action != nil {
			action()
		}
	}
	s.list.SetSelectedFunc(func(int, string, string, rune) { activate() })
	bar := tui.NewButtonBar(s.ui.app).
		AddButtonWithHelp("Change", "Change selected setting", activate).
		AddButtonWithHelp("Save", "Save settings and return", s.save).
		AddButtonWithHelp("Cancel", "Discard changes and return", s.back).SetupNavigation(s.back)
	frame := tui.NewPageFrame(s.ui.app).SetTitle(appTitle, "Settings").
		SetContent(s.list).SetButtonBar(bar).SetOnEscape(s.back)
	bar.SetHelpCallback(func(text string) { frame.SetHelpText(text) })
	s.list.SetRowChangedFunc(func(index int) {
		if index >= 0 && index < len(s.help) {
			frame.SetHelpText(s.help[index])
		}
	})
	frame.SetupContentToButtonNavigation()
	s.list.SetCurrentItem(min(selection, s.list.GetItemCount()-1))
	frame.SetHelpText(s.help[s.list.GetCurrentItem()])
	s.ui.pages.AddAndSwitchToPage(pageSettings, frame, true)
	s.ui.app.SetFocus(s.list)
}

func (s *settingsPage) addRow(label, value string, action func()) {
	text := label
	if value != "" {
		text = label + ": " + value
	}
	s.list.AddRow(text, tui.MenuRowAction)
	s.actions = append(s.actions, action)
	s.help = append(s.help, settingHelp(label))
}

func (s *settingsPage) back() {
	if s.staged.Equal(&s.original) {
		s.ui.renderMain()
		return
	}
	tui.ShowConfirmModal(
		s.ui.pages, s.ui.app, "Discard Settings", "Discard unsaved settings?",
		s.ui.renderMain, func() { s.show(s.list.GetCurrentItem()) },
	)
}

func (s *settingsPage) save() {
	if err := gamesmenu.SaveSettings(s.ui.cfg.IniPath, &s.staged); err != nil {
		s.ui.showError(err, func() { s.show(s.list.GetCurrentItem()) })
		return
	}
	s.staged.ApplyTo(s.ui.cfg)
	s.ui.options = tui.ApplicationOptions{
		Theme: s.ui.cfg.TUI.Theme, Mouse: s.ui.cfg.TUI.Mouse, CRTMode: s.ui.cfg.TUI.CRTMode,
	}
	tui.ApplyOptions(s.ui.app, s.ui.pages, s.ui.options)
	s.ui.renderMain()
}
