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

import "github.com/wizzomafizzo/mrext/pkg/tui"

var settingsHelp = map[string]string{
	"Startup sound delay":     "Seconds before initial audio; applies next time BGM starts.",
	"Boot sounds in rotation": "Add playlist and global boot sounds to normal playback after saving.",
	"Start on boot":           "Start the BGM service automatically when MiSTer boots.",
	"Play music in cores":     "Keep playing music while a core is running instead of only in the menu.",
	"Core boot delay":         "Seconds to wait before a core boot sound plays, for slow display sync.",
	"Menu volume":             "MiSTer volume while the menu is open; needs Default volume too.",
	"Default volume":          "MiSTer volume restored when a core runs; needs Menu volume too.",
	"Debug logging":           "Write detailed service output to /tmp/bgm.log.",
}

func init() {
	// The four interface settings read the same in every app.
	for label, help := range tui.InterfaceSettingsHelp {
		settingsHelp[label] = help
	}
}
