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
	"context"
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"
	"time"

	"github.com/wizzomafizzo/mrext/pkg/service"
	"github.com/wizzomafizzo/mrext/pkg/tracker"
)

// PlayLog's schema, as cmd/playlog creates it. Remote must keep reading it.
const schema = `
create table events (
	timestamp timestamp not null, action integer not null, target text not null, total_time integer not null
);
create table core_times (name integer not null unique, time integer not null);
create table game_times (
	id text not null unique, path text not null, name text not null, folder text not null, time integer not null
);`

var base = time.Date(2026, 9, 20, 20, 0, 0, 0, time.UTC)

type event struct {
	target string
	at     time.Duration
	action int
	total  int
}

// fixture writes a database the way PlayLog does: time.Time values through
// the same driver.
func fixture(t *testing.T, games map[string][2]string, times map[string]int, events []event) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "playlog.db")
	db, err := sql.Open("sqlite", path)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = db.Close() }()
	ctx := context.Background()
	if _, err := db.ExecContext(ctx, schema); err != nil {
		t.Fatal(err)
	}
	for id, info := range games {
		_, err := db.ExecContext(ctx, "insert into game_times (id, path, name, folder, time) values (?, ?, ?, ?, ?)",
			id, info[0], info[1], "", times[id])
		if err != nil {
			t.Fatal(err)
		}
	}
	if _, err := db.ExecContext(ctx, "insert into core_times (name, time) values (?, ?)", "SNES", 5000); err != nil {
		t.Fatal(err)
	}
	for _, ev := range events {
		_, err := db.ExecContext(ctx, "insert into events (timestamp, action, target, total_time) values (?, ?, ?, ?)",
			base.Add(ev.at), ev.action, ev.target, ev.total)
		if err != nil {
			t.Fatal(err)
		}
	}
	return path
}

const (
	metroid = "SNES/Super Metroid.sfc"
	mario   = "SNES/Super Mario World.sfc"
	sf2     = "sf2"
)

func library(t *testing.T) string {
	t.Helper()
	return fixture(t,
		map[string][2]string{
			metroid: {"/media/fat/games/SNES/Super Metroid.sfc", "Super Metroid"},
			mario:   {"/media/fat/games/SNES/Super Mario World.sfc", "Super Mario World"},
			sf2:     {"/media/fat/_Arcade/Street Fighter II.mra", "Street Fighter II"},
		},
		map[string]int{metroid: 3000, mario: 600, sf2: 900},
		[]event{
			{metroid, 0, tracker.EventActionGameStart, 0},
			{metroid, time.Hour, tracker.EventActionGameStop, 1800},
			// Power cut: no stop, closed by the next start of the same game.
			{metroid, 24 * time.Hour, tracker.EventActionGameStart, 1800},
			{mario, 24*time.Hour + 5*time.Minute, tracker.EventActionGameStart, 0},
			{mario, 24*time.Hour + 15*time.Minute, tracker.EventActionGameStop, 600},
			{metroid, 48 * time.Hour, tracker.EventActionGameStart, 2400},
			{metroid, 48*time.Hour + 10*time.Minute, tracker.EventActionGameStop, 3000},
			// Still running: closed by the saved total.
			{sf2, 48*time.Hour + 20*time.Minute, tracker.EventActionGameStart, 300},
			{"SNES", 50 * time.Hour, tracker.EventActionCoreStart, 0},
		},
	)
}

func TestOpenRebuildsSessionsAndPlays(t *testing.T) {
	t.Parallel()
	history, err := Open(context.Background(), library(t))
	if err != nil {
		t.Fatal(err)
	}

	durations := make([]int, 0, len(history.Sessions))
	for _, session := range history.Sessions {
		durations = append(durations, session.Duration)
	}
	want := []int{1800, 600, 600, 600, 600}
	if len(durations) != len(want) {
		t.Fatalf("sessions = %v, want %v", durations, want)
	}
	for i := range want {
		if durations[i] != want[i] {
			t.Fatalf("sessions = %v, want %v", durations, want)
		}
	}

	if got := history.Games[metroid].Plays; got != 3 {
		t.Errorf("metroid plays = %d, want 3", got)
	}
	if got := history.Games[sf2].System; got != tracker.ArcadeSystem {
		t.Errorf("sf2 system = %q, want %q", got, tracker.ArcadeSystem)
	}
	if got := history.Games[mario].LastPlayed; got == nil || !got.Equal(base.Add(24*time.Hour+5*time.Minute)) {
		t.Errorf("mario last played = %v", got)
	}
	if history.CoreTime != 5000 {
		t.Errorf("core time = %d, want 5000", history.CoreTime)
	}
}

func TestOpenMissingDatabaseIsEmpty(t *testing.T) {
	t.Parallel()
	history, err := Open(context.Background(), filepath.Join(t.TempDir(), "none.db"))
	if err != nil {
		t.Fatal(err)
	}
	if len(history.Games) != 0 || len(history.Sessions) != 0 {
		t.Fatalf("history = %+v, want empty", history)
	}
}

