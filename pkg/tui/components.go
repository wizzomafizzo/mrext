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
	"github.com/rivo/tview"
)

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

type ProgressUpdate struct {
	Text    string
	Current int
	Total   int
}
