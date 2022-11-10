package main

import (
	"embed"
	"flag"
	"fmt"
	"io/fs"
	"net/http"
	"os"
	"path"
	"strings"
	"time"

	"github.com/wizzomafizzo/mrext/pkg/config"
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
		if err := srv.ListenAndServe(); err != nil {
			logger.Error("critical server error: %s", err)
			os.Exit(1)
		}
	}()

	return func() error {
		srv.Close()
		return nil
	}, nil
}

func tryAddStartup() error {
	var startup mister.Startup

	err := startup.Load()
	if err != nil {
		return err
	}

	if !startup.Exists("mrext/" + appName) {
		if utils.YesOrNoPrompt("Configure Remote to launch on MiSTer startup?") {
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

	subrouter.HandleFunc("/music", musicStatus).Methods("GET")
	subrouter.HandleFunc("/music/play", musicPlay).Methods("POST")
	subrouter.HandleFunc("/music/stop", musicStop).Methods("POST")
	subrouter.HandleFunc("/music/next", musicSkip).Methods("POST")
	subrouter.HandleFunc("/music/playback/{playback}", setMusicPlayback).Methods("POST")
	subrouter.HandleFunc("/music/playlist", musicPlaylists).Methods("GET")
	subrouter.HandleFunc("/music/playlist/{playlist}", setMusicPlaylist).Methods("POST")

	subrouter.HandleFunc("/games/search", searchGames).Methods("POST")
	subrouter.HandleFunc("/games/launch", launchGame).Methods("POST")
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

func main() {
	svcOpt := flag.String("service", "", "manage playlog service (start, stop, restart, status)")
	flag.Parse()

	cfg, err := config.LoadUserConfig(appName, &config.UserConfig{})
	if err != nil {
		logger.Error("error loading user config: %s", err)
		fmt.Println("Error loading config:", err)
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
		logger.Error("error creating service: %s", err)
		fmt.Println("Error creating service:", err)
		os.Exit(1)
	}

	svc.ServiceHandler(svcOpt)

	err = tryAddStartup()
	if err != nil {
		logger.Error("error adding startup: %s", err)
		fmt.Println("Error adding to startup:", err)
	}

	if !svc.Running() {
		err := svc.Start()
		if err != nil {
			logger.Error("error starting service: %s", err)
			fmt.Println("Error starting service:", err)
			os.Exit(1)
		}
	}

	ip, err := utils.GetLocalIp()
	appUrl := ""
	if err != nil {
		logger.Error("could not get local ip: %s", err)
		appUrl = fmt.Sprintf("http://<MiSTer IP>:%d", appPort)
	} else {
		appUrl = fmt.Sprintf("http://%s:%d", ip, appPort)
	}

	fmt.Printf("Remote URL: %s\n", appUrl)
}