func TestSortingAndLookup(t *testing.T) {
	t.Parallel()
	history, err := Open(context.Background(), library(t))
	if err != nil {
		t.Fatal(err)
	}

	byTime := history.Played()
	SortByTime(byTime)
	if byTime[0].ID != metroid || byTime[2].ID != mario {
		t.Errorf("by time = %s, %s, %s", byTime[0].ID, byTime[1].ID, byTime[2].ID)
	}
	recent := history.Played()
	SortByRecent(recent)
	if recent[0].ID != sf2 || recent[1].ID != metroid {
		t.Errorf("by recent = %s, %s", recent[0].ID, recent[1].ID)
	}

	if game, ok := history.Find("/media/fat/_Arcade/Street Fighter II.mra", ""); !ok || game.ID != sf2 {
		t.Errorf("find by path = %+v, %v", game, ok)
	}
	// Launched from USB, recorded under the same system and filename.
	if game, ok := history.Find("/media/usb0/games/SNES/Super Metroid.sfc", "SNES"); !ok || game.ID != metroid {
		t.Errorf("find by id = %+v, %v", game, ok)
	}
	if _, ok := history.Find("/media/fat/games/NES/Metroid.nes", "NES"); ok {
		t.Error("found a game never played")
	}
}

func TestSummarise(t *testing.T) {
	t.Parallel()
	history, err := Open(context.Background(), library(t))
	if err != nil {
		t.Fatal(err)
	}
	now := base.Add(72 * time.Hour)
	summary := history.Summarise(now, time.UTC, 2, 7)

	if summary.TotalTime != 4500 || summary.GamesPlayed != 3 || summary.Sessions != 5 {
		t.Errorf("totals = %d, %d, %d", summary.TotalTime, summary.GamesPlayed, summary.Sessions)
	}
	if len(summary.TopGames) != 2 || summary.TopGames[0].ID != metroid {
		t.Errorf("top games = %+v", summary.TopGames)
	}
	if first := summary.TopSystems[0]; first.System != "SNES" || first.Time != 3600 || first.Games != 2 {
		t.Errorf("top systems = %+v", summary.TopSystems)
	}
	if len(summary.Days) != 7 || summary.Days[6].Date != "2026-09-23" || summary.Days[3].Time != 1800 {
		t.Errorf("days = %+v", summary.Days)
	}
	if summary.Hours[20] != 4200 {
		t.Errorf("20:00 = %d, want 4200", summary.Hours[20])
	}
	// Played the 20th, 21st and 22nd; nothing yet on the 23rd.
	if summary.LongestStreak != 3 || summary.CurrentStreak != 3 {
		t.Errorf("streaks = %d current, %d longest", summary.CurrentStreak, summary.LongestStreak)
	}
	if summary.LongestSession == nil || summary.LongestSession.ID != metroid {
		t.Errorf("longest = %+v", summary.LongestSession)
	}

	// The same sessions in a zone four hours ahead land on the next day.
	shifted := history.Summarise(now, time.FixedZone("ahead", 4*3600), 2, 7)
	if shifted.Hours[0] != 4200 {
		t.Errorf("shifted midnight = %d, want 4200", shifted.Hours[0])
	}
}

func testEnv(t *testing.T, database string) *Env {
	t.Helper()
	return &Env{
		Now:      func() time.Time { return base.Add(72 * time.Hour) },
		Running:  func() bool { return true },
		Recents:  func() (bool, error) { return true, nil },
		Startup:  func() (bool, error) { return false, nil },
		Logger:   service.NewLogger("playlog-test"),
		Database: database,
		Script:   filepath.Join(t.TempDir(), "playlog.sh"),
	}
}

func get(t *testing.T, handler http.HandlerFunc, target string, into any) int {
	t.Helper()
	recorder := httptest.NewRecorder()
	handler(recorder, httptest.NewRequestWithContext(context.Background(), http.MethodGet, target, http.NoBody))
	if recorder.Code == http.StatusOK && into != nil {
		if err := json.Unmarshal(recorder.Body.Bytes(), into); err != nil {
			t.Fatalf("decode %s: %v", target, err)
		}
	}
	return recorder.Code
}

