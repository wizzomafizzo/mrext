package games

import (
	"encoding/json"
	"fmt"
	"github.com/wizzomafizzo/mrext/cmd/remote/systems"
	"github.com/wizzomafizzo/mrext/cmd/remote/websocket"
	"github.com/wizzomafizzo/mrext/pkg/service"
	"net/http"
	"sync"

	"github.com/wizzomafizzo/mrext/pkg/config"
	"github.com/wizzomafizzo/mrext/pkg/games"
	"github.com/wizzomafizzo/mrext/pkg/txtindex"
)

const pageSize = 250

type SearchResultGame struct {
	System systems.System `json:"system"`
	Name   string         `json:"name"`
	Path   string         `json:"path"`
}

type SearchResults struct {
	Data     []SearchResultGame `json:"data"`
	Total    int                `json:"total"`
	PageSize int                `json:"pageSize"`
	Page     int                `json:"page"`
}

type Index struct {
	mu          sync.Mutex
	Indexing    bool   `json:"indexing"`
	TotalSteps  int    `json:"totalSteps"`
	CurrentStep int    `json:"currentStep"`
	CurrentDesc string `json:"currentDesc"`
}

func GetIndexingStatus() string {
	status := "indexStatus:"

	if txtindex.Exists() {
		status += "y,"
	} else {
		status += "n,"
	}

	if IndexInstance.Indexing {
		status += "y,"
	} else {
		status += "n,"
	}

	status += fmt.Sprintf(
		"%d,%d,%s",
		IndexInstance.TotalSteps,
		IndexInstance.CurrentStep,
		IndexInstance.CurrentDesc,
	)

	return status
}

func (s *Index) GenerateIndex() {
	// TODO: this probably does need some sort of logging
	if s.Indexing {
		return
	}

	s.mu.Lock()
	s.Indexing = true

	_ = websocket.Broadcast(GetIndexingStatus())

	go func() {
		systemPaths := make(map[string][]string)

		for _, path := range games.GetSystemPaths(games.AllSystems()) {
			systemPaths[path.System.Id] = append(systemPaths[path.System.Id], path.Path)
		}

		s.TotalSteps = 0
		s.CurrentStep = 1
		for _, syss := range systemPaths {
			s.TotalSteps += len(syss)
		}

		s.TotalSteps += 3
		s.CurrentStep = 2
		_ = websocket.Broadcast(GetIndexingStatus())

		files, _ := games.GetAllFiles(systemPaths, func(systemId string, path string) {
			system, _ := games.GetSystem(systemId)
			s.CurrentDesc = system.Name
			s.CurrentStep++
			_ = websocket.Broadcast(GetIndexingStatus())
		})

		s.CurrentDesc = "Writing to database"
		_ = websocket.Broadcast(GetIndexingStatus())
		_ = txtindex.Generate(files, config.SearchDbFile)

		s.CurrentStep++
		s.Indexing = false
		s.TotalSteps = 0
		s.CurrentStep = 0
		s.CurrentDesc = ""
		_ = websocket.Broadcast(GetIndexingStatus())
		s.mu.Unlock()
	}()
}

func NewIndex() *Index {
	return &Index{}
}

var IndexInstance = NewIndex()

func GenerateSearchIndex(_ http.ResponseWriter, _ *http.Request) {
	IndexInstance.GenerateIndex()
}

type listSystemsPayloadSystem struct {
	Id   string `json:"id"`
	Name string `json:"name"`
}

type listSystemsPayload struct {
	Systems []listSystemsPayloadSystem `json:"systems"`
}

func ListSystems(logger *service.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, _ *http.Request) {
		index, err := txtindex.Open(config.SearchDbFile)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			logger.Error("search games: reading index: %s", err)
			return
		}

		payload := listSystemsPayload{
			Systems: make([]listSystemsPayloadSystem, 0),
		}

		for _, system := range index.Systems() {
			id := system
			sysDef, ok := games.Systems[id]
			if !ok {
				continue
			}

			payload.Systems = append(payload.Systems, listSystemsPayloadSystem{
				Id:   id,
				Name: sysDef.Name,
			})
		}

		err = json.NewEncoder(w).Encode(payload)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			logger.Error("search games: encoding response: %s", err)
			return
		}
	}
}

func Search(logger *service.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var args struct {
			Query  string `json:"query"`
			System string `json:"system"`
		}

		err := json.NewDecoder(r.Body).Decode(&args)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			logger.Error("search games: decoding request: %s", err)
			return
		}

		index, err := txtindex.Open(config.SearchDbFile)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			logger.Error("search games: reading index: %s", err)
			return
		}

		var results = make([]SearchResultGame, 0)
		var search []txtindex.SearchResult

		if args.System == "all" || args.System == "" {
			search = index.SearchAllByWords(args.Query)
		} else {
			search = index.SearchSystemByWords(args.System, args.Query)
		}

		for _, result := range search {
			system, err := games.GetSystem(result.System)
			if err != nil {
				continue
			}

			results = append(results, SearchResultGame{
				System: systems.System{
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

		err = json.NewEncoder(w).Encode(&SearchResults{
			Data:     results,
			Total:    total,
			PageSize: pageSize,
			Page:     1,
		})
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			logger.Error("search games: encoding response: %s", err)
			return
		}
	}
}
