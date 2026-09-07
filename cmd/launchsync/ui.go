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
	"strings"
	"sync/atomic"

	"github.com/rivo/tview"
	"github.com/wizzomafizzo/mrext/pkg/config"
	"github.com/wizzomafizzo/mrext/pkg/tui"
	"github.com/wizzomafizzo/mrext/pkg/version"
)

const (
	appTitle    = "LaunchSync"
	pageSummary = "launchsync_summary"
)

// showSyncScreen runs the sync behind a progress modal and then shows what it
// did. The console log this replaces printed "Building games index... " with
// no newline and then blocked through the whole scan, so a slow step looked
// like a hang.
//
// It reports whether a screen was ever obtained, so a headless run can fall
// back to the console instead of exiting having done nothing.
func showSyncScreen(cfg *config.UserConfig) (started bool, err error) {
	options := tui.ApplicationOptions{
		Theme:   cfg.TUI.Theme,
		Mouse:   cfg.TUI.Mouse,
		CRTMode: cfg.TUI.CRTMode,
	}
	var drew *atomic.Bool
	builder := func() (*tview.Application, error) {
		app, buildErr := tui.NewApplication(options)
		if buildErr != nil {
			return nil, fmt.Errorf("create TUI application: %w", buildErr)
		}
		pages := tview.NewPages()
		app.SetRoot(tui.WrapRoot(options, pages), true)

		var outcomes []syncOutcome
		var lines []string
		progress := tui.ProgressOptions{
			Initial: tui.ProgressUpdate{Text: "Searching for sync files..."},
		}
		// Deferred until this application is drawing. Starting it here would
		// run a second sync when BuildAndRetry rebuilds for /dev/tty2, with
		// both writing the same shortcut files.
		drew = tui.RunWhenStarted(app, func() {
			tui.ShowProgressModal(pages, app, progress, func(update func(tui.ProgressUpdate)) error {
				result, syncErr := runSync(cfg, func(line string) {
					lines = append(lines, line)
					update(tui.ProgressUpdate{Text: line})
				})
				outcomes = result
				if syncErr != nil {
					return syncErr
				}
				return nil
			}, func(syncErr error) {
				summary := func() { showSummary(pages, app, outcomes, lines) }
				if syncErr != nil {
					// Go to the summary afterwards rather than stopping. A sync
					// that fails partway has already linked games, and Details is
					// the only place the log exists when the console path is not
					// the one being used.
					tui.ShowErrorModal(pages, app, syncErr.Error(), summary)
					return
				}
				summary()
			})
		})
		return app, nil
	}
	runErr := tui.BuildAndRetry(builder)
	started = drew != nil && drew.Load()
	if runErr != nil {
		return started, fmt.Errorf("run LaunchSync TUI: %w", runErr)
	}
	return started, nil
}

// showSummary lists what each sync file produced. The full log stays available
// behind Details, because "not found" for a specific game is the thing people
// actually need to read.
func showSummary(pages *tview.Pages, app *tview.Application, outcomes []syncOutcome, lines []string) {
	list := tui.NewMenuList()
	list.AddHeader("Sync files")
	if len(outcomes) == 0 {
		list.AddRow("Nothing to sync", tui.MenuRowItem)
	}
	for _, outcome := range outcomes {
		list.AddRow(fmt.Sprintf("%s  %d found, %d missing, %d failed",
			outcome.name, outcome.found, outcome.missing, outcome.failed), tui.MenuRowItem)
	}

	bar := tui.NewButtonBar(app).
		AddButtonWithHelp("Details", "Show the full sync log", func() {
			tui.ShowInfoModal(pages, app, appTitle+" log", strings.Join(lines, "\n"),
				func() { showSummary(pages, app, outcomes, lines) })
		}).
		AddButtonWithHelp("Exit", "Exit LaunchSync", app.Stop).
		SetupNavigation(app.Stop)
	frame := tui.NewPageFrame(app).
		SetTitle(appTitle).
		SetContent(list).
		SetHelpText(summaryHelp(outcomes)).
		SetButtonBar(bar).
		SetOnEscape(app.Stop)
	frame.SetVersion(version.Short())
	bar.SetHelpCallback(func(text string) { frame.SetHelpText(text) })
	frame.SetupContentToButtonNavigation()
	pages.AddAndSwitchToPage(pageSummary, frame, true)
	app.SetFocus(list)
}

func summaryHelp(outcomes []syncOutcome) string {
	found, missing, failed := 0, 0, 0
	for _, outcome := range outcomes {
		found += outcome.found
		missing += outcome.missing
		failed += outcome.failed
	}
	return fmt.Sprintf("%d shortcuts created, %d games not found, %d errors.", found, missing, failed)
}
