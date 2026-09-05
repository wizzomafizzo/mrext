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
	"fmt"
	"slices"
	"strings"

	"github.com/wizzomafizzo/mrext/pkg/favorites"
	"github.com/wizzomafizzo/mrext/pkg/tui"
)

const pageSettings = "favorites_settings"

//nolint:govet // Field order groups staged values and page runtime state.
type settingsPage struct {
	staged   favorites.Settings
	original favorites.Settings
	actions  map[int]func()
	help     map[int]string
	list     *tui.MenuList
	ui       *ui
}

func (u *ui) startSettings() {
	page := &settingsPage{
		staged:   favorites.SettingsFromConfig(u.cfg),
		original: favorites.SettingsFromConfig(u.cfg),
		ui:       u,
	}
	page.show(0)
}

func (s *settingsPage) show(selection int) {
	s.actions = make(map[int]func())
	s.help = make(map[int]string)
	s.list = tui.NewMenuList()
	s.list.AddHeader("Favorites")
	s.addText("Default folder", s.staged.DefaultFolder, favorites.ValidateFolderName, func(value string) {
		s.staged.DefaultFolder = value
	})
	s.addList("Folder name matches", s.staged.FolderNameContains, false, func(values []string) {
		s.staged.FolderNameContains = values
	})
	s.addToggle("Create default folder", &s.staged.CreateDefaultFolder)
	s.addToggle("Arcade core links", &s.staged.ManageArcadeCoreLinks)
	s.addToggle("Hide root files", &s.staged.HideRootFiles)
	s.addText("USB shortcut folder", s.staged.ExternalFolder, nil, func(value string) {
		s.staged.ExternalFolder = value
	})
	s.addText("Core prefix", s.staged.CorePrefix, nil, func(value string) { s.staged.CorePrefix = value })
	s.addList("Additional game folders", s.staged.GamesFolders, true, func(values []string) {
		s.staged.GamesFolders = values
	})

	s.list.AddHeader("Cores")
	modes := []string{"", "llapi", "yc"}
	s.addChoice("Alternate core", alternateCoreLabel(s.staged.AlternateCore), []tui.Choice{
		{Label: "Standard", Help: "Use standard MiSTer cores."},
		{Label: "LLAPI", Help: "Use installed LLAPI cores; otherwise use standard cores."},
		{Label: "YC", Help: "Use installed YC cores; otherwise use standard cores."},
	}, slices.Index(modes, s.staged.AlternateCore), func(index int) {
		s.staged.AlternateCore = modes[index]
	})

	s.list.AddHeader("Interface")
	themes := make([]tui.Choice, 0, len(tui.ThemeNames))
	for _, name := range tui.ThemeNames {
		themes = append(themes, tui.Choice{Label: themeLabel(name), Help: "Applied after saving settings."})
	}
	themeIndex := slices.Index(tui.ThemeNames, s.staged.Theme)
	s.addChoice("Theme", themeLabel(s.staged.Theme), themes, themeIndex, func(index int) {
		s.staged.Theme = tui.ThemeNames[index]
	})
	s.addToggle("Mouse", &s.staged.Mouse)
	s.addToggle("CRT mode", &s.staged.CRTMode)
	s.addToggle("On-screen keyboard", &s.staged.OnScreenKeyboard)

	s.list.SetCurrentItem(selection)
	change := func() {
		if action := s.actions[s.list.GetCurrentItem()]; action != nil {
			action()
		}
	}
	s.list.SetSelectedFunc(func(int, string, string, rune) { change() })
	back := func() { s.back() }
	bar := tui.NewButtonBar(s.ui.app).
		AddButtonWithHelp("Change", "Change selected setting", change).
		AddButtonWithHelp("Save", "Save settings and return", s.save).
		AddButtonWithHelp("Cancel", "Discard changes and return", back).
		SetupNavigation(back)
	frame := tui.NewPageFrame(s.ui.app).
		SetTitle("Favorites Manager", "Settings").
		SetContent(s.list).
		SetHelpText("Change settings, then save or cancel.").
		SetButtonBar(bar).
		SetOnEscape(back)
	showHelp := func(index int) { frame.SetHelpText(s.help[index]) }
	s.list.SetRowChangedFunc(showHelp)
	showHelp(s.list.GetCurrentItem())
	frame.SetupContentToButtonNavigation()
	s.ui.pages.AddAndSwitchToPage(pageSettings, frame, true)
	s.ui.app.SetFocus(s.list)
}

