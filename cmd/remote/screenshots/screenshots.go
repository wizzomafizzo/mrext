package screenshots

import (
	"encoding/json"
	"github.com/wizzomafizzo/mrext/pkg/service"
	"io/fs"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/gorilla/mux"

	"github.com/wizzomafizzo/mrext/pkg/config"
)

const screenshotsFolder = config.SdFolder + "/screenshots"

type ScreenshotPayload struct {
	Game     string    `json:"game"`
	Filename string    `json:"filename"`
	Path     string    `json:"path"`
	Core     string    `json:"core"`
	Modified time.Time `json:"modified"`
}

func AllScreenshots(logger *service.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var screenshots []ScreenshotPayload

		err := filepath.WalkDir(screenshotsFolder, func(path string, info fs.DirEntry, err error) error {
			if err != nil {
				return err
			}

			if !info.IsDir() && strings.HasSuffix(info.Name(), ".png") {
				path := strings.Replace(path, screenshotsFolder+"/", "", 1)
				if strings.Count(path, "/") == 1 {
					core := strings.Split(path, "/")[0]

					fd, err := info.Info()
					if err != nil {
						return err
					}

					gp := strings.SplitN(info.Name(), "-", 2)
					game := gp[0]
					if len(gp) > 1 && len(gp[1]) > 4 {
						game = gp[1][:len(gp[1])-4]
					}

					screenshots = append(screenshots, ScreenshotPayload{
						Game:     game,
						Filename: info.Name(),
						Path:     path,
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

		core := vars["core"]
		image := vars["image"]

		path := filepath.Join(screenshotsFolder, core, image)

		if _, err := os.Stat(path); os.IsNotExist(err) {
			http.NotFound(w, r)
			return
		}

		http.ServeFile(w, r, path)
	}
}

func TakeScreenshot(logger *service.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		screenshot := ScreenshotPayload{}

		cmd, err := os.OpenFile(config.CmdInterface, os.O_RDWR, 0)
		if err != nil {
			logger.Error("take screenshot: open dev: %s", err)
			return
		}
		defer func(cmd *os.File) {
			err := cmd.Close()
			if err != nil {
				logger.Error("take screenshot: close dev: %s", err)
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

		path := filepath.Join(screenshotsFolder, core, image)

		if _, err := os.Stat(path); os.IsNotExist(err) {
			http.NotFound(w, r)
			return
		}

		err := os.Remove(path)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			logger.Error("delete screenshot: %s", err)
			return
		}
	}
}
