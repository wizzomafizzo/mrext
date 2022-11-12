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

const pageSize = 250

type SearchResultGame struct {
	System System `json:"system"`
	Name   string `json:"name"`
	Path   string `json:"path"`
}

type SearchResults struct {
	Data     []SearchResultGame `json:"data"`
	Total    int                `json:"total"`
	PageSize int                `json:"pageSize"`
	Page     int                `json:"page"`
}

// TODO: being naughty and using a global with multiple threads
type SearchService struct {
	Ready       bool   `json:"ready"`
	Indexing    bool   `json:"indexing"`
	TotalSteps  int    `json:"totalSteps"`
	CurrentStep int    `json:"currentStep"`
	CurrentDesc string `json:"currentDesc"`
}

// TODO: i think this makes the disk light blink all the time. annoying
func (s *SearchService) checkIndexReady() {
	s.Ready = txtindex.Exists()
}

func (s *SearchService) generateIndex() {
	if s.Indexing {
		return
	}

	s.Indexing = true

	go func() {
		logger.Info("generating search index")
		systemPaths := make(map[string][]string)

		for _, path := range games.GetSystemPaths(games.AllSystems()) {
			systemPaths[path.System.Id] = append(systemPaths[path.System.Id], path.Path)
		}

		s.TotalSteps = 0
		s.CurrentStep = 1
		for _, systems := range systemPaths {
			s.TotalSteps += len(systems)
		}
		s.TotalSteps += 3
		s.CurrentStep = 2
		logger.Info("generating search index: %d steps", s.TotalSteps)

		files, _ := games.GetAllFiles(systemPaths, func(systemId string, path string) {
			logger.Info("generating search index: %s", path)
			system, _ := games.GetSystem(systemId)
			s.CurrentDesc = system.Name
			s.CurrentStep++
		})

		s.CurrentDesc = "Writing to database"
		if err := txtindex.Generate(files, config.SearchDbFile); err != nil {
			logger.Error("error generating search index: %s", err)
		}
		s.CurrentStep++

		s.Indexing = false
		s.TotalSteps = 0
		s.CurrentStep = 0
		logger.Info("search index generated")
	}()
}

var searchService = SearchService{}

func generateSearchIndex(w http.ResponseWriter, r *http.Request) {
	searchService.generateIndex()
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

	var results []SearchResultGame
	search := index.SearchAllByWords(args.Query)

	for _, result := range search {
		system, err := games.GetSystem(result.System)
		if err != nil {
			continue
		}

		results = append(results, SearchResultGame{
			System: System{
				Id:   system.Id,
				Name: system.Name,
			},
			Name: result.Name,
			Path: result.Path,
		})
	}

	total := len(results)

	if len(results) > pageSize {
		results = results[:pageSize]
	}

	json.NewEncoder(w).Encode(&SearchResults{
		Data:     results,
		Total:    total,
		PageSize: pageSize,
		Page:     1,
	})
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