func (s *settingsPage) addRow(label, value string, action func()) {
	if value == "" {
		value = "(none)"
	}
	index := s.list.GetItemCount()
	s.list.AddRow(label+": "+value, tui.MenuRowAction)
	s.actions[index] = action
	s.help[index] = settingsHelp[label]
}

func (s *settingsPage) addToggle(label string, value *bool) {
	display := "Off"
	if *value {
		display = "On"
	}
	s.addRow(label, display, func() {
		*value = !*value
		s.show(s.list.GetCurrentItem())
	})
}

func (s *settingsPage) addChoice(label, value string, choices []tui.Choice, selected int, apply func(int)) {
	s.addRow(label, value, func() {
		selection := s.list.GetCurrentItem()
		tui.ShowChoiceModal(s.ui.pages, s.ui.app, label, choices, max(selected, 0), func(index int) {
			apply(index)
			s.show(selection)
		}, func() { s.show(selection) })
	})
}

func (s *settingsPage) addText(
	label, value string,
	validate func(string) error,
	apply func(string),
) {
	s.addRow(label, value, func() {
		selection := s.list.GetCurrentItem()
		s.ui.showNameInput(tui.InputOptions{
			Title:        "Edit " + label,
			Prompt:       "Enter " + strings.ToLower(label) + ".",
			InitialValue: value,
			Validate:     validate,
		}, func(updated string) {
			apply(strings.TrimSpace(updated))
			s.show(selection)
		}, func() { s.show(selection) })
	})
}

func (s *settingsPage) addList(
	label string,
	values []string,
	allowEmpty bool,
	apply func([]string),
) {
	s.addRow(label, fmt.Sprintf("%d entries (list)", len(values)), func() {
		selection := s.list.GetCurrentItem()
		s.editList(label, slices.Clone(values), allowEmpty, apply, func() { s.show(selection) })
	})
}

func (s *settingsPage) back() {
	if s.staged.Equal(&s.original) {
		s.ui.mustShowMain()
		return
	}
	tui.ShowConfirmModal(
		s.ui.pages,
		s.ui.app,
		"Discard Settings",
		"Discard unsaved settings?",
		s.ui.mustShowMain,
		func() { s.show(s.list.GetCurrentItem()) },
	)
}

func (s *settingsPage) save() {
	if _, exists := tui.AvailableThemes[s.staged.Theme]; !exists {
		s.ui.showError(fmt.Errorf("unknown TUI theme: %s", s.staged.Theme), func() {
			s.show(s.list.GetCurrentItem())
		})
		return
	}
	if err := favorites.SaveSettings(s.ui.cfg.IniPath, &s.staged); err != nil {
		s.ui.showError(err, func() { s.show(s.list.GetCurrentItem()) })
		return
	}
	s.apply()
	if err := s.ui.manager.SetupArcadeLinks(); err != nil {
		s.ui.showError(err, s.ui.mustShowMain)
		return
	}
	s.ui.mustShowMain()
}

func (s *settingsPage) apply() {
	gamesFolders := slices.Clone(s.ui.cfg.Systems.GamesFolder)
	for _, oldFolder := range s.original.GamesFolders {
		gamesFolders = slices.DeleteFunc(gamesFolders, func(folder string) bool { return folder == oldFolder })
	}
	s.staged.ApplyTo(s.ui.cfg)
	gamesFolders = append(gamesFolders, s.staged.GamesFolders...)
	s.ui.cfg.Systems.GamesFolder = gamesFolders
	s.ui.options = tui.ApplicationOptions{
		Theme:            s.ui.cfg.TUI.Theme,
		Mouse:            s.ui.cfg.TUI.Mouse,
		CRTMode:          s.ui.cfg.TUI.CRTMode,
		OnScreenKeyboard: s.ui.cfg.TUI.OnScreenKeyboard,
	}
	_ = tui.SetCurrentTheme(s.ui.options.Theme)
	s.ui.app.EnableMouse(s.ui.options.Mouse)
	s.ui.app.SetRoot(tui.WrapRoot(s.ui.options, s.ui.pages), true)
}

func alternateCoreLabel(mode string) string {
	switch mode {
	case "llapi":
		return "LLAPI"
	case "yc":
		return "YC"
	default:
		return "Standard"
	}
}

func themeLabel(name string) string {
	if theme := tui.AvailableThemes[name]; theme != nil {
		return theme.DisplayName
	}
	return name
}
