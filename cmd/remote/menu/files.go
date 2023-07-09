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
