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

// Package playlog serves PlayLog's play history over Remote's API.
//
// PlayLog is the mrext app that records how long each core and game is
// played, in an SQLite database on the SD card. Remote only reads it: the
// file is opened read-only for each request and closed after, so PlayLog
// stays its sole writer and Remote adds no SD-card writes. Without PlayLog
// the status endpoint says so and the others answer with empty results.
package playlog

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/wizzomafizzo/mrext/pkg/tracker"
	_ "modernc.org/sqlite" // PlayLog's database driver.
)

// GameStats is one game's play history.
type GameStats struct {
	LastPlayed *time.Time `json:"lastPlayed"`
	ID         string     `json:"id"`
	Path       string     `json:"path"`
	Name       string     `json:"name"`
	System     string     `json:"system"`
	Time       int        `json:"time"`
	Plays      int        `json:"plays"`
}

// Session is one sitting with one game.
type Session struct {
	Start    time.Time `json:"start"`
	ID       string    `json:"id"`
	Path     string    `json:"path"`
	Name     string    `json:"name"`
	System   string    `json:"system"`
	Duration int       `json:"duration"`
	// open is a session with no stop event yet.
	open bool
}

// SystemStats is the play time of every game of one system.
type SystemStats struct {
	System string `json:"system"`
	Time   int    `json:"time"`
	Games  int    `json:"games"`
}

// DayStats is the play time of sessions started on one day.
type DayStats struct {
	Date string `json:"date"`
	Time int    `json:"time"`
}

// History is everything read from the database, with sessions rebuilt from
// the event log. Sessions are oldest first.
type History struct {
	Games    map[string]*GameStats
	Sessions []Session
	CoreTime int
}

// Open reads the database at path. A missing file is an empty history, not
// an error: PlayLog creates it the first time it runs.
func Open(ctx context.Context, path string) (*History, error) {
	history := &History{Games: map[string]*GameStats{}}
	if _, err := os.Stat(path); errors.Is(err, os.ErrNotExist) {
		return history, nil
	} else if err != nil {
		return nil, fmt.Errorf("stat play log database: %w", err)
	}

	// Read-only, and patient with PlayLog's own writes.
	dsn := "file:" + url.PathEscape(filepath.ToSlash(path)) + "?mode=ro&_pragma=busy_timeout(5000)"
	db, err := sql.Open("sqlite", dsn)
	if err != nil {
		return nil, fmt.Errorf("open play log database: %w", err)
	}
	defer func() { _ = db.Close() }()

	if err := history.readGames(ctx, db); err != nil {
		return nil, err
	}
	if err := history.readCores(ctx, db); err != nil {
		return nil, err
	}
	if err := history.readSessions(ctx, db); err != nil {
		return nil, err
	}
	return history, nil
}

func (h *History) readGames(ctx context.Context, db *sql.DB) error {
	rows, err := db.QueryContext(ctx, "select id, path, name, time from game_times")
	if err != nil {
		return fmt.Errorf("query game times: %w", err)
	}
	defer func() { _ = rows.Close() }()
	for rows.Next() {
		game := &GameStats{}
		if err := rows.Scan(&game.ID, &game.Path, &game.Name, &game.Time); err != nil {
			return fmt.Errorf("scan game time: %w", err)
		}
		game.System = systemOf(game.ID)
		h.Games[game.ID] = game
	}
	if err := rows.Err(); err != nil {
		return fmt.Errorf("read game times: %w", err)
	}
	return nil
}

func (h *History) readCores(ctx context.Context, db *sql.DB) error {
	rows, err := db.QueryContext(ctx, "select time from core_times")
	if err != nil {
		return fmt.Errorf("query core times: %w", err)
	}
	defer func() { _ = rows.Close() }()
	for rows.Next() {
		var seconds int
		if err := rows.Scan(&seconds); err != nil {
			return fmt.Errorf("scan core time: %w", err)
		}
		h.CoreTime += seconds
	}
	if err := rows.Err(); err != nil {
		return fmt.Errorf("read core times: %w", err)
	}
	return nil
}

