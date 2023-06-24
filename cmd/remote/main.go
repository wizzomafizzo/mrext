package main

import (
	"embed"
	"errors"
	"flag"
	"fmt"
	"github.com/wizzomafizzo/mrext/cmd/remote/control"
	"github.com/wizzomafizzo/mrext/cmd/remote/games"
	"github.com/wizzomafizzo/mrext/cmd/remote/menu"
	"github.com/wizzomafizzo/mrext/cmd/remote/music"
	"github.com/wizzomafizzo/mrext/cmd/remote/screenshots"
	"github.com/wizzomafizzo/mrext/cmd/remote/settings"
	"github.com/wizzomafizzo/mrext/cmd/remote/systems"
	"github.com/wizzomafizzo/mrext/cmd/remote/wallpapers"
	"github.com/wizzomafizzo/mrext/cmd/remote/websocket"
	"github.com/wizzomafizzo/mrext/pkg/input"
	"github.com/wizzomafizzo/mrext/pkg/tracker"
	"io/fs"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	gc "github.com/rthornton128/goncurses"

	"github.com/wizzomafizzo/mrext/pkg/config"
	"github.com/wizzomafizzo/mrext/pkg/curses"
	"github.com/wizzomafizzo/mrext/pkg/mister"
	"github.com/wizzomafizzo/mrext/pkg/service"
	"github.com/wizzomafizzo/mrext/pkg/utils"

	"github.com/gorilla/mux"
	"github.com/rs/cors"
)

const appName = "remote"
const appPort = 8182

var logger = service.NewLogger(appName)

//go:embed _client
var client embed.FS

func wsConnectPayload(trk *tracker.Tracker) func() []string {
	return func() []string {
		response := []string{
			games.GetIndexingStatus(),
		}

		if trk != nil {
			if trk.ActiveCore != "" {
				response = append(response, "coreStart:"+trk.ActiveCore)
			}

			if trk.ActiveGame != "" {
				response = append(response, "gameStart:"+trk.ActiveGame)
			}
		}

		return response
	}
}

func wsMsgHandler(kbd input.Keyboard) func(string) string {
	return func(msg string) string {
		parts := strings.SplitN(msg, ":", 2)
		cmd := parts[0]
		args := ""
		if len(parts) > 1 {
			args = parts[1]
		}

		switch cmd {
		case "getIndexStatus":
			return games.GetIndexingStatus()
		case "kbd":
			err := control.SendKeyboard(kbd, args)
			if err != nil {
				return "invalid"
			}
			return ""
		case "kbdRaw":
			code, err := strconv.Atoi(args)
			if err != nil {
				return "invalid"
			}

			err = control.SendRawKeyboard(kbd, code)
			if err != nil {
				return "invalid"
			}

			return ""
		default:
			return "invalid"
		}
	}
}

func startService(logger *service.Logger, cfg *config.UserConfig) (func() error, error) {
	kbd, err := input.NewKeyboard()
	if err != nil {
		logger.Error("failed to initialize keyboard: %s", err)
		return nil, err
	}

	trk, stopTracker, err := games.StartTracker(logger, cfg)
	if err != nil {
		logger.Error("failed to start tracker: %s", err)
		return nil, err
	}

	router := mux.NewRouter()
	setupApi(router.PathPrefix("/api").Subrouter(), kbd, trk, logger)
	router.PathPrefix("/").Handler(http.HandlerFunc(appHandler))

	corsHandler := cors.New(cors.Options{
		AllowedOrigins: []string{"*"},
		AllowedMethods: []string{"GET", "POST", "DELETE", "PUT"},
	})

	srv := &http.Server{
		Handler: corsHandler.Handler(router),
		Addr:    ":" + fmt.Sprint(appPort),
		// TODO: this will not work for large file uploads
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}

	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error("critical server error: %s", err)
			os.Exit(1)
		}
	}()

	return func() error {
		kbd.Close()

		err := stopTracker()
		if err != nil {
			return err
		}

		err = srv.Close()
		if err != nil {
			return err
		}

		return nil
	}, nil
}

