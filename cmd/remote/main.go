package main

import (
	"embed"
	"encoding/json"
	"flag"
	"fmt"
	"io/fs"
	"net/http"
	"os"
	"path"
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

type ServerStatus struct {
	Online        bool          `json:"online"`
	SearchService SearchService `json:"searchService"`
	MusicService  MusicService  `json:"musicService"`
}

func getServerStatus(w http.ResponseWriter, r *http.Request) {
	search := searchService
	search.checkIndexReady()

	music := getMusicServiceStatus()

	status := ServerStatus{
		Online:        true,
		SearchService: search,
		MusicService:  music,
	}

	json.NewEncoder(w).Encode(status)
}

func startService(logger *service.Logger, cfg *config.UserConfig) (func() error, error) {
	router := mux.NewRouter()
	setupApi(router.PathPrefix("/api").Subrouter())
	router.PathPrefix("/").Handler(http.HandlerFunc(appHandler))

	cors := cors.New(cors.Options{
		AllowedOrigins: []string{"*"},
		AllowedMethods: []string{"GET", "POST", "DELETE", "PUT"},
	})

	srv := &http.Server{
		Handler:      cors.Handler(router),
		Addr:         ":" + fmt.Sprint(appPort),
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
		srv.Close()
		return nil
	}, nil
}

func setupApi(subrouter *mux.Router) {
	subrouter.HandleFunc("/screenshots", allScreenshots).Methods("GET")
	subrouter.HandleFunc("/screenshots", takeScreenshot).Methods("POST")
	subrouter.HandleFunc("/screenshots/{core}/{image}", viewScreenshot).Methods("GET")
	subrouter.HandleFunc("/screenshots/{core}/{image}", deleteScreenshot).Methods("DELETE")

	subrouter.HandleFunc("/systems", allSystems).Methods("GET")
	subrouter.HandleFunc("/systems/{id}", launchCore).Methods("POST")

	subrouter.HandleFunc("/wallpaper", allWallpapers).Methods("GET")
	subrouter.HandleFunc("/wallpaper/{filename}", viewWallpaper).Methods("GET")
	subrouter.HandleFunc("/wallpaper/{filename}", setWallpaper).Methods("POST")
	subrouter.HandleFunc("/wallpaper/{filename}", deleteWallpaper).Methods("DELETE")

	subrouter.HandleFunc("/music/play", musicPlay).Methods("POST")
	subrouter.HandleFunc("/music/stop", musicStop).Methods("POST")
	subrouter.HandleFunc("/music/next", musicSkip).Methods("POST")
	subrouter.HandleFunc("/music/playback/{playback}", setMusicPlayback).Methods("POST")
	subrouter.HandleFunc("/music/playlist", musicPlaylists).Methods("GET")
	subrouter.HandleFunc("/music/playlist/{playlist}", setMusicPlaylist).Methods("POST")

	subrouter.HandleFunc("/games/search", searchGames).Methods("POST")
	subrouter.HandleFunc("/games/launch", launchGame).Methods("POST")
	subrouter.HandleFunc("/games/index", generateSearchIndex).Methods("POST")

	subrouter.HandleFunc("/server", getServerStatus).Methods("GET")
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
		return err
	}

	if !startup.Exists("mrext/" + appName) {
		win, err := curses.NewWindow(stdscr, 6, 43, "", -1)
		if err != nil {
			return err
		}
		defer win.Delete()

		var ch gc.Key
		selected := 0

		for {
			win.MovePrint(1, 3, "Add Remote service to MiSTer startup?")
			win.MovePrint(2, 2, "This won't impact MiSTer's performance.")
			curses.DrawActionButtons(win, []string{"Yes", "No"}, selected, 10)

			win.NoutRefresh()
			gc.Update()

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
	defer win.Delete()

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
		gc.Update()

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
