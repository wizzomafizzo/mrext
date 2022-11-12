package main

import (
	"encoding/json"
	"fmt"
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
		os.Mkdir(wallpaperFolder, 0755)
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

	// TODO: check for file not found
	activeFile, err := os.Stat(filepath.Join(config.SdFolder, "menu.png"))
	if err != nil {
		activeFile, err = os.Stat(filepath.Join(config.SdFolder, "menu.jpg"))
	}

	if err == nil {
		active, err := os.Readlink(filepath.Join(config.SdFolder, activeFile.Name()))
		if err == nil {
			for i, wallpaper := range wallpapers {
				if wallpaper.Filename == filepath.Base(active) {
					wallpapers[i].Active = true
				}
			}
		}
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

	jpgPath := filepath.Join(config.SdFolder, "menu.jpg")
	if f, err := os.Lstat(jpgPath); err == nil {
		if f.Mode()&os.ModeSymlink == os.ModeSymlink {
			os.Remove(jpgPath)
		} else {
			os.Rename(jpgPath, filepath.Join(wallpaperFolder, fmt.Sprintf("menu_%d.jpg", f.ModTime().Unix())))
		}
	}

	pngPath := filepath.Join(config.SdFolder, "menu.png")
	if f, err := os.Lstat(pngPath); err == nil {
		if f.Mode()&os.ModeSymlink == os.ModeSymlink {
			os.Remove(pngPath)
		} else {
			os.Rename(pngPath, filepath.Join(wallpaperFolder, fmt.Sprintf("menu_%d.jpg", f.ModTime().Unix())))
		}
	}

	err := os.Symlink(filepath.Join(wallpaperFolder, filename), filepath.Join(config.SdFolder, "menu"+ext))
	if err != nil {
		log.Println(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if _, err := os.Stat(config.CoreNameFile); err == nil {
		name, err := os.ReadFile(config.CoreNameFile)
		if err != nil {
			mister.LaunchMenu()
		} else if string(name) == config.MenuCore {
			mister.LaunchMenu()
		}
	}
}

func deleteWallpaper(w http.ResponseWriter, r *http.Request) {
	//vars := mux.Vars(r)

	//filename := vars["filename"]
}
