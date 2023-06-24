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
	wps := make([]Wallpaper, 0)

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

type AllWallpapersPayload struct {
	Active         string      `json:"active"`
	BackgroundMode int         `json:"backgroundMode"`
	Wallpapers     []Wallpaper `json:"wallpapers"`
}

func AllWallpapersHandler(logger *service.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, _ *http.Request) {
		wps, err := listWallpapers()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			logger.Error("couldn't list wallpapers: %s", err)
			return
		}

		payload := AllWallpapersPayload{
			Wallpapers: wps,
		}

		// TODO: check for file not found
		activeFile, err := os.Stat(filepath.Join(config.SdFolder, "menu.png"))
		if err != nil {
			activeFile, err = os.Stat(filepath.Join(config.SdFolder, "menu.jpg"))
		}

		if err == nil {
			active, err := os.Readlink(filepath.Join(config.SdFolder, activeFile.Name()))
			if err == nil {
				for i, wallpaper := range payload.Wallpapers {
					if wallpaper.Filename == filepath.Base(active) {
						wps[i].Active = true
						payload.Active = wallpaper.Filename
					}
				}
			}
		}

		cfg, err := mister.ReadMenuConfig()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			logger.Error("couldn't read menu config: %s", err)
			return
		}

		payload.BackgroundMode = cfg.BackgroundMode

		err = json.NewEncoder(w).Encode(payload)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			logger.Error("couldn't encode wallpapers: %s", err)
			return
		}
	}
}

func ViewWallpaperHandler(logger *service.Logger) http.HandlerFunc {
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

func SetWallpaperHandler(logger *service.Logger) http.HandlerFunc {
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

		err = mister.SetMenuBackgroundMode(mister.BackgroundModeWallpaper)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			logger.Error("set menu background mode: %s", err)
			return
		}

		err = mister.RelaunchIfInMenu()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			logger.Error("couldn't relaunch menu: %s", err)
			return
		}
	}
}

func UnsetWallpaperHandler(logger *service.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		activeFile, err := os.Stat(filepath.Join(config.SdFolder, "menu.png"))
		if err != nil {
			activeFile, err = os.Stat(filepath.Join(config.SdFolder, "menu.jpg"))
		}

		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			logger.Error("no active wallpaper set: %s", err)
			return
		}

		lFile, err := os.Lstat(filepath.Join(config.SdFolder, activeFile.Name()))
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			logger.Error("invalid path: %s", err)
			return
		}

		if lFile.Mode()&os.ModeSymlink != os.ModeSymlink {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			logger.Error("not a symlink: %s", err)
			return
		}

		err = os.Remove(filepath.Join(config.SdFolder, activeFile.Name()))
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			logger.Error("couldn't remove symlink: %s", err)
			return
		}

		err = mister.RelaunchIfInMenu()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			logger.Error("couldn't relaunch menu: %s", err)
			return
		}
	}
}

func ActiveWallpaperHandler(logger *service.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		activeFile, err := os.Stat(filepath.Join(config.SdFolder, "menu.png"))
		if err != nil {
			activeFile, err = os.Stat(filepath.Join(config.SdFolder, "menu.jpg"))
		}

		var wallpaper Wallpaper

		if err == nil {
			lFile, err := os.Lstat(filepath.Join(config.SdFolder, activeFile.Name()))
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				logger.Error("invalid path: %s", err)
				return
			}

			if lFile.Mode()&os.ModeSymlink == os.ModeSymlink {
				filename, err := os.Readlink(filepath.Join(config.SdFolder, activeFile.Name()))
				if err != nil {
					http.Error(w, err.Error(), http.StatusInternalServerError)
					logger.Error("couldn't read symlink: %s", err)
					return
				}

				wallpaper = Wallpaper{
					Filename: filepath.Base(filename),
					Active:   true,
				}
			}
		}

		err = json.NewEncoder(w).Encode(wallpaper)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			logger.Error("couldn't encode wallpaper: %s", err)
			return
		}
	}
}
