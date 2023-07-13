package menu

import (
	"encoding/json"
	"github.com/wizzomafizzo/mrext/pkg/config"
	"github.com/wizzomafizzo/mrext/pkg/mister"
	"github.com/wizzomafizzo/mrext/pkg/service"
	"github.com/wizzomafizzo/mrext/pkg/utils"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

const CreateTypeFolder = "folder"

func cleanPath(path string) string {
	path = filepath.Clean(path)
	path = removeRoot.ReplaceAllLiteralString(path, "")
	path = filepath.Join(config.SdFolder, path)
	return path
}

func HandleCreateFile(logger *service.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		logger.Info("create menu file request")

		var args struct {
			Type   string `json:"type"`
			Folder string `json:"folder"`
			Name   string `json:"name"`
		}

		err := json.NewDecoder(r.Body).Decode(&args)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			logger.Error("error decoding request: %s", err)
			return
		}

		if args.Type == CreateTypeFolder {
			folder := cleanPath(args.Folder)
			name := "_" + utils.StripBadFileChars(args.Name)
			path := filepath.Join(folder, name)
			logger.Info("creating folder: %s", path)
			err := os.Mkdir(path, 0755)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				logger.Error("error creating folder: %s", err)
				return
			}
		}
	}
}

func HandleRenameFile(logger *service.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		logger.Info("rename menu file request")

		var args struct {
			FromPath string `json:"fromPath"`
			ToPath   string `json:"toPath"`
		}

		err := json.NewDecoder(r.Body).Decode(&args)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			logger.Error("error decoding request: %s", err)
			return
		}

		fromPath := cleanPath(args.FromPath)
		toPath := cleanPath(args.ToPath)

		toParent := filepath.Dir(toPath)
		toFilename := filepath.Base(toPath)
		toFilename = utils.StripBadFileChars(toFilename)

		toPath = filepath.Join(toParent, toFilename)

		if fromPath == toPath {
			return
		}

		if _, err := os.Stat(fromPath); os.IsNotExist(err) {
			http.Error(w, err.Error(), http.StatusNotFound)
			logger.Error("menu file (%s) does not exist: %s", fromPath, err)
			return
		}

		if _, err := os.Stat(toPath); err == nil {
			http.Error(w, "file already exists", http.StatusInternalServerError)
			logger.Error("error renaming file: file already exists")
			return
		}

		logger.Info("renaming file: %s -> %s", fromPath, toPath)

		err = os.Rename(fromPath, toPath)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			logger.Error("error renaming file: %s", err)
			return
		}

		err = mister.TrySetupArcadeCoresLink(filepath.Dir(fromPath))
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			logger.Error("error creating arcade cores link: %s", err)
		}

		err = mister.TrySetupArcadeCoresLink(filepath.Dir(toPath))
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			logger.Error("error creating arcade cores link: %s", err)
		}
	}
}

func HandleDeleteFile(logger *service.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		logger.Info("delete menu file request")

		var args struct {
			Path string `json:"path"`
		}

		err := json.NewDecoder(r.Body).Decode(&args)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			logger.Error("error decoding request: %s", err)
			return
		}

		path := cleanPath(args.Path)

		file, err := os.Stat(path)
		if os.IsNotExist(err) {
			http.Error(w, err.Error(), http.StatusNotFound)
			logger.Error("menu file (%s) does not exist: %s", path, err)
			return
		}

		var invalidPath bool

		if path == "" {
			invalidPath = true
		} else if path == config.SdFolder {
			invalidPath = true
		} else if path == config.SdFolder+"/" {
			invalidPath = true
		} else if strings.HasPrefix(path, config.SdFolder+"/MiSTer") {
			invalidPath = true
		} else if path == config.SdFolder+"/menu.rbf" {
			invalidPath = true
		} else if file.IsDir() && len(file.Name()) > 0 && file.Name()[0] != '_' {
			invalidPath = true
		}

		if invalidPath {
			http.Error(w, "invalid path", http.StatusInternalServerError)
			logger.Error("invalid path: %s", path)
			return
		}

		logger.Info("deleting file: %s", path)

		err = os.RemoveAll(path)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			logger.Error("error deleting file: %s", err)
			return
		}

		err = mister.TrySetupArcadeCoresLink(filepath.Dir(path))
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			logger.Error("error creating arcade cores link: %s", err)
		}
	}
}
