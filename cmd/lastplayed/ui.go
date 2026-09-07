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
	"os"
	"path/filepath"
	"strings"

	"github.com/rivo/tview"
	"github.com/wizzomafizzo/mrext/pkg/config"
	"github.com/wizzomafizzo/mrext/pkg/mister"
	"github.com/wizzomafizzo/mrext/pkg/service"
	"github.com/wizzomafizzo/mrext/pkg/tui"
	"github.com/wizzomafizzo/mrext/pkg/version"
)

const (
	appTitle    = "LastPlayed"
	pageService = "lastplayed_service"
)

// showServiceScreen replaces the console output this app used to print. It is
// the same screen Remote has offered for years, so the state and the actions
// are visible instead of scrolling past.
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
				confirmUninstall(pages, app, svc, cfg, func() { page.Redraw() })
			},
			Status: func(running bool) string { return serviceStatus(running, cfg) },
		})
		page.Frame().SetVersion(version.Short())
		page.Show(pageService)
		app.SetRoot(tui.WrapRoot(options, pages), true)

		// Same question the console prompt used to ask, on the shared dialog.
		installed, err := startupInstalled()
		if err != nil {
			tui.ShowErrorModal(pages, app, err.Error(), func() { page.Redraw() })
			return app, nil
		}
		if !installed {
			tui.ShowConfirmModal(pages, app, appTitle,
				appTitle+" must run on MiSTer startup to keep shortcuts up to date.\n\nAdd it now?",
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
		return fmt.Errorf("run LastPlayed TUI: %w", err)
	}
	return nil
}

func serviceStatus(running bool, cfg *config.UserConfig) string {
	state := "Service is NOT RUNNING"
	if running {
		state = "Service is RUNNING"
	}
	shortcut := cfg.LastPlayed.LastPlayedName
	if shortcut == "" {
		shortcut = defaultLastPlayedName
	}
	folder := cfg.LastPlayed.RecentFolderName
	if folder == "" {
		folder = defaultRecentFolderName
	}
	return fmt.Sprintf(
		"%s\n\nShortcut: %s\nRecent folder: _%s\n\nIt's safe to exit; the service keeps running.",
		state, shortcut, folder,
	)
}

// confirmUninstall walks the removals separately, because deleting a menu
// folder someone has curated is a different decision from stopping a service.
func confirmUninstall(
	pages *tview.Pages, app *tview.Application, svc *service.Service, cfg *config.UserConfig, done func(),
) {
	tui.ShowConfirmModal(pages, app, "Uninstall "+appTitle,
		"Stop the service and remove "+appTitle+" from MiSTer startup?",
		func() { uninstallService(pages, app, svc, cfg, done) },
		done)
}

// uninstallService keeps the blocking half off the event goroutine. Stopping a
// service waits for the process to exit and rewriting user-startup.sh is disk
// I/O, and doing either inline froze the screen for as long as it took.
// ServicePage.run already keeps Start, Stop and Restart off it, but Uninstall
// is called straight from its button and has to arrange this itself.
func uninstallService(
	pages *tview.Pages, app *tview.Application, svc *service.Service, cfg *config.UserConfig, done func(),
) {
	report := &strings.Builder{}
	options := tui.ProgressOptions{Initial: tui.ProgressUpdate{Text: "Stopping the service..."}}
	tui.ShowProgressModal(pages, app, options, func(update func(tui.ProgressUpdate)) error {
		stopService(svc, report)
		update(tui.ProgressUpdate{Text: "Removing the startup entry..."})
		clearStartupEntry(report)
		return nil
	}, func(error) {
		// Config and generated menu entries are asked about separately, and only
		// after the service is actually stopped.
		askRemoveShortcuts(pages, app, cfg, report, done)
	})
}

func stopService(svc *service.Service, report *strings.Builder) {
	if !svc.Running() {
		return
	}
	if err := svc.Stop(); err != nil {
		_, _ = fmt.Fprintf(report, "Could not stop the service: %s\n", err)
		return
	}
	_, _ = report.WriteString("Stopped the service.\n")
}

func clearStartupEntry(report *strings.Builder) {
	if err := removeFromStartup(); err != nil {
		_, _ = fmt.Fprintf(report, "Could not remove the startup entry: %s\n", err)
		return
	}
	_, _ = report.WriteString("Removed the startup entry.\n")
}

func askRemoveShortcuts(
	pages *tview.Pages, app *tview.Application, cfg *config.UserConfig, report *strings.Builder, done func(),
) {
	finish := func() {
		tui.ShowInfoModal(pages, app, "Uninstall "+appTitle,
			report.String()+"\nDelete "+config.ScriptsFolder+"/lastplayed.sh to finish.", done)
	}
	tui.ShowConfirmModal(pages, app, "Remove shortcuts",
		"Also delete the generated Last Played shortcut and Recently Played folder?",
		func() {
			options := tui.ProgressOptions{Initial: tui.ProgressUpdate{Text: "Removing shortcuts..."}}
			tui.ShowProgressModal(pages, app, options, func(func(tui.ProgressUpdate)) error {
				removeGenerated(cfg, report)
				return nil
			}, func(error) { finish() })
		},
		finish)
}

func removeGenerated(cfg *config.UserConfig, report *strings.Builder) {
	shortcut := cfg.LastPlayed.LastPlayedName
	if shortcut == "" {
		shortcut = defaultLastPlayedName
	}
	folder := cfg.LastPlayed.RecentFolderName
	if folder == "" {
		folder = defaultRecentFolderName
	}
	for _, extension := range []string{".mgl", ".mra"} {
		path := filepath.Join(config.SdFolder, shortcut+extension)
		if err := os.Remove(path); err == nil {
			_, _ = fmt.Fprintf(report, "Removed %s.\n", path)
		} else if !os.IsNotExist(err) {
			_, _ = fmt.Fprintf(report, "Could not remove %s: %s\n", path, err)
		}
	}
	recent := filepath.Join(config.SdFolder, "_"+folder)
	if err := os.RemoveAll(recent); err != nil {
		_, _ = fmt.Fprintf(report, "Could not remove %s: %s\n", recent, err)
		return
	}
	_, _ = fmt.Fprintf(report, "Removed %s.\n", recent)
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
