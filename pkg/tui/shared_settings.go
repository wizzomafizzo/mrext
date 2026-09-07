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

	"github.com/rivo/tview"
	"github.com/wizzomafizzo/mrext/pkg/config"
)

// ShareInterfaceSettingsLabel names the row every settings page offers.
const ShareInterfaceSettingsLabel = "Use in all apps"

// ShareInterfaceSettingsHelp explains it in the footer.
const ShareInterfaceSettingsHelp = "Save these interface settings for every mrext app."

// ShareInterfaceSettings writes the given interface settings to the shared
// file, then removes the app's own [tui] keys so the shared file is what the
// app reads next time.
//
// Removing them is the point: an app INI that spells out theme, mouse and the
// rest always wins over the shared file, and the default INIs written on first
// run do exactly that. Without this the shared file would never take effect
// for anyone who had already run the app once.
func ShareInterfaceSettings(appINIPath string, options ApplicationOptions) error {
	if err := config.SaveSharedTUI(config.TUIConfig{
		Theme:            options.Theme,
		Mouse:            options.Mouse,
		CRTMode:          options.CRTMode,
		OnScreenKeyboard: options.OnScreenKeyboard,
	}); err != nil {
		return fmt.Errorf("save shared interface settings: %w", err)
	}
	if appINIPath == "" {
		return nil
	}
	if err := config.RemoveLocalTUIKeys(appINIPath); err != nil {
		return fmt.Errorf("stop overriding shared settings in %s: %w", appINIPath, err)
	}
	return nil
}

// ConfirmShareInterfaceSettings asks before writing, because this changes what
// other apps look like, not just this one.
func ConfirmShareInterfaceSettings(
	pages *tview.Pages, app *tview.Application, appINIPath string,
	options ApplicationOptions, onDone func(error), onCancel func(),
) {
	ShowConfirmModal(pages, app,
		"Use in all apps",
		"Save this theme and interface layout for every mrext app?\n\n"+
			"This app will follow the shared settings from now on.",
		func() { onDone(ShareInterfaceSettings(appINIPath, options)) },
		onCancel,
	)
}
