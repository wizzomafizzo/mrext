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

	"github.com/rivo/tview"
	"github.com/wizzomafizzo/mrext/pkg/config"
	"github.com/wizzomafizzo/mrext/pkg/mister"
	"github.com/wizzomafizzo/mrext/pkg/service"
	"github.com/wizzomafizzo/mrext/pkg/tui"
	"github.com/wizzomafizzo/mrext/pkg/version"
)

const (
	appTitle    = "PlayLog"
	pageService = "playlog_service"
	pageStats   = "playlog_stats"
	topCount    = 10
)

// showServiceScreen replaces the text report that used to scroll past on the
// Scripts console. The same numbers are here, on a page that stays put.
func showServiceScreen(svc *service.Service, cfg *config.UserConfig) error {
	options := tui.ApplicationOptions{
		Theme:            cfg.TUI.Theme,
		Mouse:            cfg.TUI.Mouse,
		CRTMode:          cfg.TUI.CRTMode,
		OnScreenKeyboard: cfg.TUI.OnScreenKeyboard,
	}
	builder := func() (*tview.Application, error) {
		app, err := tui.NewApplication(options)
		if err != nil {
			return nil, fmt.Errorf("create TUI application: %w", err)
		}
		pages := tview.NewPages()
		var page *tui.ServicePage
		page = tui.NewServicePage(app, pages, appTitle, tui.ServiceActions{
			Running: svc.Running,
			Start:   svc.Start,
			Stop:    svc.Stop,
			Restart: svc.Restart,
			Exit:    app.Stop,
			Uninstall: func() {
				confirmUninstall(pages, app, svc, func() { page.Redraw() })
			},
			Status: func(running bool) string {
				state := "Service is NOT RUNNING"
				if running {
					state = "Service is RUNNING"
				}
				return state + "\n\nPlay time is tracked in the background.\n" +
					"It's safe to exit; the service keeps running."
			},
		})
		page.Frame().SetVersion(version.Short())

		// Stats is the reason most people open PlayLog, so it gets its own
		// button rather than living behind a menu.
		page.ButtonBar().AddButtonWithHelp("Stats", "Show the most played cores and games",
			func() { showStats(pages, app, func() { page.Show(pageService) }) })
		page.Show(pageService)
		app.SetRoot(tui.WrapRoot(options, pages), true)

		installed, err := startupInstalled()
		if err != nil {
			tui.ShowErrorModal(pages, app, err.Error(), func() { page.Redraw() })
			return app, nil
		}
		if !installed {
			tui.ShowConfirmModal(pages, app, appTitle,
				appTitle+" must run on MiSTer startup to record play time.\n\nAdd it now?",
				func() {
					if addErr := addToStartup(); addErr != nil {
						tui.ShowErrorModal(pages, app, addErr.Error(), func() { page.Redraw() })
						return
					}
					page.Redraw()
				},
				func() { page.Redraw() })
		}
		return app, nil
	}
	if err := tui.BuildAndRetry(builder); err != nil {
		return fmt.Errorf("run PlayLog TUI: %w", err)
	}
	return nil
}

