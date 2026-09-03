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
		return err
	}
	if !addService {
		return nil
	}
	if err := startup.AddService("mrext/" + appName); err != nil {
		return err
	}
	return startup.Save()
}

func tryNonInteractiveAddToStartup(print bool) {
	var startup mister.Startup

	err := startup.Load()
	if err != nil {
		logger.Error("failed to load startup file: %s", err)
		if print {
			fmt.Printf("Failed to load startup file: %s\n", err)
		}
		return
	}

	if !startup.Exists("mrext/" + appName) {
		err = startup.AddService("mrext/" + appName)
		if err != nil {
			logger.Error("failed to add to startup: %s", err)
			if print {
				fmt.Printf("Failed to add to startup: %s\n", err)
			}
			return
		}

		err = startup.Save()
		if err != nil {
			logger.Error("failed to save startup: %s", err)
			if print {
				fmt.Printf("Failed to save startup: %s\n", err)
			}
			return
		}

		if print {
			fmt.Println("Added Remote to MiSTer startup.")
		}
	}
}

const (
	displayNothing = iota
	displayUninstall
)

func displayServiceInfo(svc *service.Service, cfg *config.UserConfig) (int, error) {
	ip, err := utils.GetLocalIp()
	appURL := fmt.Sprintf("http://<MiSTer IP>:%d", appPort)
	if err != nil {
		logger.Error("could not get local ip: %s", err)
	} else {
		appURL = fmt.Sprintf("http://%s:%d", ip, appPort)
	}
	altURL := ""
	if cfg.Remote.MdnsService {
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
				message = fmt.Sprintf("%s\n\nAccess Remote with this URL:\n%s\n%s\n\nIt's safe to exit; service will continue running.", state, appURL, altURL)
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
		return displayNothing, err
	}
	return action, nil
}

func displayNonInteractiveServiceInfo(service *service.Service) {
	ip, err := utils.GetLocalIp()
	appUrl := ""
	if err != nil {
		logger.Error("could not get local ip: %s", err)
		appUrl = fmt.Sprintf("http://<MiSTer IP>:%d", appPort)
	} else {
		appUrl = fmt.Sprintf("http://%s:%d", ip, appPort)
	}

	var statusText string
	running := service.Running()
	if running {
		statusText = "Service is RUNNING."
	} else {
		statusText = "Service is NOT RUNNING."
	}

	fmt.Println(statusText)
	fmt.Println("Access Remote with this URL:")
	fmt.Println(appUrl)
	fmt.Println("It's safe to exit, the service will continue running.")
}

func removeFromStartup() error {
	startup := mister.Startup{}

	err := startup.Load()
	if err != nil {
		logger.Error("failed to load startup: %s", err)
		return err
	}

	startupName := "mrext/" + appName

	if startup.Exists(startupName) {
		err := startup.Remove(startupName)
		if err != nil {
			logger.Error("failed to remove startup: %s", err)
			return err
		}

		err = startup.Save()
		if err != nil {
			logger.Error("failed to save startup: %s", err)
			return err
		}
	}

	return nil
}

func uninstallService(svc *service.Service) {
	fmt.Println("Uninstalling MiSTer Remote...")

	if svc.Running() {
		err := svc.Stop()
		if err != nil {
			logger.Error("failed to stop service: %s", err)
		} else {
			fmt.Println("Stopped service.")
		}
	}

	err := removeFromStartup()
	if err != nil {
		logger.Error("failed to remove from startup: %s", err)
		fmt.Println("Error removing from startup:", err)
		os.Exit(1)
	} else {
		fmt.Println("Removed from MiSTer startup.")
	}

	searchDbPath := filepath.Join(config.SdFolder, "search.db")
	if _, err := os.Stat(searchDbPath); err == nil {
		err = os.Remove(searchDbPath)
		if err != nil {
			logger.Error("failed to remove search db file: %s", err)
			fmt.Println("Error removing search db file:", err)
			os.Exit(1)
		} else {
			fmt.Println("Removed search.db file.")
		}
	}

	menuJpgPath := filepath.Join(config.SdFolder, "menu.jpg")
	menuJpg, err := os.Lstat(menuJpgPath)
	if err == nil && menuJpg.Mode()&os.ModeSymlink != 0 {
		err = os.Remove(menuJpgPath)
		if err != nil {
			logger.Error("failed to remove menu.jpg symlink: %s", err)
			fmt.Println("Error removing menu.jpg symlink:", err)
			os.Exit(1)
		} else {
			fmt.Println("Removed menu.jpg symlink.")
		}
	}

	menuPngPath := filepath.Join(config.SdFolder, "menu.png")
	menuPng, err := os.Lstat(menuPngPath)
	if err == nil && menuPng.Mode()&os.ModeSymlink != 0 {
		err = os.Remove(menuPngPath)
		if err != nil {
			logger.Error("failed to remove menu.png symlink: %s", err)
			fmt.Println("Error removing menu.png symlink:", err)
			os.Exit(1)
		} else {
			fmt.Println("Removed menu.png symlink.")
		}
	}

	fmt.Println("Uninstall complete.")
}
