package main

import (
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"

	"github.com/wizzomafizzo/mrext/pkg/games"
	"github.com/wizzomafizzo/mrext/pkg/mister"
	"github.com/wizzomafizzo/mrext/pkg/utils"
)

type System struct {
	Id   string `json:"id"`
	Name string `json:"name"`
}

var ignoreSystems = []string{
	"Arcade",
	"NESMusic",
	"SNESMusic",
}

func allSystems(w http.ResponseWriter, r *http.Request) {
	var systems []System

	for _, system := range games.Systems {
		if utils.Contains(ignoreSystems, system.Id) {
			continue
		}

		systems = append(systems, System{
			Id:   system.Id,
			Name: system.Name,
		})
	}

	json.NewEncoder(w).Encode(systems)
}

func launchCore(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	id := vars["id"]
	system, err := games.GetSystem(id)
	if err != nil {
		http.NotFound(w, r)
		return
	}

	err = mister.LaunchCore(system.Rbf)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		logger.Error("launch core: during launch: %s", err)
		return
	}
}
