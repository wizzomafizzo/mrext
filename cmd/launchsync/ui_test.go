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
	"strings"
	"testing"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	"github.com/wizzomafizzo/mrext/pkg/tui"
)

func TestSummaryCountsAndLog(t *testing.T) {
	if _, err := tui.NewApplication(tui.ApplicationOptions{Theme: "default"}); err != nil {
		t.Fatal(err)
	}
	app := tview.NewApplication()
	pages := tview.NewPages()

	outcomes := []syncOutcome{
		{name: "Best of NES", found: 12, missing: 3, failed: 1},
		{name: "SNES RPGs", found: 5},
	}
	lines := []string{"- Chrono Trigger... found /media/fat/games/SNES/ct.sfc"}
	showSummary(pages, app, outcomes, lines)

	screen := tcell.NewSimulationScreen("UTF-8")
	if err := screen.Init(); err != nil {
		t.Fatal(err)
	}
	defer screen.Fini()
	screen.SetSize(75, 15)
	pages.SetRect(0, 0, 75, 15)
	pages.Draw(screen)

	var text strings.Builder
	for y := range 15 {
		for x := range 75 {
			value, _, _ := screen.Get(x, y)
			_, _ = text.WriteString(value)
		}
		_ = text.WriteByte('\n')
	}
	drawn := text.String()

	for _, want := range []string{"Best of NES", "12 found", "3 missing", "SNES RPGs", "Details", "Exit"} {
		if !strings.Contains(drawn, want) {
			t.Errorf("summary missing %q:\n%s", want, drawn)
		}
	}
	// The footer totals across every sync file, which is what the console log
	// never told you without reading all of it.
	if got := summaryHelp(outcomes); !strings.Contains(got, "17 shortcuts created") ||
		!strings.Contains(got, "3 games not found") || !strings.Contains(got, "1 errors") {
		t.Errorf("summary help = %q", got)
	}
}
