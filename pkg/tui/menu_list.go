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

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

type MenuRowKind int

const (
	MenuRowItem MenuRowKind = iota
	MenuRowFolder
	MenuRowAction
	MenuRowAdd
	MenuRowHeader
)

type menuRow struct {
	label string
	kind  MenuRowKind
}

type MenuList struct {
	*tview.List
	onRowChanged func(int)
	rows         []menuRow
	selected     int
}

func NewMenuList() *MenuList {
	list := tview.NewList().ShowSecondaryText(false)
	list.SetWrapAround(false)
	list.SetHighlightFullLine(false)
	list.SetSelectedStyle(tcell.StyleDefault)
	list.SetMainTextStyle(tcell.StyleDefault)
	list.SetMainTextColor(CurrentTheme().PrimaryTextColor)
	menu := &MenuList{List: list, selected: -1}
	list.SetChangedFunc(func(index int, _, _ string, _ rune) {
		if menu.selectable(index) {
			previous := menu.selected
			menu.selected = index
			menu.refreshRow(previous)
			menu.refreshRow(index)
			if menu.onRowChanged != nil {
				menu.onRowChanged(index)
			}
		}
	})
	list.SetInputCapture(menu.captureInput)
	return menu
}

func (m *MenuList) AddRow(label string, kind MenuRowKind) *MenuList {
	row := menuRow{label: label, kind: kind}
	m.rows = append(m.rows, row)
	m.AddItem(formatMenuRow(row, false), "", 0, nil)
	if m.selected < 0 && kind != MenuRowHeader {
		m.List.SetCurrentItem(len(m.rows) - 1)
	}
	return m
}

func (m *MenuList) AddHeader(label string) *MenuList {
	return m.AddRow(label, MenuRowHeader)
}

// SetRowChangedFunc adds a notification without replacing selection styling.
func (m *MenuList) SetRowChangedFunc(callback func(int)) *MenuList {
	m.onRowChanged = callback
	return m
}

func (m *MenuList) NavigationList() *tview.List {
	return m.List
}

func (m *MenuList) SetCurrentItem(index int) *MenuList {
	if !m.selectable(index) {
		if target := m.scanSelectable(index, 1); target >= 0 {
			index = target
		} else if target := m.scanSelectable(index, -1); target >= 0 {
			index = target
		} else {
			return m
		}
	}
	m.List.SetCurrentItem(index)
	return m
}

func (m *MenuList) MouseHandler() func(
	tview.MouseAction,
	*tcell.EventMouse,
	func(tview.Primitive),
) (bool, tview.Primitive) {
	return m.WrapMouseHandler(func(
		action tview.MouseAction,
		event *tcell.EventMouse,
		setFocus func(tview.Primitive),
	) (bool, tview.Primitive) {
		if action == tview.MouseLeftDown || action == tview.MouseLeftClick {
			_, top, _, height := m.GetInnerRect()
			_, mouseY := event.Position()
			if mouseY >= top && mouseY < top+height {
				itemOffset, _ := m.GetOffset()
				if !m.selectable(itemOffset + mouseY - top) {
					return true, nil
				}
			}
		}
		return m.List.MouseHandler()(action, event, setFocus)
	})
}

func (m *MenuList) selectable(index int) bool {
	return index >= 0 && index < len(m.rows) && m.rows[index].kind != MenuRowHeader
}

func (m *MenuList) scanSelectable(start, direction int) int {
	for index := start; index >= 0 && index < len(m.rows); index += direction {
		if m.selectable(index) {
			return index
		}
	}
	return -1
}

func (m *MenuList) captureInput(event *tcell.EventKey) *tcell.EventKey {
	switch event.Key() {
	case tcell.KeyUp:
		if target := m.scanSelectable(m.GetCurrentItem()-1, -1); target >= 0 {
			m.List.SetCurrentItem(target)
		}
		return nil
	case tcell.KeyDown:
		if target := m.scanSelectable(m.GetCurrentItem()+1, 1); target >= 0 {
			m.List.SetCurrentItem(target)
		}
		return nil
	case tcell.KeyPgUp, tcell.KeyPgDn:
		_, _, _, height := m.GetInnerRect()
		direction := 1
		if event.Key() == tcell.KeyPgUp {
			direction = -1
		}
		start := max(0, min(m.GetCurrentItem()+direction*max(height, 1), len(m.rows)-1))
		target := m.scanSelectable(start, direction)
		if target < 0 {
			target = m.scanSelectable(start, -direction)
		}
		if target >= 0 {
			m.List.SetCurrentItem(target)
		}
		return nil
	case tcell.KeyHome:
		if target := m.scanSelectable(0, 1); target >= 0 {
			m.List.SetCurrentItem(target)
		}
		return nil
	case tcell.KeyEnd:
		if target := m.scanSelectable(len(m.rows)-1, -1); target >= 0 {
			m.List.SetCurrentItem(target)
		}
		return nil
	default:
		return event
	}
}

func (m *MenuList) refreshRow(index int) {
	if index >= 0 && index < len(m.rows) {
		m.SetItemText(index, formatMenuRow(m.rows[index], index == m.selected), "")
	}
}

func formatMenuRow(row menuRow, selected bool) string {
	theme := CurrentTheme()
	label := tview.Escape(row.label)
	if row.kind == MenuRowHeader {
		return fmt.Sprintf("[%s:%s:b]-- %s --[-:-:-]", theme.SecondaryName, theme.BackgroundName, label)
	}

	prefix := "  "
	switch row.kind {
	case MenuRowFolder, MenuRowAction:
		prefix = "> "
	case MenuRowAdd:
		prefix = "+ "
	case MenuRowItem, MenuRowHeader:
	}
	if selected {
		return fmt.Sprintf(
			"[%s:%s:b]%s[%s:%s:-]%s[-:-:-]",
			theme.AccentName,
			theme.BackgroundName,
			prefix,
			theme.HighlightForegroundName,
			theme.HighlightBackgroundName,
			label,
		)
	}
	attributes := ""
	if row.kind == MenuRowAction || row.kind == MenuRowAdd {
		attributes = "b"
	}
	return fmt.Sprintf(
		"[%s:%s:%s]%s[%s:%s]%s[-:-:-]",
		theme.AccentName,
		theme.BackgroundName,
		attributes,
		prefix,
		theme.TextName,
		theme.BackgroundName,
		label,
	)
}
