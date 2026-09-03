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

package menu

import (
	"encoding/json"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/wizzomafizzo/mrext/pkg/config"
	"github.com/wizzomafizzo/mrext/pkg/mister"
	"github.com/wizzomafizzo/mrext/pkg/service"
	"github.com/wizzomafizzo/mrext/pkg/utils"
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
			// #nosec G301 -- MiSTer menu folders must remain world-readable.
			err := os.Mkdir(path, 0o755)
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

		if _, statErr := os.Stat(fromPath); os.IsNotExist(statErr) {
			http.Error(w, statErr.Error(), http.StatusNotFound)
			logger.Error("menu file (%s) does not exist: %s", fromPath, statErr)
			return
		}

		if _, statErr := os.Stat(toPath); statErr == nil {
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
		if err != nil {
			status := http.StatusInternalServerError
			if os.IsNotExist(err) {
				status = http.StatusNotFound
			}
			http.Error(w, err.Error(), status)
			logger.Error("cannot inspect menu file (%s): %s", path, err)
			return
		}

		invalidPath := false
		switch {
		case path == "", path == config.SdFolder, path == config.SdFolder+"/":
			invalidPath = true
		case strings.HasPrefix(path, config.SdFolder+"/MiSTer"):
			invalidPath = true
		case path == config.SdFolder+"/menu.rbf":
			invalidPath = true
		case file.IsDir() && file.Name() != "" && file.Name()[0] != '_':
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
