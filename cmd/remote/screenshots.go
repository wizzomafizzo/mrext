package main

import (
	"encoding/json"
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

type Screenshot struct {
	Game     string    `json:"game"`
	Filename string    `json:"filename"`
	Path     string    `json:"path"`
	Core     string    `json:"core"`
	Modified time.Time `json:"modified"`
}

func allScreenshots(w http.ResponseWriter, r *http.Request) {
	var screenshots []Screenshot

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

				screenshots = append(screenshots, Screenshot{
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

	json.NewEncoder(w).Encode(screenshots)
}

func viewScreenshot(w http.ResponseWriter, r *http.Request) {
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

func takeScreenshot(w http.ResponseWriter, r *http.Request) {
	var screenshot Screenshot

	cmd, err := os.OpenFile(config.CmdInterface, os.O_RDWR, 0)
	if err != nil {
		logger.Error("take screenshot: %s", err)
		return
	}
	defer cmd.Close()

	cmd.WriteString("screenshot\n")

	// TODO: pretend to wait
	time.Sleep(1 * time.Second)

	json.NewDecoder(r.Body).Decode(&screenshot)
}

func deleteScreenshot(w http.ResponseWriter, r *http.Request) {
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
