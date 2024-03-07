package games

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync"

	"github.com/wizzomafizzo/mrext/cmd/remote/menu"
	"github.com/wizzomafizzo/mrext/cmd/remote/systems"
	"github.com/wizzomafizzo/mrext/cmd/remote/websocket"
	"github.com/wizzomafizzo/mrext/pkg/gamesdb"
	"github.com/wizzomafizzo/mrext/pkg/service"

	"github.com/wizzomafizzo/mrext/pkg/config"
	"github.com/wizzomafizzo/mrext/pkg/games"
)

const pageSize = 500

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

	if gamesdb.DbExists() {
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

func (s *Index) GenerateIndex(logger *service.Logger, cfg *config.UserConfig) {
	if s.Indexing {
		return
	}

	s.mu.Lock()
	s.Indexing = true

	websocket.Broadcast(logger, GetIndexingStatus())

	go func() {
		defer s.mu.Unlock()

		_, err := gamesdb.NewNamesIndex(cfg, games.AllSystems(), func(status gamesdb.IndexStatus) {
			s.TotalSteps = status.Total
			s.CurrentStep = status.Step
			if status.Step == 1 {
				s.CurrentDesc = "Finding games folders..."
			} else if status.Step == status.Total {
				s.CurrentDesc = "Writing database... (" + fmt.Sprint(status.Files) + " games)"
			} else {
				system, err := games.GetSystem(status.SystemId)
				if err != nil {
					s.CurrentDesc = "Indexing " + status.SystemId + "..."
				} else {
					s.CurrentDesc = "Indexing " + system.Name + "..."
				}
			}
			websocket.Broadcast(logger, GetIndexingStatus())
		})
		if err != nil {
			logger.Error("generate index: indexing: %s", err)
		}

		s.Indexing = false
		s.TotalSteps = 0
		s.CurrentStep = 0
		s.CurrentDesc = ""
		websocket.Broadcast(logger, GetIndexingStatus())
	}()
}

func NewIndex() *Index {
	return &Index{}
}

var IndexInstance = NewIndex()

func GenerateSearchIndex(logger *service.Logger, cfg *config.UserConfig) http.HandlerFunc {
	return func(_ http.ResponseWriter, _ *http.Request) {
		IndexInstance.GenerateIndex(logger, cfg)
	}
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
		payload := listSystemsPayload{
			Systems: make([]listSystemsPayloadSystem, 0),
		}

		indexed, err := gamesdb.IndexedSystems()
		if err != nil {
			logger.Error("list systems: getting indexed systems: %s", err)
			indexed = []string{}
		}

		for _, system := range indexed {
			id := system
			sysDef, ok := games.Systems[id]
			if !ok {
				continue
			}

			name, _ := menu.GetNamesTxt(sysDef.Name, "")
			if name == "" {
				name = sysDef.Name
			}

			payload.Systems = append(payload.Systems, listSystemsPayloadSystem{
				Id:   id,
				Name: name,
			})
		}

		err = json.NewEncoder(w).Encode(payload)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			logger.Error("list systems: encoding response: %s", err)
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

		var results = make([]SearchResultGame, 0)
		var search []gamesdb.SearchResult

		if args.System == "all" || args.System == "" {
			search, err = gamesdb.SearchNamesWords(games.AllSystems(), args.Query)
		} else {
			system, errSys := games.GetSystem(args.System)
			if errSys != nil {
				http.Error(w, errSys.Error(), http.StatusBadRequest)
				logger.Error("search games: getting system: %s", err)
				return
			}
			search, err = gamesdb.SearchNamesWords([]games.System{*system}, args.Query)
		}
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			logger.Error("search games: searching: %s", err)
			return
		}

		for _, result := range search {
			system, err := games.GetSystem(result.SystemId)
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
