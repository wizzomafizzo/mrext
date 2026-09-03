// mrext
// Copyright (c) 2026 mrext contributors.
// SPDX-License-Identifier: GPL-3.0-or-later
//
// This file is part of mrext.
//
// mrext is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// mrext is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with mrext. If not, see <http://www.gnu.org/licenses/>.

package wallpapers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/gorilla/mux"
	"github.com/wizzomafizzo/mrext/pkg/config"
	"github.com/wizzomafizzo/mrext/pkg/mister"
	"github.com/wizzomafizzo/mrext/pkg/service"
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
		// #nosec G301 -- MiSTer wallpaper folder must remain world-readable.
		if err := os.Mkdir(wallpaperFolder, 0o755); err != nil {
			return nil, fmt.Errorf("create wallpaper folder: %w", err)
		}
	}

	files, err := os.ReadDir(wallpaperFolder)
	if err != nil {
		return nil, fmt.Errorf("read wallpaper folder: %w", err)
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
	Wallpapers     []Wallpaper `json:"wallpapers"`
	BackgroundMode int         `json:"backgroundMode"`
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
			active, linkErr := os.Readlink(filepath.Join(config.SdFolder, activeFile.Name()))
			if linkErr == nil {
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
		switch {
		case strings.HasSuffix(strings.ToLower(filename), ".png"):
			ext = ".png"
		case strings.HasSuffix(strings.ToLower(filename), ".jpg"):
			ext = ".jpg"
		default:
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
				backupPath := filepath.Join(wallpaperFolder, fmt.Sprintf("menu_%d.jpg", f.ModTime().Unix()))
				renameErr := os.Rename(jpgPath, backupPath)
				if renameErr != nil {
					logger.Error("couldn't rename file: %s", renameErr)
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
				backupPath := filepath.Join(wallpaperFolder, fmt.Sprintf("menu_%d.jpg", f.ModTime().Unix()))
				renameErr := os.Rename(pngPath, backupPath)
				if renameErr != nil {
					logger.Error("couldn't rename file: %s", renameErr)
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
	return func(w http.ResponseWriter, _ *http.Request) {
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
