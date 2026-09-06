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
	"strconv"

	"github.com/wizzomafizzo/mrext/pkg/bgm"
	"github.com/wizzomafizzo/mrext/pkg/tui"
)

const pageSettings = "bgm_settings"

const (
	menuVolumeHelp = "Automatically change MiSTer's volume in the menu when it is open, to adjust music " +
		"volume. The \"default volume\" setting must also be enabled for this feature to work."
	defaultVolumeHelp = "Automatically revert MiSTer's volume when the menu is closed, to set volume back " +
		"to normal. The \"menu volume\" setting must also be enabled for this feature to work."
)

//nolint:govet // Field order groups staged values and page runtime state.
type settingsPage struct {
	staged   bgm.Settings
	original bgm.Settings
	actions  map[int]func()
	help     map[int]string
	list     *tui.MenuList
	ui       *ui
}

func (u *ui) startSettings() {
	page := &settingsPage{
		staged:   bgm.SettingsFromConfig(&u.cfg),
		original: bgm.SettingsFromConfig(&u.cfg),
		ui:       u,
	}
	page.show(0)
}

func (s *settingsPage) show(selection int) {
	s.actions = make(map[int]func())
	s.help = make(map[int]string)
	s.list = tui.NewMenuList()
	s.list.AddHeader("Playback")
	s.addToggle("Start on boot", &s.staged.Startup)
	s.addToggle("Play music in cores", &s.staged.PlayInCore)
	s.addText("Core boot delay", bgm.FormatDelay(s.staged.CoreBootDelay), func(value string) error {
		if _, err := bgm.ParseDelay(value); err != nil {
			return fmt.Errorf("invalid core boot delay: %w", err)
		}
		return nil
	}, func(value string) {
		if parsed, err := bgm.ParseDelay(value); err == nil {
			s.staged.CoreBootDelay = parsed
		}
	})

	s.addText("Startup sound delay", bgm.FormatDelay(s.staged.BootDelay), func(value string) error {
		if _, err := bgm.ParseBootDelay(value); err != nil {
			return fmt.Errorf("invalid startup delay: %w", err)
		}
		return nil
	}, func(value string) {
		if parsed, err := bgm.ParseBootDelay(value); err == nil {
			s.staged.BootDelay = parsed
		}
	})

	s.list.AddHeader("Volume")
	s.addVolume("Menu volume", menuVolumeHelp, &s.staged.MenuVolume, true)
	s.addVolume("Default volume", defaultVolumeHelp, &s.staged.DefaultVolume, false)

	s.list.AddHeader("Logging")
	s.addToggle("Debug logging", &s.staged.Debug)

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
		SetTitle(appTitle, "Settings").
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

func (s *settingsPage) addText(label, value string, validate func(string) error, apply func(string)) {
	s.addRow(label, value, func() {
		selection := s.list.GetCurrentItem()
		options := tui.InputOptions{
			Title:            "Edit " + label,
			Prompt:           "Enter the number of seconds to wait, such as 0 or 1.5.",
			InitialValue:     value,
			OnScreenKeyboard: s.ui.options.OnScreenKeyboard,
			Validate:         validate,
		}
		tui.ShowInputModal(s.ui.pages, s.ui.app, options, func(updated string) {
			apply(updated)
			s.show(selection)
		}, func() { s.show(selection) })
	})
}

// addVolume offers the same nine levels as the Python dialog. The menu
// volume is applied live when chosen, exactly as before.
func (s *settingsPage) addVolume(label, help string, value *int, live bool) {
	choices := make([]tui.Choice, 0, 9)
	choices = append(choices,
		tui.Choice{Label: "Disabled (make no changes to volume)", Help: help},
		tui.Choice{Label: "Mute", Help: help},
	)
	for level := 1; level <= 7; level++ {
		choice := tui.Choice{Label: "Level " + strconv.Itoa(level), Help: help}
		if level == 7 {
			choice.Label += " (max)"
		}
		choices = append(choices, choice)
	}
	s.addChoice(label, volumeLabel(*value), choices, *value+1, func(index int) {
		*value = index - 1
		if live && *value >= 0 {
			bgm.VolumeSet(&s.ui.paths, s.ui.logger, *value)
		}
	})
}

func volumeLabel(volume int) string {
	switch {
	case volume < 0:
		return "Disabled"
	case volume == 0:
		return "Mute"
	default:
		return "Level " + strconv.Itoa(volume)
	}
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
	if err := bgm.SaveSettings(s.ui.paths.IniFile, &s.staged); err != nil {
		s.ui.showError(err, func() { s.show(s.list.GetCurrentItem()) })
		return
	}
	// The file is written, so the in-memory configuration follows it even
	// when the running service cannot be told about the change.
	s.apply()
	if s.staged.PlayInCore != s.original.PlayInCore {
		message := "set playincore no"
		if s.staged.PlayInCore {
			message = "set playincore yes"
		}
		if _, _, err := s.ui.send(message); err != nil {
			s.ui.showError(err, s.ui.mustShowMain)
			return
		}
	}
	s.ui.mustShowMain()
}

func (s *settingsPage) apply() {
	s.staged.ApplyTo(&s.ui.cfg)
	s.ui.options = s.ui.applicationOptions()
	_ = tui.SetCurrentTheme(s.ui.options.Theme)
	s.ui.app.EnableMouse(s.ui.options.Mouse)
	s.ui.app.SetRoot(tui.WrapRoot(s.ui.options, s.ui.pages), true)
}

func themeLabel(name string) string {
	if theme := tui.AvailableThemes[name]; theme != nil {
		return theme.DisplayName
	}
	return name
}
