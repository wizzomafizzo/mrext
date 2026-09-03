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
	"time"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	"github.com/wizzomafizzo/mrext/pkg/config"
	"github.com/wizzomafizzo/mrext/pkg/mister"
	"github.com/wizzomafizzo/mrext/pkg/service"
	"github.com/wizzomafizzo/mrext/pkg/tui"
	"github.com/wizzomafizzo/mrext/pkg/utils"
)

func tryAddStartup() error {
	var startup mister.Startup
	if err := startup.Load(); err != nil {
		logger.Error("failed to load startup file: %s", err)
	}
	if startup.Exists("mrext/" + appName) {
		return nil
	}

	addService := false
	builder := func() (*tview.Application, error) {
		app := tview.NewApplication()
		modal := tview.NewModal().
			SetText("Add Remote service to MiSTer startup?\nThis won't impact MiSTer's performance.").
			AddButtons([]string{"Yes", "No"}).
			SetDoneFunc(func(_ int, label string) {
				addService = label == "Yes"
				app.Stop()
			})
		app.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
			if event.Key() == tcell.KeyEscape {
				app.Stop()
				return nil
			}
			return event
		})
		return app.SetRoot(modal, true).SetFocus(modal), nil
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

	selected := 3
	action := displayNothing
	builder := func() (*tview.Application, error) {
		app := tview.NewApplication()
		status := tview.NewTextView().SetTextAlign(tview.AlignCenter)
		footer := tview.NewTextView().SetDynamicColors(true).SetTextAlign(tview.AlignCenter)
		content := tview.NewFlex().SetDirection(tview.FlexRow).
			AddItem(status, 0, 1, false).
			AddItem(footer, 1, 0, false)
		content.SetBorder(true)

		draw := func() {
			running := svc.Running()
			state := "Service is NOT RUNNING"
			toggle := "Start"
			message := state
			if running {
				state = "Service is RUNNING"
				toggle = "Stop"
				message = fmt.Sprintf(
					"%s\n\nAccess Remote with this URL:\n%s\n%s\n\n"+
						"It's safe to exit; service will continue running.",
					state,
					appURL,
					altURL,
				)
			}
			status.SetText(message)
			footer.SetText(tui.ButtonBar([]string{toggle, "Restart", "Uninstall", "Exit"}, selected))
		}
		draw()

		app.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
			switch event.Key() {
			case tcell.KeyEscape:
				app.Stop()
				return nil
			case tcell.KeyLeft:
				selected = (selected + 3) % 4
				draw()
				return nil
			case tcell.KeyRight:
				selected = (selected + 1) % 4
				draw()
				return nil
			case tcell.KeyEnter:
				switch selected {
				case 0:
					if svc.Running() {
						err = svc.Stop()
					} else {
						err = svc.Start()
					}
					if err != nil {
						logger.Error("could not toggle service: %s", err)
					}
					time.Sleep(time.Second)
					draw()
				case 1:
					if err = svc.Restart(); err != nil {
						logger.Error("could not restart service: %s", err)
					}
					time.Sleep(time.Second)
					draw()
				case 2:
					action = displayUninstall
					app.Stop()
				case 3:
					app.Stop()
				}
				return nil
			default:
				return event
			}
		})
		return app.SetRoot(tui.Centered(57, 11, content), true), nil
	}
	if err := tui.BuildAndRetry(builder); err != nil {
		return displayNothing, fmt.Errorf("show service controls: %w", err)
	}
	return action, nil
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
