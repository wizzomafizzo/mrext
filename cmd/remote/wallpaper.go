package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/gorilla/mux"

	"github.com/wizzomafizzo/mrext/pkg/config"
	"github.com/wizzomafizzo/mrext/pkg/mister"
)

type Wallpaper struct {
	Name     string `json:"name"`
	Filename string `json:"filename"`
	Width    int    `json:"width"`
	Height   int    `json:"height"`
	Active   bool   `json:"active"`
}

const wallpaperFolder = config.SdFolder + "/wallpapers"

func listWallpapers() ([]Wallpaper, error) {
	var wallpapers []Wallpaper

	if _, err := os.Stat(wallpaperFolder); os.IsNotExist(err) {
		return wallpapers, nil
	}

	files, err := os.ReadDir(wallpaperFolder)
	if err != nil {
		return nil, err
	}

	for _, file := range files {
		if file.IsDir() {
			continue
		}

		fn := file.Name()
		if strings.HasSuffix(strings.ToLower(fn), ".png") || strings.HasSuffix(strings.ToLower(fn), ".jpg") {

			wallpapers = append(wallpapers, Wallpaper{
				Name:     strings.TrimSuffix(fn, filepath.Ext(fn)),
				Filename: fn,
			})
		}
	}

	return wallpapers, nil
}

func allWallpapers(w http.ResponseWriter, r *http.Request) {
	var wallpapers []Wallpaper

	wallpapers, err := listWallpapers()
	if err != nil {
		log.Println(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(wallpapers)
}

func viewWallpaper(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	filename := vars["filename"]

	available, err := listWallpapers()
	if err != nil {
		log.Println(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	for _, wallpaper := range available {
		if wallpaper.Filename == filename {
			http.ServeFile(w, r, filepath.Join(wallpaperFolder, wallpaper.Filename))
			return
		}
	}

	http.NotFound(w, r)
}

func setWallpaper(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	filename := vars["filename"]

	var ext string
	if strings.HasSuffix(strings.ToLower(filename), ".png") {
		ext = ".png"
	} else if strings.HasSuffix(strings.ToLower(filename), ".jpg") {
		ext = ".jpg"
	} else {
		http.Error(w, "Invalid file type", http.StatusBadRequest)
		return
	}

	// TODO: need to check if it's a symlink and back up if not
	os.Remove(filepath.Join(config.SdFolder, "menu.png"))
	os.Remove(filepath.Join(config.SdFolder, "menu.jpg"))

	err := os.Symlink(filepath.Join(wallpaperFolder, filename), filepath.Join(config.SdFolder, "menu"+ext))
	if err != nil {
		log.Println(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// TODO: don't do this if the menu isn't running
	mister.LaunchMenu()
}

func deleteWallpaper(w http.ResponseWriter, r *http.Request) {
	//vars := mux.Vars(r)

	//filename := vars["filename"]
}