func showStats(pages *tview.Pages, app *tview.Application, back func()) {
	list := tui.NewMenuList()
	db, err := openPlayLogDb()
	if err != nil {
		tui.ShowErrorModal(pages, app, err.Error(), back)
		return
	}
	defer func() { _ = db.db.Close() }()

	cores, coresErr := db.topCores(topCount)
	games, gamesErr := db.topGames(topCount)
	if coresErr != nil || gamesErr != nil {
		problem := coresErr
		if problem == nil {
			problem = gamesErr
		}
		tui.ShowErrorModal(pages, app, problem.Error(), back)
		return
	}

	list.AddHeader("Top played cores")
	if len(cores) == 0 {
		list.AddRow("Nothing recorded yet", tui.MenuRowItem)
	}
	for _, core := range cores {
		list.AddRow(fmt.Sprintf("%s  %s", core.Name, formatPlayTime(core.Time)), tui.MenuRowItem)
	}
	list.AddHeader("Top played games")
	if len(games) == 0 {
		list.AddRow("Nothing recorded yet", tui.MenuRowItem)
	}
	for _, game := range games {
		list.AddRow(fmt.Sprintf("%s  %s", game.Name, formatPlayTime(game.Time)), tui.MenuRowItem)
	}

	bar := tui.NewButtonBar(app).
		AddButtonWithHelp("Back", "Return to the service screen", back).
		SetupNavigation(back)
	frame := tui.NewPageFrame(app).
		SetTitle(appTitle, "Stats").
		SetContent(list).
		SetHelpText("Time played per core and per game.").
		SetButtonBar(bar).
		SetOnEscape(back)
	frame.SetVersion(version.Short())
	bar.SetHelpCallback(func(text string) { frame.SetHelpText(text) })
	frame.SetupContentToButtonNavigation()
	pages.AddAndSwitchToPage(pageStats, frame, true)
	app.SetFocus(list)
}

// formatPlayTime keeps the "Xh Ym" shape the console report used.
func formatPlayTime(seconds int) string {
	return fmt.Sprintf("%dh %dm", seconds/3600, (seconds%3600)/60)
}

func confirmUninstall(pages *tview.Pages, app *tview.Application, svc *service.Service, done func()) {
	tui.ShowConfirmModal(pages, app, "Uninstall "+appTitle,
		"Stop the service and remove "+appTitle+" from MiSTer startup?\n\n"+
			"Your play history in playlog.db is kept.",
		func() { uninstallService(pages, app, svc, done) },
		done)
}

func uninstallService(pages *tview.Pages, app *tview.Application, svc *service.Service, done func()) {
	var report strings.Builder
	if svc.Running() {
		if err := svc.Stop(); err != nil {
			_, _ = fmt.Fprintf(&report, "Could not stop the service: %s\n", err)
		} else {
			_, _ = report.WriteString("Stopped the service.\n")
		}
	}
	if err := removeFromStartup(); err != nil {
		_, _ = fmt.Fprintf(&report, "Could not remove the startup entry: %s\n", err)
	} else {
		_, _ = report.WriteString("Removed the startup entry.\n")
	}
	// playlog.db is never touched here: it is the user's play history, and the
	// docs are explicit that it should outlive an uninstall.
	_, _ = fmt.Fprintf(&report, "\nYour history is still in %s.\n", config.PlayLogDbFile)
	_, _ = fmt.Fprintf(&report, "Delete %s/playlog.sh to finish.", config.ScriptsFolder)
	tui.ShowInfoModal(pages, app, "Uninstall "+appTitle, report.String(), done)
}

func startupInstalled() (bool, error) {
	var startup mister.Startup
	if err := startup.Load(); err != nil {
		return false, fmt.Errorf("load startup configuration: %w", err)
	}
	return startup.Exists("mrext/" + appName), nil
}

func addToStartup() error {
	var startup mister.Startup
	if err := startup.Load(); err != nil {
		return fmt.Errorf("load startup configuration: %w", err)
	}
	if startup.Exists("mrext/" + appName) {
		return nil
	}
	if err := startup.AddService("mrext/" + appName); err != nil {
		return fmt.Errorf("add PlayLog startup service: %w", err)
	}
	if err := startup.Save(); err != nil {
		return fmt.Errorf("save startup configuration: %w", err)
	}
	return nil
}

func removeFromStartup() error {
	var startup mister.Startup
	if err := startup.Load(); err != nil {
		return fmt.Errorf("load startup configuration: %w", err)
	}
	name := "mrext/" + appName
	if !startup.Exists(name) {
		return nil
	}
	if err := startup.Remove(name); err != nil {
		return fmt.Errorf("remove startup entry: %w", err)
	}
	if err := startup.Save(); err != nil {
		return fmt.Errorf("save startup configuration: %w", err)
	}
	return nil
}
