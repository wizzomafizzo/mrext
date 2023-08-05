package games

import (
	"encoding/base64"
	"encoding/json"
	"github.com/gorilla/mux"
	"github.com/wizzomafizzo/mrext/pkg/config"
	"github.com/wizzomafizzo/mrext/pkg/games"
	"github.com/wizzomafizzo/mrext/pkg/mister"
	"github.com/wizzomafizzo/mrext/pkg/service"
	"github.com/wizzomafizzo/mrext/pkg/utils"
	"net/http"
	"path/filepath"
	"strings"
)

func LaunchGame(logger *service.Logger, cfg *config.UserConfig) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var args struct {
			Path string `json:"path"`
		}

		err := json.NewDecoder(r.Body).Decode(&args)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			logger.Error("launch game: decoding request: %s", err)
			return
		}

		system, err := games.BestSystemMatch(cfg, args.Path)
		if err != nil {
			http.Error(w, "no system found for game", http.StatusBadRequest)
			logger.Error("launch game: no system found for game: %s", args.Path)
			return
		}

		err = mister.LaunchGame(system, args.Path)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			logger.Error("launch game: during launch: %s", err)
			return
		}
	}
}

func LaunchQRGame(logger *service.Logger, cfg *config.UserConfig) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		data := vars["data"]

		path, err := base64.URLEncoding.DecodeString(data)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			logger.Error("launch qr game: decoding data: %s", err)
			return
		}

		system, err := games.BestSystemMatch(cfg, string(path))
		if err != nil {
			http.Error(w, "no system found for game", http.StatusBadRequest)
			logger.Error("launch qr game: no system found for game: %s", path)
			return
		}

		err = mister.LaunchGame(system, string(path))
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			logger.Error("launch qr game: during launch: %s", err)
			return
		}
	}
}

func LaunchFile(logger *service.Logger, cfg *config.UserConfig) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var args struct {
			Path string `json:"path"`
		}

		err := json.NewDecoder(r.Body).Decode(&args)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			logger.Error("launch file: decoding request: %s", err)
			return
		}

		err = mister.LaunchGenericFile(cfg, args.Path)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			logger.Error("launch file: during launch: %s", err)
			return
		}
	}
}

func LaunchMenu(w http.ResponseWriter, _ *http.Request) {
	err := mister.LaunchMenu()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

type CreateLauncherRequest struct {
	GamePath string `json:"gamePath"`
	Folder   string `json:"folder"`
	Name     string `json:"name"`
}

type CreateLauncherResponse struct {
	Path string `json:"path"`
}

func CreateLauncher(logger *service.Logger, cfg *config.UserConfig) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var args CreateLauncherRequest

		err := json.NewDecoder(r.Body).Decode(&args)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			logger.Error("create launcher: decoding request: %s", err)
			return
		}

		//file, err := os.Stat(args.GamePath)
		//if err != nil {
		//	http.Error(w, err.Error(), http.StatusInternalServerError)
		//	logger.Error("create launcher: path is not accessible: %s", err)
		//	return
		//}
		//
		//if file.IsDir() {
		//	http.Error(w, err.Error(), http.StatusInternalServerError)
		//	logger.Error("create launcher: path is a directory")
		//	return
		//}

		system, err := games.BestSystemMatch(cfg, args.GamePath)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			logger.Error("create launcher: unknown file type or folder")
			return
		}

		if !strings.HasPrefix(args.Folder, config.SdFolder) {
			args.Folder = filepath.Join(config.SdFolder, args.Folder)
		}

		args.Name = utils.StripBadFileChars(args.Name)

		mglPath, err := mister.CreateLauncher(
			&system,
			args.GamePath,
			args.Folder,
			args.Name,
		)

		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			logger.Error("create launcher: creation: %s", err)
			return
		} else {
			err = json.NewEncoder(w).Encode(CreateLauncherResponse{
				Path: mglPath,
			})
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				logger.Error("create launcher: encoding response: %s", err)
				return
			}
		}
	}
}
