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
	"errors"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"

	"github.com/wizzomafizzo/mrext/pkg/tracker"
)

func TestInitializeTrackerStateReconcilesChangesDuringWatchSetup(t *testing.T) {
	var calls []string
	activeGame := "game-a"
	var loadedGames []string

	err := initializeTrackerState(
		func() bool { return true },
		func() { calls = append(calls, "core") },
		func() {
			calls = append(calls, "game")
			loadedGames = append(loadedGames, activeGame)
		},
		func() { t.Fatal("enabled active-game tracking was cleared") },
		func() error {
			calls = append(calls, "watch")
			activeGame = "game-b"
			return nil
		},
	)
	if err != nil {
		t.Fatal(err)
	}
	if want := []string{"core", "game", "watch", "core", "game"}; !reflect.DeepEqual(calls, want) {
		t.Fatalf("calls = %v, want %v", calls, want)
	}
	if want := []string{"game-a", "game-b"}; !reflect.DeepEqual(loadedGames, want) {
		t.Fatalf("loaded games = %v, want %v", loadedGames, want)
	}
}

func TestInitializeTrackerStatePreservesDisabledCompatibility(t *testing.T) {
	enabled := false
	cleared := false
	loaded := false

	err := initializeTrackerState(
		func() bool { return enabled },
		func() {},
		func() { loaded = true },
		func() { cleared = true },
		func() error { return nil },
	)
	if err != nil {
		t.Fatal(err)
	}
	if !cleared || loaded {
		t.Fatalf("cleared=%t loaded=%t", cleared, loaded)
	}
}

func TestInitializeTrackerStateStopsAfterWatchFailure(t *testing.T) {
	watchErr := errors.New("watch failed")
	coreLoads := 0
	err := initializeTrackerState(
		func() bool { return true },
		func() { coreLoads++ },
		func() {},
		func() {},
		func() error { return watchErr },
	)
	if !errors.Is(err, watchErr) {
		t.Fatalf("error = %v, want %v", err, watchErr)
	}
	if coreLoads != 1 {
		t.Fatalf("core loads after failed watch = %d, want 1", coreLoads)
	}
}

type fakePlaying struct {
	now        tracker.Playing
	core, game int
}

func (f *fakePlaying) Playing() tracker.Playing { return f.now }

func (f *fakePlaying) SessionTimes() (core, game int) { return f.core, f.game }

var _ playingSource = (*tracker.Tracker)(nil)

func playingRequest(t *testing.T) *http.Request {
	t.Helper()
	return httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/api/games/playing", http.NoBody)
}

func TestHandlePlayingReportsTheTrackerSnapshot(t *testing.T) {
	source := &fakePlaying{
		now: tracker.Playing{
			Core: "MegaDrive", System: "Genesis", SystemName: "Genesis",
			Game: "Genesis/Rocket Knight Adventures.md", GameName: "Rocket Knight Adventures",
			GamePath: "/media/fat/games/Genesis/Rocket Knight Adventures.md",
		},
		core: 12, game: 9,
	}
	rec := httptest.NewRecorder()
	HandlePlaying(source)(rec, playingRequest(t))
	var got map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
		t.Fatal(err)
	}
	want := map[string]any{
		"core": "MegaDrive", "system": "Genesis", "systemName": "Genesis",
		"game": "Genesis/Rocket Knight Adventures.md", "gameName": "Rocket Knight Adventures",
		"coreSessionTime": float64(12), "gameSessionTime": float64(9),
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("playing = %v, want %v", got, want)
	}
}

func TestHandlePlayingBeforeTheTrackerRuns(t *testing.T) {
	rec := httptest.NewRecorder()
	HandlePlaying(&tracker.Tracker{})(rec, playingRequest(t))
	var got PlayingPayload
	if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
		t.Fatal(err)
	}
	if got != (PlayingPayload{}) {
		t.Fatalf("playing = %+v, want empty", got)
	}
}
