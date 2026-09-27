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
package playlog

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/wizzomafizzo/mrext/pkg/config"
	"github.com/wizzomafizzo/mrext/pkg/mister"
	"github.com/wizzomafizzo/mrext/pkg/service"
	"github.com/wizzomafizzo/mrext/pkg/tracker"
)

const (
	appName      = "playlog"
	defaultTop   = 10
	defaultDays  = 30
	maxDays      = 366
	defaultLimit = 50
	maxLimit     = 500
)

// Status is what a client needs to explain PlayLog: whether it is on the
// card, running, has recorded anything, and whether the two things it needs
// to keep running (MiSTer.ini's recents option and the boot hook) are set.
type Status struct {
	Installed bool `json:"installed"`
	Running   bool `json:"running"`
	Database  bool `json:"database"`
	Recents   bool `json:"recents"`
	Startup   bool `json:"startup"`
}

// Env holds everything the handlers read from the MiSTer, so tests can run
// them against a temporary root.
type Env struct {
	Now     func() time.Time
	Running func() bool
	Recents func() (bool, error)
	Startup func() (bool, error)
	// Live is the running game and how long it has run, from Remote's own
	// tracker. Nil in tests that do not need it.
	Live     func() (tracker.RunningGame, int)
	Logger   *service.Logger
	Database string
	Script   string
}

// DefaultEnv reads the real MiSTer, and the running game from trk.
func DefaultEnv(logger *service.Logger, trk *tracker.Tracker) *Env {
	return &Env{
		Live: func() (tracker.RunningGame, int) {
			if trk == nil {
				return tracker.RunningGame{}, 0
			}
			return trk.ActiveSession()
		},
		Now:     time.Now,
		Running: func() bool { return service.AppRunning(appName) },
		Recents: mister.RecentsOptionEnabled,
		Startup: func() (bool, error) {
			var startup mister.Startup
			if err := startup.Load(); err != nil {
				return false, fmt.Errorf("load startup script: %w", err)
			}
			return startup.Exists("mrext/" + appName), nil
		},
		Logger:   logger,
		Database: config.PlayLogDbFile,
		Script:   filepath.Join(config.ScriptsFolder, appName+".sh"),
	}
}

// HandleStatus serves GET /playlog/status.
func HandleStatus(env *Env) http.HandlerFunc {
	return func(w http.ResponseWriter, _ *http.Request) {
		status := Status{
			Installed: exists(env.Script),
			Running:   env.Running(),
			Database:  exists(env.Database),
		}
		var err error
		if status.Recents, err = env.Recents(); err != nil {
			env.Logger.Warn("playlog status: %s", err)
		}
		if status.Startup, err = env.Startup(); err != nil {
			env.Logger.Warn("playlog status: %s", err)
		}
		writeJSON(w, env, status)
	}
}

// HandleSummary serves GET /playlog/summary.
func HandleSummary(env *Env) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		history, ok := load(w, r, env)
		if !ok {
			return
		}
		query := r.URL.Query()
		top := intParam(query.Get("top"), defaultTop, 1, maxLimit)
		days := intParam(query.Get("days"), defaultDays, 1, maxDays)
		writeJSON(w, env, history.Summarise(env.Now(), zone(query.Get("offset")), top, days))
	}
}

// HandleGames serves GET /playlog/games.
func HandleGames(env *Env) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		history, ok := load(w, r, env)
		if !ok {
			return
		}
		query := r.URL.Query()
		games := history.Played()
		switch query.Get("sort") {
		case "", "time":
			SortByTime(games)
		case "recent":
			SortByRecent(games)
		default:
			http.Error(w, "sort must be time or recent", http.StatusBadRequest)
			return
		}
		limit := intParam(query.Get("limit"), defaultLimit, 1, maxLimit)
		offset := intParam(query.Get("offset"), 0, 0, len(games))
		page := games[offset:min(offset+limit, len(games))]
		writeJSON(w, env, struct {
			Games []GameStats `json:"games"`
			Total int         `json:"total"`
		}{Games: page, Total: len(games)})
	}
}

// HandleGame serves GET /playlog/game. A game never played answers with
// zero time and plays rather than an error, so a client can show "never
// played" without telling it apart from a missing endpoint.
func HandleGame(env *Env) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		query := r.URL.Query()
		path := query.Get("path")
		if path == "" {
			http.Error(w, "path is required", http.StatusBadRequest)
			return
		}
		history, ok := load(w, r, env)
		if !ok {
			return
		}
		game, found := history.Find(path, query.Get("system"))
		if !found {
			game = GameStats{Path: path, System: query.Get("system")}
		}
		writeJSON(w, env, game)
	}
}

// HandleSessions serves GET /playlog/sessions, newest first.
func HandleSessions(env *Env) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		history, ok := load(w, r, env)
		if !ok {
			return
		}
		query := r.URL.Query()
		total := len(history.Sessions)
		limit := intParam(query.Get("limit"), defaultLimit, 1, maxLimit)
		offset := intParam(query.Get("offset"), 0, 0, total)
		sessions := make([]Session, 0, limit)
		for i := total - 1 - offset; i >= 0 && len(sessions) < limit; i-- {
			sessions = append(sessions, history.Sessions[i])
		}
		writeJSON(w, env, struct {
			Sessions []Session `json:"sessions"`
			Total    int       `json:"total"`
		}{Sessions: sessions, Total: total})
	}
}

func load(w http.ResponseWriter, r *http.Request, env *Env) (*History, bool) {
	history, err := Open(r.Context(), env.Database)
	if err != nil {
		env.Logger.Error("read play log: %s", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return nil, false
	}
	// Only while PlayLog records: stopped, it saves none of the running
	// game's time, and a session left open by a crash would take on a
	// later run's seconds.
	if env.Live != nil && env.Running() {
		history.ApplyLive(env.Live())
	}
	return history, true
}

func writeJSON(w http.ResponseWriter, env *Env, value any) {
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(value); err != nil {
		env.Logger.Error("encode play log response: %s", err)
	}
}

func exists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

// intParam parses an optional query parameter, clamped to [low, high].
func intParam(raw string, fallback, low, high int) int {
	value, err := strconv.Atoi(raw)
	if raw == "" || err != nil {
		value = fallback
	}
	return max(low, min(high, value))
}

// zone is the client's zone from its UTC offset in seconds, or the MiSTer's
// own when it sends none.
func zone(offset string) *time.Location {
	seconds, err := strconv.Atoi(offset)
	if offset == "" || err != nil || seconds < -14*3600 || seconds > 14*3600 {
		return time.Now().Location()
	}
	return time.FixedZone("client", seconds)
}
