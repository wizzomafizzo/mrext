package main

import (
	"embed"
	"flag"
	"fmt"
	"github.com/wizzomafizzo/mrext/cmd/remote/control"
	"github.com/wizzomafizzo/mrext/cmd/remote/games"
	"github.com/wizzomafizzo/mrext/cmd/remote/menu"
	"github.com/wizzomafizzo/mrext/cmd/remote/music"
	"github.com/wizzomafizzo/mrext/cmd/remote/screenshots"
	"github.com/wizzomafizzo/mrext/cmd/remote/systems"
	"github.com/wizzomafizzo/mrext/cmd/remote/wallpapers"
	"github.com/wizzomafizzo/mrext/cmd/remote/websocket"
	"github.com/wizzomafizzo/mrext/pkg/input"
	"github.com/wizzomafizzo/mrext/pkg/tracker"
	"io/fs"
	"net/http"
	"os"
	"path"
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

	sub.HandleFunc("/wallpapers", wallpapers.AllWallpapers(logger)).Methods("GET")
	sub.HandleFunc("/wallpapers/{filename:.*}", wallpapers.ViewWallpaper(logger)).Methods("GET")
	sub.HandleFunc("/wallpapers/{filename:.*}", wallpapers.SetWallpaper(logger)).Methods("POST")

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
	sub.HandleFunc("/controls/keyboard_raw/{key}", control.HandleRawKeyboard(kbd, logger)).Methods("POST")

	sub.HandleFunc("/menu/view/", menu.ListFolder(logger)).Methods("GET")
	sub.HandleFunc("/menu/view/{path:.*}", menu.ListFolder(logger)).Methods("GET")
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

func displayServiceInfo(stdscr *gc.Window, service *service.Service) error {
	width := 57
	height := 10

	win, err := curses.NewWindow(stdscr, height, width, "", -1)
	if err != nil {
		return err
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
	selected := 2

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
		curses.DrawActionButtons(win, []string{toggleText, "Restart", "Exit"}, selected, 5)

		win.NoutRefresh()
		err := gc.Update()
		if err != nil {
			return err
		}

		ch = win.GetChar()

		if ch == gc.KEY_LEFT {
			if selected == 0 {
				selected = 2
			} else {
				selected--
			}
		} else if ch == gc.KEY_RIGHT {
			if selected == 2 {
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
			} else {
				break
			}
		} else if ch == gc.KEY_ESC {
			break
		}
	}

	return nil
}

func main() {
	svcOpt := flag.String("service", "", "manage playlog service (start, stop, restart, status)")
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

	svc.ServiceHandler(svcOpt)

	if !svc.Running() {
		err := svc.Start()
		if err != nil {
			logger.Error("starting service: %s", err)
			fmt.Println("Error starting service:", err)
			os.Exit(1)
		}
	}

	stdscr, err := curses.Setup()
	if err != nil {
		logger.Error("starting curses: %s", err)
	}
	defer gc.End()

	err = tryAddStartup(stdscr)
	if err != nil {
		gc.End()
		logger.Error("adding startup: %s", err)
		fmt.Println("Error adding to startup:", err)
	}

	err = displayServiceInfo(stdscr, svc)
	if err != nil {
		gc.End()
		logger.Error("displaying service info: %s", err)
		fmt.Println("Error displaying service info:", err)
	}
}
