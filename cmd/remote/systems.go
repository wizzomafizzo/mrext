package main

import (
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"

	"github.com/wizzomafizzo/mrext/pkg/games"
	"github.com/wizzomafizzo/mrext/pkg/mister"
)

type System struct {
	Id   string `json:"id"`
	Name string `json:"name"`
}

func allSystems(w http.ResponseWriter, r *http.Request) {
	var systems []System

	for _, system := range games.Systems {
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
		return
	}
}
