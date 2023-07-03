package main

import (
	"fmt"
	"github.com/rthornton128/goncurses"
	"github.com/wizzomafizzo/mrext/pkg/config"
	"github.com/wizzomafizzo/mrext/pkg/curses"
	"github.com/wizzomafizzo/mrext/pkg/mister"
	"github.com/wizzomafizzo/mrext/pkg/service"
	"github.com/wizzomafizzo/mrext/pkg/utils"
	"os"
	"path/filepath"
	"strings"
	"time"
)

func tryAddStartup(stdscr *goncurses.Window) error {
	var startup mister.Startup

	err := startup.Load()
	if err != nil {
		logger.Error("failed to load startup file: %s", err)
	}

	if !startup.Exists("mrext/" + appName) {
		win, err := curses.NewWindow(stdscr, 6, 43, "", -1)
		if err != nil {
			return err
		}
		defer func(win *goncurses.Window) {
			err := win.Delete()
			if err != nil {
				logger.Error("failed to delete window: %s", err)
			}
		}(win)

		var ch goncurses.Key
		selected := 0

		for {
			win.MovePrint(1, 3, "Add Remote service to MiSTer startup?")
			win.MovePrint(2, 2, "This won't impact MiSTer's performance.")
			curses.DrawActionButtons(win, []string{"Yes", "No"}, selected, 10)

			win.NoutRefresh()
			err := goncurses.Update()
			if err != nil {
				return err
			}

			ch = win.GetChar()

			if ch == goncurses.KEY_LEFT {
				if selected == 0 {
					selected = 1
				} else if selected == 1 {
					selected = 0
				}
			} else if ch == goncurses.KEY_RIGHT {
				if selected == 0 {
					selected = 1
				} else if selected == 1 {
					selected = 0
				}
			} else if ch == goncurses.KEY_ENTER || ch == 10 || ch == 13 {
				break
			} else if ch == goncurses.KEY_ESC {
				selected = 1
				break
			}
		}

		if selected == 0 {
			err = startup.AddService("mrext/" + appName)
			if err != nil {
				return err
			}

			err = startup.Save()
			if err != nil {
				return err
			}
		}
	}

	return nil
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

func displayServiceInfo(stdscr *goncurses.Window, service *service.Service, cfg *config.UserConfig) (int, error) {
	width := 57
	height := 11

	win, err := curses.NewWindow(stdscr, height, width, "", -1)
	if err != nil {
		return displayNothing, err
	}
	defer func(win *goncurses.Window) {
		err := win.Delete()
		if err != nil {
			logger.Error("failed to delete window: %s", err)
		}
	}(win)

	printCenter := func(y int, text string) {
		x := (width - len(text)) / 2
		win.MovePrint(y, x, text)
	}

	clearLine := func(y int) {
		win.MovePrint(y, 2, strings.Repeat(" ", width-4))
	}

	ip, err := utils.GetLocalIp()
	appUrl := ""
	if err != nil {
		logger.Error("could not get local ip: %s", err)
		appUrl = fmt.Sprintf("http://<MiSTer IP>:%d", appPort)
	} else {
		appUrl = fmt.Sprintf("http://%s:%d", ip, appPort)
	}

	altUrl := ""
	if cfg.Remote.MdnsService {
		hostname, _ := os.Hostname()
		altUrl = "OR " + fmt.Sprintf("http://%s.local:%d", hostname, appPort)
	}

	var ch goncurses.Key
	selected := 3

	for {
		var statusText string
		var toggleText string
		running := service.Running()
		if running {
			statusText = "Service is RUNNING"
			toggleText = "Stop"
		} else {
			statusText = "Service is NOT RUNNING"
			toggleText = "Start"
		}

		clearLine(1)
		printCenter(1, statusText)
		clearLine(3)
		clearLine(4)
		clearLine(6)
		if running {
			printCenter(3, "Access Remote with this URL:")
			printCenter(4, appUrl)
			printCenter(5, altUrl)
			printCenter(7, "It's safe to exit, the service will continue running.")
		}

		clearLine(8)
		curses.DrawActionButtons(win, []string{toggleText, "Restart", "Uninstall", "Exit"}, selected, 1)

		win.NoutRefresh()
		err := goncurses.Update()
		if err != nil {
			return displayNothing, err
		}

		ch = win.GetChar()

		if ch == goncurses.KEY_LEFT {
			if selected == 0 {
				selected = 3
			} else {
				selected--
			}
		} else if ch == goncurses.KEY_RIGHT {
			if selected == 3 {
				selected = 0
			} else {
				selected++
			}
		} else if ch == goncurses.KEY_ENTER || ch == 10 || ch == 13 {
			if selected == 0 {
				if service.Running() {
					err := service.Stop()
					if err != nil {
						logger.Error("could not stop service: %s", err)
					}
				} else {
					err := service.Start()
					if err != nil {
						logger.Error("could not start service: %s", err)
					}
				}
				time.Sleep(1 * time.Second)
			} else if selected == 1 {
				err := service.Restart()
				if err != nil {
					logger.Error("could not restart service: %s", err)
				}
				time.Sleep(1 * time.Second)
			} else if selected == 2 {
				return displayUninstall, nil
			} else {
				break
			}
		} else if ch == goncurses.KEY_ESC {
			break
		}
	}

	return displayNothing, nil
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
