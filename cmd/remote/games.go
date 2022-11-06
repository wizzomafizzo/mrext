package main

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/wizzomafizzo/mrext/pkg/config"
	"github.com/wizzomafizzo/mrext/pkg/games"
	"github.com/wizzomafizzo/mrext/pkg/mister"
	"github.com/wizzomafizzo/mrext/pkg/txtindex"
)

type SearchResult struct {
	System System `json:"system"`
	Name   string `json:"name"`
	Path   string `json:"path"`
}

func searchGames(w http.ResponseWriter, r *http.Request) {
	var args struct {
		Query string `json:"query"`
	}

	err := json.NewDecoder(r.Body).Decode(&args)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		log.Println(err)
		return
	}

	index, err := txtindex.Open(config.SearchDbFile)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		log.Println(err)
		return
	}

	var results []SearchResult
	search := index.SearchAllByWords(args.Query)

	for _, result := range search {
		system, err := games.GetSystem(result.System)
		if err != nil {
			continue
		}

		results = append(results, SearchResult{
			System: System{
				Id:   system.Id,
				Name: system.Name,
			},
			Name: result.Name,
			Path: result.Path,
		})
	}

	json.NewEncoder(w).Encode(results)
}

func launchGame(w http.ResponseWriter, r *http.Request) {
	var args struct {
		Path string `json:"path"`
	}

	err := json.NewDecoder(r.Body).Decode(&args)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		log.Println(err)
		return
	}

	systems := games.FolderToSystems(args.Path)
	if len(systems) == 0 {
		http.Error(w, "no system found for game", http.StatusBadRequest)
		log.Println("no system found for game")
		return
	}

	err = mister.LaunchGame(&systems[0], args.Path)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		log.Println(err)
		return
	}
}
