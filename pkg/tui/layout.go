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

// Portions adapted from Zaparoo Core, Copyright (c) 2026 The Zaparoo Project Contributors.

package tui

import (
	"fmt"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

const (
	DefaultMaxWidth  = 100
	DefaultMaxHeight = 30
	CRTWidth         = 75
	CRTHeight        = 15
)

type ApplicationOptions struct {
	Theme            string
	Mouse            bool
	CRTMode          bool
	OnScreenKeyboard bool
}

func NewApplication(options ApplicationOptions) (*tview.Application, error) {
	if !SetCurrentTheme(options.Theme) {
		return nil, fmt.Errorf("unknown TUI theme: %s", options.Theme)
	}
	app := tview.NewApplication()
	app.EnableMouse(options.Mouse)
	return app, nil
}

func WrapRoot(options ApplicationOptions, primitive tview.Primitive) tview.Primitive {
	if options.CRTMode {
		return Centered(CRTWidth, CRTHeight, primitive)
	}
	return ResponsiveMaxWidget(DefaultMaxWidth, DefaultMaxHeight, primitive)
}

type responsiveWrapper struct {
	*tview.Box
	child     tview.Primitive
	maxWidth  int
	maxHeight int
}

func ResponsiveMaxWidget(maxWidth, maxHeight int, primitive tview.Primitive) tview.Primitive {
	return &responsiveWrapper{
		Box:       tview.NewBox(),
		child:     primitive,
		maxWidth:  maxWidth,
		maxHeight: maxHeight,
	}
}

func (r *responsiveWrapper) Draw(screen tcell.Screen) {
	x, y, width, height := r.GetInnerRect()
	actualWidth := min(width, r.maxWidth)
	actualHeight := min(height, r.maxHeight)
	r.child.SetRect(x+(width-actualWidth)/2, y+(height-actualHeight)/2, actualWidth, actualHeight)
	r.child.Draw(screen)
}

func (r *responsiveWrapper) Focus(delegate func(tview.Primitive)) {
	delegate(r.child)
}

func (r *responsiveWrapper) HasFocus() bool {
	return r.child.HasFocus()
}

func (r *responsiveWrapper) InputHandler() func(*tcell.EventKey, func(tview.Primitive)) {
	return r.child.InputHandler()
}

func (r *responsiveWrapper) MouseHandler() func(
	tview.MouseAction,
	*tcell.EventMouse,
	func(tview.Primitive),
) (bool, tview.Primitive) {
	return r.child.MouseHandler()
}

// BoolLabel renders a toggle. One spelling across every app: GamesMenu used to
// say Yes/No while BGM and Favorites said On/Off for the same kind of setting.
func BoolLabel(value bool) string {
	if value {
		return "On"
	}
	return "Off"
}

// ApplyOptions applies saved interface settings to a running application.
// Every settings page repeated this same three-step block after saving.
func ApplyOptions(app *tview.Application, root tview.Primitive, options ApplicationOptions) {
	_ = SetCurrentTheme(options.Theme)
	app.EnableMouse(options.Mouse)
	app.SetRoot(WrapRoot(options, root), true)
}

// InterfaceSettingsHelp is the footer help for the shared [tui] settings, so
// the same setting reads the same way in every app.
var InterfaceSettingsHelp = map[string]string{
	"Theme":              "Choose a color palette to apply after saving.",
	"Mouse":              "Enable mouse input alongside keyboard and controller navigation.",
	"CRT mode":           "Use a compact 75-column, 15-row layout for CRT displays.",
	"On-screen keyboard": "Use controller-friendly text entry; Off needs a physical keyboard.",
}
