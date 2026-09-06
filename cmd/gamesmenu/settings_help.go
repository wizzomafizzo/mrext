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

func settingHelp(label string) string {
	switch label {
	case "Theme":
		return "Choose interface colors; changes apply after Save"
	case "Mouse":
		return "Allow mouse navigation; controller navigation always works"
	case "CRT mode":
		return "Limit the interface to a centered 75 x 15 area on larger screens"
	default:
		return ""
	}
}
