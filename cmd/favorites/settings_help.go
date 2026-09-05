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

var settingsHelp = map[string]string{
	"Default folder":          "Folder created at startup when no Favorites folders exist.",
	"Folder name matches":     "Name fragments used to find Favorites folders, such as fav.",
	"Create default folder":   "Automatically create the default folder at startup if needed.",
	"Arcade core links":       "Maintain cores links so arcade favorites can find their RBFs.",
	"Hide root files":         "Show only familiar MiSTer folders at the SD browser root.",
	"USB shortcut folder":     "Folder opened by the USB shortcut, normally /media/usb0.",
	"Core prefix":             "Prefix standard RBF paths for custom setups; normally leave empty.",
	"Additional game folders": "Extra game-library roots to recognize alongside SD and USB roots.",
	"Alternate core":          "Choose Standard, LLAPI, or YC for newly created game favorites.",
	"Theme":                   "Choose a color palette to apply after saving.",
	"Mouse":                   "Enable mouse input alongside keyboard and controller navigation.",
	"CRT mode":                "Use a compact 75-column, 15-row layout for CRT displays.",
	"On-screen keyboard":      "Use controller-friendly text entry; Off needs a physical keyboard.",
}
