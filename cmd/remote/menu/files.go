package menu

import (
	"encoding/json"
	"github.com/wizzomafizzo/mrext/pkg/service"
	"github.com/wizzomafizzo/mrext/pkg/utils"
	"net/http"
	"os"
	"path/filepath"
)

const CreateTypeFolder = "folder"

func cleanPath(path string) string {
	path = removeRoot.ReplaceAllLiteralString(path, "")
	path = filepath.Clean(path)
	path = filepath.Join(menuRoot, path)
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
			Folder  string `json:"folder"`
			OldName string `json:"oldName"`
			NewName string `json:"newName"`
		}

		err := json.NewDecoder(r.Body).Decode(&args)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			logger.Error("error decoding request: %s", err)
			return
		}

		folder := cleanPath(args.Folder)
		oldName := utils.StripBadFileChars(args.OldName)
		newName := utils.StripBadFileChars(args.NewName)
		oldPath := filepath.Join(folder, oldName)
		newPath := filepath.Join(folder, newName)

		if oldPath == newPath {
			return
		}

		if _, err := os.Stat(oldPath); os.IsNotExist(err) {
			http.Error(w, err.Error(), http.StatusNotFound)
			logger.Error("menu file (%s) does not exist: %s", oldPath, err)
			return
		}

		if _, err := os.Stat(newPath); err == nil {
			http.Error(w, "file already exists", http.StatusInternalServerError)
			logger.Error("error renaming file: file already exists")
			return
		}

		logger.Info("renaming file: %s -> %s", oldPath, newPath)

		err = os.Rename(oldPath, newPath)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			logger.Error("error renaming file: %s", err)
			return
		}
	}
}
