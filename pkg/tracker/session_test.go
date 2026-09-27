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
package tracker

import (
	"testing"
	"time"
)

func TestSessionTimesCountOnlyTheCurrentSession(t *testing.T) {
	t.Parallel()

	tr := &Tracker{
		ActiveCore: "SNES",
		ActiveGame: "SNES/Super Metroid.sfc",
		CoreTimes:  map[string]CoreTime{"SNES": {Name: "SNES", Time: 500}},
		GameTimes: map[string]GameTime{
			"SNES/Super Metroid.sfc": {Id: "SNES/Super Metroid.sfc", Time: 340},
		},
		Events: []EventAction{
			// An earlier session of the same game, then the current one,
			// which started with 300 seconds already on the clock.
			{Action: EventActionGameStart, Target: "SNES/Super Metroid.sfc", TotalTime: 0},
			{Action: EventActionGameStop, Target: "SNES/Super Metroid.sfc", TotalTime: 300},
			{Action: EventActionCoreStart, Target: "SNES", TotalTime: 380},
			{Action: EventActionGameStart, Target: "SNES/Super Metroid.sfc", TotalTime: 300},
		},
	}

	tr.updateSessions()
	core, game := tr.SessionTimes()
	if core != 120 {
		t.Errorf("core session = %d, want 120", core)
	}
	if game != 40 {
		t.Errorf("game session = %d, want 40", game)
	}
}

func TestSessionTimesAreZeroWithNothingRunning(t *testing.T) {
	t.Parallel()

	tr := &Tracker{
		CoreTimes: map[string]CoreTime{},
		GameTimes: map[string]GameTime{},
		Events:    []EventAction{{Action: EventActionGameStart, Target: "NES/Metroid.nes"}},
	}

	tr.updateSessions()
	if core, game := tr.SessionTimes(); core != 0 || game != 0 {
		t.Errorf("sessions = %d, %d; want 0, 0", core, game)
	}
}

func TestSessionTimesDoNotWaitForTheLock(t *testing.T) {
	t.Parallel()

	tr := &Tracker{CoreTimes: map[string]CoreTime{}, GameTimes: map[string]GameTime{}}
	tr.coreSession.Store(7)
	tr.mu.Lock()
	defer tr.mu.Unlock()

	done := make(chan int)
	go func() {
		core, _ := tr.SessionTimes()
		done <- core
	}()
	select {
	case core := <-done:
		if core != 7 {
			t.Errorf("core session = %d, want 7", core)
		}
	case <-time.After(time.Second):
		t.Fatal("SessionTimes waited for the tracker lock")
	}
}

func TestActiveSessionReportsTheRunningGame(t *testing.T) {
	t.Parallel()

	tr := &Tracker{
		ActiveGame:     "NES/Metroid.nes",
		ActiveGamePath: "/media/fat/games/NES/Metroid.nes",
		ActiveGameName: "Metroid",
		CoreTimes:      map[string]CoreTime{},
		GameTimes:      map[string]GameTime{"NES/Metroid.nes": {Time: 90}},
		Events:         []EventAction{{Action: EventActionGameStart, Target: "NES/Metroid.nes", TotalTime: 30}},
	}
	if game, _ := tr.ActiveSession(); game.ID != "" {
		t.Errorf("before the first tick = %+v, want empty", game)
	}
	tr.updateSessions()
	want := RunningGame{ID: "NES/Metroid.nes", Path: "/media/fat/games/NES/Metroid.nes", Name: "Metroid"}
	if game, seconds := tr.ActiveSession(); game != want || seconds != 60 {
		t.Errorf("active session = %+v, %d; want %+v, 60", game, seconds, want)
	}
}
