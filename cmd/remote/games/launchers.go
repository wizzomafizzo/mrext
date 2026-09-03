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

package games

import (
	"encoding/base64"
	"encoding/json"
	"net/http"
	"path/filepath"
	"strings"

	"github.com/gorilla/mux"
	"github.com/wizzomafizzo/mrext/pkg/config"
	"github.com/wizzomafizzo/mrext/pkg/games"
	"github.com/wizzomafizzo/mrext/pkg/input"
	"github.com/wizzomafizzo/mrext/pkg/mister"
	"github.com/wizzomafizzo/mrext/pkg/service"
	"github.com/wizzomafizzo/mrext/pkg/utils"
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

		err = mister.LaunchGame(cfg, &system, args.Path)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			logger.Error("launch game: during launch: %s", err)
			return
		}
	}
}

func LaunchToken(logger *service.Logger, cfg *config.UserConfig, kbd input.Keyboard) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		data := vars["data"]

		logger.Info("launch token: %s", data)

		encoding := base64.URLEncoding.WithPadding(base64.NoPadding)
		path, err := encoding.DecodeString(data)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			logger.Error("error decoding data: %s", err)
			return
		}

		err = mister.LaunchToken(cfg, false, kbd, string(path))
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			logger.Error("error during launch: %s", err)
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
			cfg,
			&system,
			args.GamePath,
			args.Folder,
			args.Name,
		)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			logger.Error("create launcher: creation: %s", err)
			return
		}
		err = json.NewEncoder(w).Encode(CreateLauncherResponse{Path: mglPath})
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			logger.Error("create launcher: encoding response: %s", err)
		}
	}
}