func TestHandlers(t *testing.T) {
	t.Parallel()
	env := testEnv(t, library(t))

	var status Status
	get(t, HandleStatus(env), "/playlog/status", &status)
	if status.Installed || !status.Running || !status.Database || !status.Recents || status.Startup {
		t.Errorf("status = %+v", status)
	}

	var games struct {
		Games []GameStats `json:"games"`
		Total int         `json:"total"`
	}
	get(t, HandleGames(env), "/playlog/games?sort=recent&limit=1&offset=1", &games)
	if games.Total != 3 || len(games.Games) != 1 || games.Games[0].ID != metroid {
		t.Errorf("games = %+v", games)
	}
	if code := get(t, HandleGames(env), "/playlog/games?sort=name", nil); code != http.StatusBadRequest {
		t.Errorf("bad sort = %d", code)
	}

	var game GameStats
	get(t, HandleGame(env), "/playlog/game?path=/media/fat/games/NES/Metroid.nes&system=NES", &game)
	if game.Plays != 0 || game.Time != 0 || game.LastPlayed != nil {
		t.Errorf("unplayed game = %+v", game)
	}
	if code := get(t, HandleGame(env), "/playlog/game", nil); code != http.StatusBadRequest {
		t.Errorf("missing path = %d", code)
	}

	var sessions struct {
		Sessions []Session `json:"sessions"`
		Total    int       `json:"total"`
	}
	get(t, HandleSessions(env), "/playlog/sessions?limit=2", &sessions)
	if sessions.Total != 5 || len(sessions.Sessions) != 2 || sessions.Sessions[0].ID != sf2 {
		t.Errorf("sessions = %+v", sessions)
	}

	var summary Summary
	get(t, HandleSummary(env), "/playlog/summary?days=3&offset=7200", &summary)
	if len(summary.Days) != 3 || summary.TotalTime != 4500 {
		t.Errorf("summary = %+v", summary)
	}
}

func TestHandlersWithoutDatabase(t *testing.T) {
	t.Parallel()
	env := testEnv(t, filepath.Join(t.TempDir(), "missing.db"))

	var summary Summary
	if code := get(t, HandleSummary(env), "/playlog/summary", &summary); code != http.StatusOK {
		t.Fatalf("summary = %d", code)
	}
	if summary.Sessions != 0 || len(summary.Days) != defaultDays || summary.TopGames == nil {
		t.Errorf("empty summary = %+v", summary)
	}
}

func TestApplyLiveNamesAGameNotSavedYet(t *testing.T) {
	t.Parallel()
	// PlayLog has recorded the start of 1943mii but not saved its time, so
	// the game has only its set name.
	path := fixture(t, nil, nil, []event{{"1943mii", 0, tracker.EventActionGameStart, 0}})
	history, err := Open(context.Background(), path)
	if err != nil {
		t.Fatal(err)
	}
	history.ApplyLive(tracker.RunningGame{
		ID:   "1943mii",
		Path: "/media/fat/_Arcade/1943 Mark II (US).mra",
		Name: "1943 Mark II (US)",
	}, 120)
	game := history.Games["1943mii"]
	if game.Path != "/media/fat/_Arcade/1943 Mark II (US).mra" || game.Name != "1943 Mark II (US)" {
		t.Errorf("game = %+v, want the tracker's path and name", game)
	}
	last := history.Sessions[len(history.Sessions)-1]
	if last.Path != game.Path || last.Duration != 120 {
		t.Errorf("running session = %+v, want the path and 120s", last)
	}
}

func TestApplyLiveExtendsTheRunningSession(t *testing.T) {
	t.Parallel()
	history, err := Open(context.Background(), library(t))
	if err != nil {
		t.Fatal(err)
	}
	// sf2 is still running: PlayLog last saved 900s, 600 of them this session.
	history.ApplyLive(tracker.RunningGame{ID: sf2}, 1000)
	last := history.Sessions[len(history.Sessions)-1]
	if last.ID != sf2 || last.Duration != 1000 {
		t.Errorf("running session = %+v, want sf2 at 1000s", last)
	}
	if got := history.Games[sf2].Time; got != 1300 {
		t.Errorf("sf2 time = %d, want 1300", got)
	}

	// A closed session and a shorter live time change nothing.
	history.ApplyLive(tracker.RunningGame{ID: metroid}, 9999)
	history.ApplyLive(tracker.RunningGame{ID: sf2}, 10)
	if got := history.Games[metroid].Time; got != 3000 {
		t.Errorf("metroid time = %d, want 3000", got)
	}
	if got := history.Games[sf2].Time; got != 1300 {
		t.Errorf("sf2 time after a shorter live time = %d, want 1300", got)
	}
}

func TestHandlersIgnoreLiveTimeWhilePlayLogIsStopped(t *testing.T) {
	t.Parallel()
	env := testEnv(t, library(t))
	env.Running = func() bool { return false }
	env.Live = func() (tracker.RunningGame, int) { return tracker.RunningGame{ID: sf2}, 1000 }

	var game GameStats
	get(t, HandleGame(env), "/playlog/game?path=/media/fat/_Arcade/Street%20Fighter%20II.mra", &game)
	if game.Time != 900 {
		t.Errorf("game time with PlayLog stopped = %d, want the saved 900", game.Time)
	}
}

func TestHandlersIncludeTheRunningSession(t *testing.T) {
	t.Parallel()
	env := testEnv(t, library(t))
	env.Live = func() (tracker.RunningGame, int) { return tracker.RunningGame{ID: sf2}, 1000 }

	var game GameStats
	get(t, HandleGame(env), "/playlog/game?path=/media/fat/_Arcade/Street%20Fighter%20II.mra", &game)
	if game.Time != 1300 {
		t.Errorf("live game time = %d, want 1300", game.Time)
	}
	var summary Summary
	get(t, HandleSummary(env), "/playlog/summary", &summary)
	if summary.TotalTime != 4900 {
		t.Errorf("live total = %d, want 4900", summary.TotalTime)
	}
}
