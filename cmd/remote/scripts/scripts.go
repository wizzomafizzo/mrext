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

package scripts

import (
	"context"
	"encoding/json"
	"net/http"
	"os/exec"
	"path/filepath"

	"github.com/gorilla/mux"
	"github.com/wizzomafizzo/mrext/pkg/config"
	"github.com/wizzomafizzo/mrext/pkg/input"
	"github.com/wizzomafizzo/mrext/pkg/mister"
	"github.com/wizzomafizzo/mrext/pkg/service"
)

func HandleLaunchScript(logger *service.Logger, kbd input.Keyboard) http.HandlerFunc {
	return func(_ http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		filename := vars["filename"]

		logger.Info("launch script request")

		path := filepath.Join(config.ScriptsFolder, filename)
		logger.Info("running script: %s", path)

		go func() {
			err := mister.RunScript(kbd, path)
			if err != nil {
				logger.Error("error running script: %s", err)
			}
		}()
	}
}

func HandleListScripts(logger *service.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, _ *http.Request) {
		logger.Info("list scripts request")

		files, err := mister.GetAllScripts()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			logger.Error("error listing scripts: %s", err)
			return
		}

		var payload struct {
			Scripts   []mister.Script `json:"scripts"`
			CanLaunch bool            `json:"canLaunch"`
		}

		payload.CanLaunch = mister.ScriptCanLaunch()
		payload.Scripts = files

		err = json.NewEncoder(w).Encode(payload)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			logger.Error("error encoding response: %s", err)
			return
		}
	}
}

func HandleOpenScriptsConsole(logger *service.Logger, kbd input.Keyboard) http.HandlerFunc {
	return func(w http.ResponseWriter, _ *http.Request) {
		logger.Info("open scripts console request")

		err := mister.OpenConsole(kbd)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			logger.Error("error opening console: %s", err)
			return
		}

		if mister.IsScriptRunning() {
			err = exec.CommandContext(context.Background(), "chvt", "2").Run()
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				logger.Error("error changing vt: %s", err)
				return
			}
		}
	}
}

func HandleKillActiveScript(logger *service.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, _ *http.Request) {
		logger.Info("kill active script request")

		err := mister.KillActiveScript()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			logger.Error("error killing active script: %s", err)
			return
		}
	}
}
