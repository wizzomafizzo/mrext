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

package screenshots

import (
	"encoding/json"
	"fmt"
	"io/fs"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/gorilla/mux"
	"github.com/wizzomafizzo/mrext/pkg/config"
	"github.com/wizzomafizzo/mrext/pkg/service"
)

const screenshotsFolder = config.SdFolder + "/screenshots"

type ScreenshotPayload struct {
	Modified time.Time `json:"modified"`
	Game     string    `json:"game"`
	Filename string    `json:"filename"`
	Path     string    `json:"path"`
	Core     string    `json:"core"`
}

func AllScreenshots(logger *service.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, _ *http.Request) {
		var screenshots []ScreenshotPayload

		err := filepath.WalkDir(screenshotsFolder, func(path string, info fs.DirEntry, err error) error {
			if err != nil {
				return err
			}

			if !info.IsDir() && strings.HasSuffix(info.Name(), ".png") {
				relPath := strings.Replace(path, screenshotsFolder+"/", "", 1)
				if strings.Count(relPath, "/") == 1 {
					core := strings.Split(relPath, "/")[0]

					fd, infoErr := info.Info()
					if infoErr != nil {
						return fmt.Errorf("read screenshot info: %w", infoErr)
					}

					gp := strings.SplitN(info.Name(), "-", 2)
					game := gp[0]
					if len(gp) > 1 && len(gp[1]) > 4 {
						game = gp[1][:len(gp[1])-4]
					}

					screenshots = append(screenshots, ScreenshotPayload{
						Game:     game,
						Filename: info.Name(),
						Path:     relPath,
						Core:     core,
						Modified: fd.ModTime(),
					})
				}
			}

			return nil
		})
		if err != nil {
			logger.Error("all screenshots: %s", err)
		}

		err = json.NewEncoder(w).Encode(screenshots)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			logger.Error("all screenshots: %s", err)
			return
		}
	}
}

func ViewScreenshot(_ *service.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		core, image := vars["core"], vars["image"]
		if filepath.Base(core) != core || filepath.Base(image) != image {
			http.NotFound(w, r)
			return
		}

		root, err := os.OpenRoot(screenshotsFolder)
		if err != nil {
			http.NotFound(w, r)
			return
		}
		defer func() { _ = root.Close() }()
		file, err := root.Open(filepath.Join(core, image))
		if err != nil {
			http.NotFound(w, r)
			return
		}
		defer func() { _ = file.Close() }()
		info, err := file.Stat()
		if err != nil {
			http.NotFound(w, r)
			return
		}
		http.ServeContent(w, r, image, info.ModTime(), file)
	}
}

func TakeScreenshot(logger *service.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, _ *http.Request) {
		screenshot := ScreenshotPayload{}

		cmd, err := os.OpenFile(config.CmdInterface, os.O_RDWR, 0)
		if err != nil {
			logger.Error("take screenshot: open dev: %s", err)
			return
		}
		defer func(cmd *os.File) {
			closeErr := cmd.Close()
			if closeErr != nil {
				logger.Error("take screenshot: close dev: %s", closeErr)
			}
		}(cmd)

		_, err = cmd.WriteString("screenshot\n")
		if err != nil {
			logger.Error("take screenshot: write dev: %s", err)
			return
		}

		// TODO: don't pretend to wait
		time.Sleep(1 * time.Second)

		err = json.NewEncoder(w).Encode(screenshot)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			logger.Error("take screenshot: encode: %s", err)
			return
		}
	}
}

func DeleteScreenshot(logger *service.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)

		core := vars["core"]
		image := vars["image"]

		if filepath.Base(core) != core || filepath.Base(image) != image {
			http.NotFound(w, r)
			return
		}

		root, err := os.OpenRoot(screenshotsFolder)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			logger.Error("open screenshots folder: %s", err)
			return
		}
		defer func() { _ = root.Close() }()
		err = root.Remove(filepath.Join(core, image))
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			logger.Error("delete screenshot: %s", err)
			return
		}
	}
}
