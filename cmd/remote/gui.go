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

	"github.com/rivo/tview"
	"github.com/wizzomafizzo/mrext/pkg/config"
	"github.com/wizzomafizzo/mrext/pkg/mister"
	"github.com/wizzomafizzo/mrext/pkg/service"
	"github.com/wizzomafizzo/mrext/pkg/tui"
	"github.com/wizzomafizzo/mrext/pkg/utils"
)

func tryAddStartup(cfg *config.UserConfig) error {
	var startup mister.Startup
	if err := startup.Load(); err != nil {
		// Continuing with a zero-value Startup would rewrite user-startup.sh
		// from scratch and drop every other entry in it.
		logger.Error("failed to load startup file: %s", err)
		return fmt.Errorf("load startup configuration: %w", err)
	}
	if startup.Exists("mrext/" + appName) {
		return nil
	}

	// Same question and the same Yes/No, on the shared confirm modal so it
	// looks like the rest of the app instead of a bare tview.Modal.
	addService := false
	options := tui.ApplicationOptions{
		Theme: cfg.TUI.Theme, Mouse: cfg.TUI.Mouse, CRTMode: cfg.TUI.CRTMode,
	}
	builder := func() (*tview.Application, error) {
		app, err := tui.NewApplication(options)
		if err != nil {
			return nil, fmt.Errorf("create TUI application: %w", err)
		}
		pages := tview.NewPages()
		done := func(yes bool) {
			addService = yes
			app.Stop()
		}
		tui.ShowConfirmModal(pages, app, appTitle,
			"Add Remote service to MiSTer startup?\nThis won't impact MiSTer's performance.",
			func() { done(true) }, func() { done(false) })
		app.SetRoot(tui.WrapRoot(options, pages), true)
		return app, nil
	}
	if err := tui.BuildAndRetry(builder); err != nil {
		return fmt.Errorf("show startup prompt: %w", err)
	}
	if !addService {
		return nil
	}
	if err := startup.AddService("mrext/" + appName); err != nil {
		return fmt.Errorf("add Remote startup service: %w", err)
	}
	if err := startup.Save(); err != nil {
		return fmt.Errorf("save startup configuration: %w", err)
	}
	return nil
}

func tryNonInteractiveAddToStartup(printOutput bool) {
	var startup mister.Startup

	err := startup.Load()
	if err != nil {
		logger.Error("failed to load startup file: %s", err)
		if printOutput {
			_, _ = fmt.Printf("Failed to load startup file: %s\n", err)
		}
		return
	}

	if !startup.Exists("mrext/" + appName) {
		err = startup.AddService("mrext/" + appName)
		if err != nil {
			logger.Error("failed to add to startup: %s", err)
			if printOutput {
				_, _ = fmt.Printf("Failed to add to startup: %s\n", err)
			}
			return
		}

		err = startup.Save()
		if err != nil {
			logger.Error("failed to save startup: %s", err)
			if printOutput {
				_, _ = fmt.Printf("Failed to save startup: %s\n", err)
			}
			return
		}

		if printOutput {
			_, _ = fmt.Println("Added Remote to MiSTer startup.")
		}
	}
}

const (
	displayNothing = iota
	displayUninstall
)

// appTitle names the Scripts-menu screen.
const appTitle = "Remote"

func displayServiceInfo(svc *service.Service, cfg *config.UserConfig) (int, error) {
	ip, err := utils.GetLocalIP()
	appURL := fmt.Sprintf("http://<MiSTer IP>:%d", appPort)
	if err != nil {
		logger.Error("could not get local ip: %s", err)
	} else {
		appURL = fmt.Sprintf("http://%s:%d", ip, appPort)
	}
	altURL := ""
	if cfg.Remote.MDNSService {
		hostname, _ := os.Hostname()
		altURL = fmt.Sprintf("OR http://%s.local:%d", hostname, appPort)
	}

	action := displayNothing
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
		status := tview.NewTextView().SetTextAlign(tview.AlignCenter)

		var bar *tui.ButtonBar
		draw := func() {
			running := svc.Running()
			message := "Service is NOT RUNNING"
			toggle := "Start"
			if running {
				toggle = "Stop"
				message = fmt.Sprintf(
					"Service is RUNNING\n\nAccess Remote with this URL:\n%s\n%s\n\n"+
						"It's safe to exit; service will continue running.",
					appURL,
					altURL,
				)
			}
			status.SetText(message)
			bar.UpdateButtonLabel(0, toggle)
		}

		exit := func() { app.Stop() }
		bar = tui.NewButtonBar(app).
			AddButtonWithHelp("Start", "Start or stop the Remote service", func() {
				if svc.Running() {
					runServiceAction(app, status, "Stopping service...", svc.Stop, draw)
					return
				}
				runServiceAction(app, status, "Starting service...", svc.Start, draw)
			}).
			AddButtonWithHelp("Restart", "Stop and start the service again", func() {
				runServiceAction(app, status, "Restarting service...", svc.Restart, draw)
			}).
			AddButtonWithHelp("Uninstall", "Remove Remote's startup entry and generated files", func() {
				action = displayUninstall
				app.Stop()
			}).
			AddButtonWithHelp("Exit", "Leave this screen; the service keeps running", exit).
			SetupNavigation(exit)
		// Exit has been the default since this screen existed: people open it
		// to read the URL and leave.
		bar.SetFocusedIndex(3)

		frame := tui.NewPageFrame(app).
			SetTitle(appTitle).
			SetContent(status).
			SetFocusTarget(bar).
			SetHelpText("Leave this screen; the service keeps running").
			SetButtonBar(bar).
			SetOnEscape(exit)
		bar.SetHelpCallback(func(text string) { frame.SetHelpText(text) })
		draw()

		pages.AddAndSwitchToPage("remote_service", frame, true)
		app.SetRoot(tui.WrapRoot(options, pages), true)
		app.SetFocus(bar)
		return app, nil
	}
	if err := tui.BuildAndRetry(builder); err != nil {
		return displayNothing, fmt.Errorf("show service controls: %w", err)
	}
	return action, nil
}