func setupApi(sub *mux.Router, kbd input.Keyboard, trk *tracker.Tracker, logger *service.Logger) {
	sub.HandleFunc("/ws", websocket.Handle(logger, wsConnectPayload(trk), wsMsgHandler(kbd)))

	sub.HandleFunc("/screenshots", screenshots.AllScreenshots(logger)).Methods("GET")
	sub.HandleFunc("/screenshots", screenshots.TakeScreenshot(logger)).Methods("POST")
	sub.HandleFunc("/screenshots/{core}/{image}", screenshots.ViewScreenshot(logger)).Methods("GET")
	sub.HandleFunc("/screenshots/{core}/{image}", screenshots.DeleteScreenshot(logger)).Methods("DELETE")

	sub.HandleFunc("/systems", systems.ListSystems(logger)).Methods("GET")
	sub.HandleFunc("/systems/{id}", systems.LaunchCore(logger)).Methods("POST")

	sub.HandleFunc("/wallpapers", wallpapers.AllWallpapersHandler(logger)).Methods("GET")
	sub.HandleFunc("/wallpapers", wallpapers.UnsetWallpaperHandler(logger)).Methods("DELETE")
	sub.HandleFunc("/wallpapers/{filename:.*}", wallpapers.ViewWallpaperHandler(logger)).Methods("GET")
	sub.HandleFunc("/wallpapers/{filename:.*}", wallpapers.SetWallpaperHandler(logger)).Methods("POST")

	sub.HandleFunc("/music/status", music.Status(logger)).Methods("GET")
	sub.HandleFunc("/music/play", music.Play(logger)).Methods("POST")
	sub.HandleFunc("/music/stop", music.Stop(logger)).Methods("POST")
	sub.HandleFunc("/music/next", music.Skip(logger)).Methods("POST")
	sub.HandleFunc("/music/playback/{playback}", music.SetPlayback(logger)).Methods("POST")
	sub.HandleFunc("/music/playlist", music.AllPlaylists(logger)).Methods("GET")
	sub.HandleFunc("/music/playlist/{playlist}", music.SetPlaylist(logger)).Methods("POST")

	sub.HandleFunc("/games/search", games.Search(logger)).Methods("POST")
	sub.HandleFunc("/games/search/systems", games.ListSystems(logger)).Methods("GET")
	sub.HandleFunc("/games/launch", games.LaunchGame(logger)).Methods("POST")
	sub.HandleFunc("/games/index", games.GenerateSearchIndex).Methods("POST")
	sub.HandleFunc("/games/playing", games.HandlePlaying(trk)).Methods("GET")

	sub.HandleFunc("/launch", games.LaunchFile(logger)).Methods("POST")
	sub.HandleFunc("/launch/menu", games.LaunchMenu).Methods("POST")
	sub.HandleFunc("/launch/new", games.CreateLauncher(logger)).Methods("POST")

	sub.HandleFunc("/controls/keyboard/{key}", control.HandleKeyboard(kbd)).Methods("POST")
	// TODO: change to keyboard-raw
	sub.HandleFunc("/controls/keyboard_raw/{key}", control.HandleRawKeyboard(kbd, logger)).Methods("POST")

	sub.HandleFunc("/menu/view/", menu.ListFolder(logger)).Methods("GET")
	sub.HandleFunc("/menu/view/{path:.*}", menu.ListFolder(logger)).Methods("GET")

	sub.HandleFunc("/settings/inis", settings.HandleListInis(logger)).Methods("GET")
	sub.HandleFunc("/settings/inis", settings.HandleSetActiveIni(logger)).Methods("PUT")
	sub.HandleFunc("/settings/inis/1", settings.HandleLoadIni(logger, 1)).Methods("GET")
	sub.HandleFunc("/settings/inis/1", settings.HandleSaveIni(logger, 1)).Methods("PUT")
	sub.HandleFunc("/settings/inis/2", settings.HandleLoadIni(logger, 2)).Methods("GET")
	sub.HandleFunc("/settings/inis/2", settings.HandleSaveIni(logger, 2)).Methods("PUT")
	sub.HandleFunc("/settings/inis/3", settings.HandleLoadIni(logger, 3)).Methods("GET")
	sub.HandleFunc("/settings/inis/3", settings.HandleSaveIni(logger, 3)).Methods("PUT")
	sub.HandleFunc("/settings/inis/4", settings.HandleLoadIni(logger, 4)).Methods("GET")
	sub.HandleFunc("/settings/inis/4", settings.HandleSaveIni(logger, 4)).Methods("PUT")

	sub.HandleFunc("/settings/cores/menu", settings.HandleSetMenuBackgroundMode(logger)).Methods("PUT")
	sub.HandleFunc("/settings/remote/restart", settings.HandleRestartRemote()).Methods("POST")
	sub.HandleFunc("/settings/remote/log", settings.HandleDownloadRemoteLog(logger)).Methods("GET")
}