// readSessions pairs each game start with what ended it. Every event carries
// the game's running total, so a session lasts its stop's total minus its
// start's. A start with no stop (a power cut, or the game still running) is
// closed by the game's next start, or by its saved total if there is none.
func (h *History) readSessions(ctx context.Context, db *sql.DB) error {
	rows, err := db.QueryContext(
		ctx,
		"select timestamp, action, target, total_time from events where action in (?, ?) order by rowid",
		tracker.EventActionGameStart, tracker.EventActionGameStop,
	)
	if err != nil {
		return fmt.Errorf("query events: %w", err)
	}
	defer func() { _ = rows.Close() }()

	type open struct {
		index int
		total int
	}
	opened := map[string]open{}
	for rows.Next() {
		var (
			stamp  time.Time
			action int
			target string
			total  int
		)
		if err := rows.Scan(&stamp, &action, &target, &total); err != nil {
			return fmt.Errorf("scan event: %w", err)
		}
		if previous, ok := opened[target]; ok {
			h.Sessions[previous.index].Duration = max(0, total-previous.total)
			delete(opened, target)
		}
		if action != tracker.EventActionGameStart {
			continue
		}
		session := Session{Start: stamp, ID: target, System: systemOf(target)}
		if game, ok := h.Games[target]; ok {
			session.Path, session.Name = game.Path, game.Name
		} else {
			session.Name = displayName(target)
		}
		opened[target] = open{index: len(h.Sessions), total: total}
		h.Sessions = append(h.Sessions, session)
	}
	if err := rows.Err(); err != nil {
		return fmt.Errorf("read events: %w", err)
	}
	for target, still := range opened {
		h.Sessions[still.index].open = true
		if game, ok := h.Games[target]; ok {
			h.Sessions[still.index].Duration = max(0, game.Time-still.total)
		}
	}

	for i := range h.Sessions {
		session := h.Sessions[i]
		game, ok := h.Games[session.ID]
		if !ok {
			// Events for a game whose time was never saved still count.
			game = &GameStats{ID: session.ID, Name: session.Name, System: session.System}
			h.Games[session.ID] = game
		}
		game.Plays++
		if game.LastPlayed == nil || session.Start.After(*game.LastPlayed) {
			start := session.Start
			game.LastPlayed = &start
		}
	}
	return nil
}

// ApplyLive brings the running game up to date. PlayLog saves a running
// game's time only every few minutes, so its open session and total trail
// the truth by up to that long; Remote's own tracker knows how long the
// game has actually been running. Only an open session of that game is
// extended, and never shortened.
//
// Until that first save a game has no path, and an arcade game only its
// set name, so a client cannot match it to a game it lists. The tracker
// knows both, and fills them in.
func (h *History) ApplyLive(live tracker.RunningGame, seconds int) {
	id := live.ID
	if id == "" {
		return
	}
	if game, ok := h.Games[id]; ok && game.Path == "" && live.Path != "" {
		game.Path = live.Path
		if live.Name != "" && game.Name == displayName(id) {
			game.Name = live.Name
		}
	}
	if seconds <= 0 {
		return
	}
	for i := len(h.Sessions) - 1; i >= 0; i-- {
		session := &h.Sessions[i]
		if session.ID != id {
			continue
		}
		if session.Path == "" {
			if game, ok := h.Games[id]; ok {
				session.Path, session.Name = game.Path, game.Name
			}
		}
		if session.open && seconds > session.Duration {
			if game, ok := h.Games[id]; ok {
				game.Time += seconds - session.Duration
			}
			session.Duration = seconds
		}
		return
	}
}

// Played lists the games with any time or plays.
func (h *History) Played() []GameStats {
	games := make([]GameStats, 0, len(h.Games))
	for _, game := range h.Games {
		if game.Time > 0 || game.Plays > 0 {
			games = append(games, *game)
		}
	}
	return games
}

// SortByTime puts the most played first, then the most recent.
func SortByTime(games []GameStats) {
	sort.SliceStable(games, func(i, j int) bool {
		if games[i].Time != games[j].Time {
			return games[i].Time > games[j].Time
		}
		return after(games[i].LastPlayed, games[j].LastPlayed)
	})
}

// SortByRecent puts the most recently played first.
func SortByRecent(games []GameStats) {
	sort.SliceStable(games, func(i, j int) bool {
		if !sameTime(games[i].LastPlayed, games[j].LastPlayed) {
			return after(games[i].LastPlayed, games[j].LastPlayed)
		}
		return games[i].Time > games[j].Time
	})
}

// Find looks a game up by the path a client launched it from, falling back
// to PlayLog's "System/filename" id, since PlayLog records the path the core
// resolved, which is not always the one that was launched.
func (h *History) Find(path, system string) (GameStats, bool) {
	for _, game := range h.Games {
		if path != "" && game.Path == path {
			return *game, true
		}
	}
	if system != "" && path != "" {
		if game, ok := h.Games[system+"/"+filepath.Base(path)]; ok {
			return *game, true
		}
	}
	return GameStats{}, false
}

func systemOf(id string) string {
	if system, _, found := strings.Cut(id, "/"); found {
		return system
	}
	// Arcade games are keyed on the set name alone.
	return tracker.ArcadeSystem
}

func displayName(id string) string {
	_, file, found := strings.Cut(id, "/")
	if !found {
		return id
	}
	return strings.TrimSuffix(file, filepath.Ext(file))
}

func after(a, b *time.Time) bool {
	switch {
	case a == nil:
		return false
	case b == nil:
		return true
	default:
		return a.After(*b)
	}
}

func sameTime(a, b *time.Time) bool {
	if a == nil || b == nil {
		return a == b
	}
	return a.Equal(*b)
}
