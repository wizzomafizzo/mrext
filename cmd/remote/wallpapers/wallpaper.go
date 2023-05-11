package wallpapers

import (
	"encoding/json"
	"fmt"
	"github.com/wizzomafizzo/mrext/pkg/service"
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
	var wps []Wallpaper

	if _, err := os.Stat(wallpaperFolder); os.IsNotExist(err) {
		err := os.Mkdir(wallpaperFolder, 0755)
		if err != nil {
			return nil, err
		}
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

			wps = append(wps, Wallpaper{
				Name:     strings.TrimSuffix(fn, filepath.Ext(fn)),
				Filename: fn,
			})
		}
	}

	return wps, nil
}

func AllWallpapers(logger *service.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, _ *http.Request) {
		var wps []Wallpaper

		wps, err := listWallpapers()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			logger.Error("couldn't list wallpapers: %s", err)
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
				for i, wallpaper := range wps {
					if wallpaper.Filename == filepath.Base(active) {
						wps[i].Active = true
					}
				}
			}
		}

		err = json.NewEncoder(w).Encode(wps)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			logger.Error("couldn't encode wallpapers: %s", err)
			return
		}
	}
}

func ViewWallpaper(logger *service.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		filename := vars["filename"]

		available, err := listWallpapers()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			logger.Error("couldn't list wallpapers: %s", err)
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
}

func SetWallpaper(logger *service.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)

		filename := vars["filename"]

		var ext string
		if strings.HasSuffix(strings.ToLower(filename), ".png") {
			ext = ".png"
		} else if strings.HasSuffix(strings.ToLower(filename), ".jpg") {
			ext = ".jpg"
		} else {
			http.Error(w, "invalid file type", http.StatusBadRequest)
			return
		}

		jpgPath := filepath.Join(config.SdFolder, "menu.jpg")
		if f, err := os.Lstat(jpgPath); err == nil {
			if f.Mode()&os.ModeSymlink == os.ModeSymlink {
				err := os.Remove(jpgPath)
				if err != nil {
					logger.Error("couldn't remove symlink: %s", err)
				}
			} else {
				err := os.Rename(jpgPath, filepath.Join(wallpaperFolder, fmt.Sprintf("menu_%d.jpg", f.ModTime().Unix())))
				if err != nil {
					logger.Error("couldn't rename file: %s", err)
				}
			}
		}

		pngPath := filepath.Join(config.SdFolder, "menu.png")
		if f, err := os.Lstat(pngPath); err == nil {
			if f.Mode()&os.ModeSymlink == os.ModeSymlink {
				err := os.Remove(pngPath)
				if err != nil {
					logger.Error("couldn't remove symlink: %s", err)
				}
			} else {
				err := os.Rename(pngPath, filepath.Join(wallpaperFolder, fmt.Sprintf("menu_%d.jpg", f.ModTime().Unix())))
				if err != nil {
					logger.Error("couldn't rename file: %s", err)
				}
			}
		}

		err := os.Symlink(filepath.Join(wallpaperFolder, filename), filepath.Join(config.SdFolder, "menu"+ext))
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			logger.Error("couldn't set wallpaper symlink: %s", err)
			return
		}

		if _, err := os.Stat(config.CoreNameFile); err == nil {
			name, err := os.ReadFile(config.CoreNameFile)
			if err != nil {
				err := mister.LaunchMenu()
				if err != nil {
					http.Error(w, err.Error(), http.StatusInternalServerError)
					logger.Error("couldn't launch menu: %s", err)
					return
				}
			} else if string(name) == config.MenuCore {
				err := mister.LaunchMenu()
				if err != nil {
					http.Error(w, err.Error(), http.StatusInternalServerError)
					logger.Error("couldn't launch menu: %s", err)
					return
				}
			}
		}
	}
}