func appHandler(rw http.ResponseWriter, req *http.Request) {
	if !strings.HasPrefix(req.URL.Path, "/") {
		req.URL.Path = "/" + req.URL.Path
	}

	build, err := fs.Sub(client, "_client/build")
	if err != nil {
		logger.Error("could not create client sub fs: %s", err)
		return
	}

	filePath := strings.TrimLeft(path.Clean(req.URL.Path), "/")
	if _, err := build.Open(filePath); err != nil {
		req.URL.Path = "/"
	}

	http.FileServer(http.FS(build)).ServeHTTP(rw, req)
}

func tryAddStartup(stdscr *gc.Window) error {
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
		defer func(win *gc.Window) {
			err := win.Delete()
			if err != nil {
				logger.Error("failed to delete window: %s", err)
			}
		}(win)

		var ch gc.Key
		selected := 0

		for {
			win.MovePrint(1, 3, "Add Remote service to MiSTer startup?")
			win.MovePrint(2, 2, "This won't impact MiSTer's performance.")
			curses.DrawActionButtons(win, []string{"Yes", "No"}, selected, 10)

			win.NoutRefresh()
			err := gc.Update()
			if err != nil {
				return err
			}

			ch = win.GetChar()

			if ch == gc.KEY_LEFT {
				if selected == 0 {
					selected = 1
				} else if selected == 1 {
					selected = 0
				}
			} else if ch == gc.KEY_RIGHT {
				if selected == 0 {
					selected = 1
				} else if selected == 1 {
					selected = 0
				}
			} else if ch == gc.KEY_ENTER || ch == 10 || ch == 13 {
				break
			} else if ch == gc.KEY_ESC {
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

func displayServiceInfo(stdscr *gc.Window, service *service.Service) (int, error) {
	width := 57
	height := 10

	win, err := curses.NewWindow(stdscr, height, width, "", -1)
	if err != nil {
		return displayNothing, err
	}
	defer func(win *gc.Window) {
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

	var ch gc.Key
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
			printCenter(6, "It's safe to exit, the service will continue running.")
		}

		clearLine(8)
		curses.DrawActionButtons(win, []string{toggleText, "Restart", "Uninstall", "Exit"}, selected, 1)

		win.NoutRefresh()
		err := gc.Update()
		if err != nil {
			return displayNothing, err
		}

		ch = win.GetChar()

		if ch == gc.KEY_LEFT {
			if selected == 0 {
				selected = 3
			} else {
				selected--
			}
		} else if ch == gc.KEY_RIGHT {
			if selected == 3 {
				selected = 0
			} else {
				selected++
			}
		} else if ch == gc.KEY_ENTER || ch == 10 || ch == 13 {
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
		} else if ch == gc.KEY_ESC {
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

func main() {
	svcOpt := flag.String("service", "", "manage playlog service (start, stop, restart, status)")
	uninstallOpt := flag.Bool("uninstall", false, "uninstall MiSTer Remote")
	flag.Parse()

	cfg, err := config.LoadUserConfig(appName, &config.UserConfig{})
	if err != nil {
		logger.Error("error loading user config: %s", err)
		fmt.Println("Error loading config file:", err)
		os.Exit(1)
	}

	svc, err := service.NewService(service.ServiceArgs{
		Name:   appName,
		Logger: logger,
		Entry: func() (func() error, error) {
			return startService(logger, cfg)
		},
	})
	if err != nil {
		logger.Error("creating service: %s", err)
		fmt.Println("Error creating service:", err)
		os.Exit(1)
	}

	if *uninstallOpt {
		uninstallService(svc)
		os.Exit(0)
	}

	svc.ServiceHandler(svcOpt)

	if !svc.Running() {
		err := svc.Start()
		if err != nil {
			logger.Error("starting service: %s", err)
			fmt.Println("Error starting service:", err)
			os.Exit(1)
		}
	}

	interactive := true

	stdscr, err := curses.Setup()
	if err != nil {
		logger.Error("starting curses: %s", err)
		interactive = false
	}
	defer gc.End()

	if interactive {
		err = tryAddStartup(stdscr)
		if err != nil {
			gc.End()
			logger.Error("adding startup: %s", err)

			if errors.As(err, &curses.SetupWindowError{}) {
				interactive = false
			} else {
				fmt.Println("Error adding to startup:", err)
			}
		}
	}

	if interactive {
		action, err := displayServiceInfo(stdscr, svc)
		if err != nil {
			gc.End()
			logger.Error("displaying service info: %s", err)

			if errors.As(err, &curses.SetupWindowError{}) {
				interactive = false
			} else {
				fmt.Println("Error displaying service info:", err)
			}
		} else if action == displayUninstall {
			gc.End()
			uninstallService(svc)
			os.Exit(0)
		}
	}

	if !interactive {
		tryNonInteractiveAddToStartup(true)
		displayNonInteractiveServiceInfo(svc)
	}
}