// runServiceAction runs a service command off the UI goroutine.
//
// The previous version called Stop, Start or Restart inline and then slept a
// second on the UI goroutine before redrawing, so the screen was frozen and
// unclickable with nothing to say why. A second is not long enough for a
// restart either. Say what is happening, then redraw when it is really done.
func runServiceAction(
	app *tview.Application, status *tview.TextView, message string, action func() error, redraw func(),
) {
	status.SetText(message)
	go func() {
		err := action()
		app.QueueUpdateDraw(func() {
			if err != nil {
				logger.Error("service action failed: %s", err)
			}
			redraw()
		})
	}()
}

func displayNonInteractiveServiceInfo(svc *service.Service) {
	ip, err := utils.GetLocalIP()
	appURL := ""
	if err != nil {
		logger.Error("could not get local ip: %s", err)
		appURL = fmt.Sprintf("http://<MiSTer IP>:%d", appPort)
	} else {
		appURL = fmt.Sprintf("http://%s:%d", ip, appPort)
	}

	var statusText string
	running := svc.Running()
	if running {
		statusText = "Service is RUNNING."
	} else {
		statusText = "Service is NOT RUNNING."
	}

	_, _ = fmt.Println(statusText)
	_, _ = fmt.Println("Access Remote with this URL:")
	_, _ = fmt.Println(appURL)
	_, _ = fmt.Println("It's safe to exit, the service will continue running.")
}

func removeFromStartup() error {
	startup := mister.Startup{}

	err := startup.Load()
	if err != nil {
		logger.Error("failed to load startup: %s", err)
		return fmt.Errorf("load startup configuration: %w", err)
	}

	startupName := "mrext/" + appName

	if startup.Exists(startupName) {
		err := startup.Remove(startupName)
		if err != nil {
			logger.Error("failed to remove startup: %s", err)
			return fmt.Errorf("remove Remote startup service: %w", err)
		}

		err = startup.Save()
		if err != nil {
			logger.Error("failed to save startup: %s", err)
			return fmt.Errorf("save startup configuration: %w", err)
		}
	}

	return nil
}

func uninstallService(svc *service.Service) {
	_, _ = fmt.Println("Uninstalling MiSTer Remote...")

	if svc.Running() {
		err := svc.Stop()
		if err != nil {
			logger.Error("failed to stop service: %s", err)
		} else {
			_, _ = fmt.Println("Stopped service.")
		}
	}

	err := removeFromStartup()
	if err != nil {
		logger.Error("failed to remove from startup: %s", err)
		_, _ = fmt.Println("Error removing from startup:", err)
		os.Exit(1)
	}
	_, _ = fmt.Println("Removed from MiSTer startup.")

	searchDbPath := filepath.Join(config.SdFolder, "search.db")
	if _, statErr := os.Stat(searchDbPath); statErr == nil {
		err = os.Remove(searchDbPath)
		if err != nil {
			logger.Error("failed to remove search db file: %s", err)
			_, _ = fmt.Println("Error removing search db file:", err)
			os.Exit(1)
		}
		_, _ = fmt.Println("Removed search.db file.")
	}

	menuJpgPath := filepath.Join(config.SdFolder, "menu.jpg")
	menuJpg, err := os.Lstat(menuJpgPath)
	if err == nil && menuJpg.Mode()&os.ModeSymlink != 0 {
		err = os.Remove(menuJpgPath)
		if err != nil {
			logger.Error("failed to remove menu.jpg symlink: %s", err)
			_, _ = fmt.Println("Error removing menu.jpg symlink:", err)
			os.Exit(1)
		}
		_, _ = fmt.Println("Removed menu.jpg symlink.")
	}

	menuPngPath := filepath.Join(config.SdFolder, "menu.png")
	menuPng, err := os.Lstat(menuPngPath)
	if err == nil && menuPng.Mode()&os.ModeSymlink != 0 {
		err = os.Remove(menuPngPath)
		if err != nil {
			logger.Error("failed to remove menu.png symlink: %s", err)
			_, _ = fmt.Println("Error removing menu.png symlink:", err)
			os.Exit(1)
		}
		_, _ = fmt.Println("Removed menu.png symlink.")
	}

	_, _ = fmt.Println("Uninstall complete.")
}
