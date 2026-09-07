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
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"sync"

	"github.com/wizzomafizzo/mrext/cmd/remote/menu"
	"github.com/wizzomafizzo/mrext/cmd/remote/systems"
	"github.com/wizzomafizzo/mrext/cmd/remote/websocket"
	"github.com/wizzomafizzo/mrext/pkg/config"
	"github.com/wizzomafizzo/mrext/pkg/games"
	"github.com/wizzomafizzo/mrext/pkg/gamesdb"
	"github.com/wizzomafizzo/mrext/pkg/service"
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
	CurrentDesc string `json:"currentDesc"`
	TotalSteps  int    `json:"totalSteps"`
	CurrentStep int    `json:"currentStep"`
	mu          sync.Mutex
	Indexing    bool `json:"indexing"`
}

func GetIndexingStatus() string {
	return IndexInstance.status(gamesdb.DBExists())
}

func (s *Index) status(exists bool) string {
	s.mu.Lock()
	defer s.mu.Unlock()
	status := "indexStatus:"

	if exists {
		status += "y,"
	} else {
		status += "n,"
	}

	if s.Indexing {
		status += "y,"
	} else {
		status += "n,"
	}

	status += fmt.Sprintf(
		"%d,%d,%s",
		s.TotalSteps,
		s.CurrentStep,
		strings.NewReplacer(",", ";", "\n", " ", "\r", " ").Replace(s.CurrentDesc),
	)

	return status
}

func (s *Index) GenerateIndex(logger *service.Logger, cfg *config.UserConfig) {
	s.start(func(update func(gamesdb.IndexStatus)) (int, error) {
		return gamesdb.RebuildNamesIndex(cfg, games.AllSystems(), update)
	}, func() {
		websocket.Broadcast(logger, s.status(gamesdb.DBExists()))
	}, func(err error) {
		logger.Error("generate index: indexing: %s", err)
	})
}

func (s *Index) start(build func(func(gamesdb.IndexStatus)) (int, error), publish func(), failed func(error)) bool {
	s.mu.Lock()
	if s.Indexing {
		s.mu.Unlock()
		return false
	}
	s.Indexing = true
	s.TotalSteps, s.CurrentStep = 0, 0
	s.CurrentDesc = "Finding games folders..."
	s.mu.Unlock()
	publish()
	go func() {
		_, err := build(func(status gamesdb.IndexStatus) {
			s.mu.Lock()
			s.TotalSteps = status.Total
			s.CurrentStep = status.Step
			switch status.Step {
			case 1:
				s.CurrentDesc = "Finding games folders..."
			case status.Total:
				s.CurrentDesc = "Writing database... (" + strconv.Itoa(status.Files) + " games)"
			default:
				system, lookupErr := games.GetSystem(status.SystemID)
				if lookupErr != nil {
					s.CurrentDesc = "Indexing " + status.SystemID + "..."
				} else {
					s.CurrentDesc = "Indexing " + system.Name + "..."
				}
			}
			s.mu.Unlock()
			publish()
		})
		if err != nil {
			failed(err)
		}
		s.mu.Lock()
		s.Indexing = false
		s.TotalSteps, s.CurrentStep = 0, 0
		s.CurrentDesc = ""
		if err != nil {
			s.CurrentDesc = "Index failed: " + err.Error()
		}
		s.mu.Unlock()
		publish()
	}()
	return true
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
	ID   string `json:"id"`
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
				ID:   id,
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

		results := make([]SearchResultGame, 0)
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
			system, systemErr := games.GetSystem(result.SystemID)
			if systemErr != nil {
				continue
			}

			results = append(results, SearchResultGame{
				System: systems.System{
					ID:   system.Id,
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
