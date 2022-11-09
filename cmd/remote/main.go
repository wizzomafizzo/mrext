package main

import (
	"embed"
	"io/fs"
	"log"
	"net/http"
	"path"
	"strings"
	"time"

	"github.com/gorilla/mux"
	"github.com/rs/cors"
)

//go:embed _client
var client embed.FS

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
		log.Fatal(err)
	}

	filePath := strings.TrimLeft(path.Clean(req.URL.Path), "/")
	if _, err := build.Open(filePath); err != nil {
		req.URL.Path = "/"
	}

	http.FileServer(http.FS(build)).ServeHTTP(rw, req)
}

func main() {
	router := mux.NewRouter()

	setupApi(router.PathPrefix("/api").Subrouter())

	router.PathPrefix("/").Handler(http.HandlerFunc(appHandler))

	cors := cors.New(cors.Options{
		AllowedOrigins: []string{"*"},
		AllowedMethods: []string{"GET", "POST", "DELETE", "PUT"},
	})

	srv := &http.Server{
		Handler:      cors.Handler(router),
		Addr:         ":8182",
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}

	log.Fatal(srv.ListenAndServe())
}
