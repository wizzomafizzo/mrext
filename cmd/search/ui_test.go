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

//nolint:gosec // Tests only operate on temporary fixture paths.
package main

import (
	"strings"
	"testing"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	"github.com/wizzomafizzo/mrext/pkg/config"
	"github.com/wizzomafizzo/mrext/pkg/gamesdb"
	"github.com/wizzomafizzo/mrext/pkg/tui"
)

func newTestUI(t *testing.T, launch bool) *ui {
	t.Helper()
	view := newUI(&config.UserConfig{}, launch)
	view.options = tui.ApplicationOptions{Theme: "default"}
	app, err := tui.NewApplication(view.options)
	if err != nil {
		t.Fatal(err)
	}
	view.app, view.pages = app, tview.NewPages()
	return view
}

func drawn(t *testing.T, view *ui) string {
	t.Helper()
	screen := tcell.NewSimulationScreen("UTF-8")
	if err := screen.Init(); err != nil {
		t.Fatal(err)
	}
	defer screen.Fini()
	screen.SetSize(75, 15)
	view.pages.SetRect(0, 0, 75, 15)
	view.pages.Draw(screen)
	var text strings.Builder
	for y := range 15 {
		for x := range 75 {
			value, _, _ := screen.Get(x, y)
			_, _ = text.WriteString(value)
		}
		_ = text.WriteByte('\n')
	}
	return text.String()
}

func TestResultsKeepTheSystemPrefixAndDropDuplicates(t *testing.T) {
	view := newTestUI(t, true)
	view.setResults([]gamesdb.SearchResult{
		{SystemID: "NES", Name: "Super Mario Bros", Path: "/a/mario.nes"},
		{SystemID: "NES", Name: "Super Mario Bros", Path: "/b/mario.nes"},
		{SystemID: "SNES", Name: "Super Mario World", Path: "/a/world.sfc"},
	})

	// The old picker formatted rows this way and dropped repeats of the same
	// display name; both are what a user recognises in the list.
	want := []string{"[NES] Super Mario Bros", "[SNES] Super Mario World"}
	if len(view.names) != len(want) {
		t.Fatalf("names = %v, want %v", view.names, want)
	}
	for i, name := range want {
		if view.names[i] != name {
			t.Errorf("names[%d] = %q, want %q", i, view.names[i], name)
		}
	}
	if len(view.results) != len(want) {
		t.Fatalf("results = %d, want %d", len(view.results), len(want))
	}
}

func TestMainPageKeepsTheFamiliarButtons(t *testing.T) {
	view := newTestUI(t, true)
	view.showMain()

	// Same actions as the standalone windows offered, plus Search in place of
	// the keyboard having been the screen itself.
	screen := drawn(t, view)
	for _, label := range []string{"Launch", "Search", "PgUp", "PgDn", "Options", "Exit"} {
		if !strings.Contains(screen, label) {
			t.Errorf("button %q missing:\n%s", label, screen)
		}
	}

	// -print picks a game instead of launching it, and says so.
	picker := newTestUI(t, false)
	picker.showMain()
	if !strings.Contains(drawn(t, picker), "Select") {
		t.Error("print mode should offer Select rather than Launch")
	}
}

func TestStatusLineReportsQueryAndPosition(t *testing.T) {
	view := newTestUI(t, true)
	view.showMain()
	if got := view.status(); !strings.Contains(got, "Press Search") {
		t.Errorf("empty state = %q", got)
	}

	view.query = "mario"
	view.setResults([]gamesdb.SearchResult{
		{SystemID: "NES", Name: "Super Mario Bros", Path: "/a/mario.nes"},
		{SystemID: "SNES", Name: "Super Mario World", Path: "/a/world.sfc"},
	})
	view.showMain()
	if got := view.status(); !strings.Contains(got, "1/2") {
		t.Errorf("position = %q, want a 1/2 counter", got)
	}

	view.page(1)
	if got := view.status(); !strings.Contains(got, "2/2") {
		t.Errorf("after paging = %q", got)
	}
	// Paging past the end clamps rather than wrapping.
	view.page(resultsPageSize)
	if got := view.status(); !strings.Contains(got, "2/2") {
		t.Errorf("clamped position = %q", got)
	}

	view.query = "nothing"
	view.setResults(nil)
	view.showMain()
	if got := view.status(); !strings.Contains(got, "No results") {
		t.Errorf("empty results = %q", got)
	}
}

func TestOptionsPageOffersTheDatabaseUpdate(t *testing.T) {
	view := newTestUI(t, true)
	view.showMain()
	view.showOptions()

	screen := drawn(t, view)
	if !strings.Contains(screen, "Update games database") {
		t.Errorf("options page missing its action:\n%s", screen)
	}
	// The footer says what it costs, which the old one-line picker did not.
	if !strings.Contains(screen, "several minutes") {
		t.Errorf("options page does not warn about the cost:\n%s", screen)
	}
}
