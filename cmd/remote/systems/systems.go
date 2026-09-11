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

package systems

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/gorilla/mux"
	"github.com/wizzomafizzo/mrext/cmd/remote/menu"
	"github.com/wizzomafizzo/mrext/pkg/config"
	"github.com/wizzomafizzo/mrext/pkg/games"
	"github.com/wizzomafizzo/mrext/pkg/mister"
	"github.com/wizzomafizzo/mrext/pkg/service"
	"github.com/wizzomafizzo/mrext/pkg/utils"
)

type System struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Category string `json:"category"`
}

type SystemRBF struct {
	RBF      string `json:"rbf"`
	Name     string `json:"name"`
	Core     string `json:"core"`
	SetName  string `json:"setname"`
	Filename string `json:"filename"`
	Path     string `json:"path"`
	Default  bool   `json:"default"`
}

var ignoreSystems = []string{
	"Arcade",
	"NESMusic",
	"SNESMusic",
}

func ListSystems(logger *service.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, _ *http.Request) {
		var systems []System

		existingSystems := utils.MapKeys(games.SystemsWithRBF())

		for id := range games.Systems {
			system := games.Systems[id]
			if utils.Contains(ignoreSystems, system.Id) {
				continue
			}

			if !utils.Contains(existingSystems, system.Id) {
				continue
			}

			name, _ := menu.GetNamesTxt(system.Name, "")
			if name == "" {
				name = system.Name
			}

			systems = append(systems, System{
				ID:   system.Id,
				Name: name,
				// TODO: error checking
				Category: strings.Split(system.Rbf, "/")[0][1:],
			})
		}

		err := json.NewEncoder(w).Encode(systems)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			logger.Error("list systems: during encode: %s", err)
			return
		}
	}
}

func ListSystemRBFs(cfg *config.UserConfig, logger *service.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := mux.Vars(r)["id"]
		system, err := games.GetSystem(id)
		if err != nil {
			http.NotFound(w, r)
			return
		}

		cores, err := games.SystemLaunchCores(system)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			logger.Error("list system rbfs: %s", err)
			return
		}

		defaultRBF := mister.SystemRBF(cfg, system)
		results := make([]SystemRBF, 0, len(cores))
		for _, core := range cores {
			results = append(results, SystemRBF{
				RBF:      core.Launch,
				Name:     core.Name,
				Core:     core.RBF,
				SetName:  core.SetName,
				Filename: core.Filename,
				Path:     core.Path,
				Default:  !core.IsLauncher() && strings.EqualFold(core.RBF, defaultRBF),
			})
		}

		err = json.NewEncoder(w).Encode(results)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			logger.Error("list system rbfs: during encode: %s", err)
			return
		}
	}
}

func LaunchCore(cfg *config.UserConfig, logger *service.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)

		id := vars["id"]
		system, err := games.GetSystem(id)
		if err != nil {
			http.NotFound(w, r)
			return
		}

		err = mister.LaunchCore(cfg, system)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			logger.Error("launch core: during launch: %s", err)
			return
		}
	}
}
