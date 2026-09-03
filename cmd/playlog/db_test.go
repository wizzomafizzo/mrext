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

package main

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/wizzomafizzo/mrext/pkg/tracker"
)

func TestPlayLogDatabaseRemainsSQLiteCompatible(t *testing.T) {
	t.Parallel()

	path := filepath.Join(t.TempDir(), "playlog.db")
	db, err := openPlayLogDbAt(path)
	if err != nil {
		t.Fatal(err)
	}
	core := tracker.CoreTime{Name: "SNES", Time: 42}
	game := tracker.GameTime{Id: "game-id", Path: "SNES/game.sfc", Name: "Game", Folder: "SNES", Time: 21}
	event := tracker.EventAction{
		Timestamp: time.Unix(1_700_000_000, 0).UTC(),
		Action:    tracker.EventActionGameStart,
		Target:    game.Id,
		TotalTime: 21,
	}
	if updateErr := db.UpdateCore(core); updateErr != nil {
		t.Fatal(updateErr)
	}
	if updateErr := db.UpdateGame(game); updateErr != nil {
		t.Fatal(updateErr)
	}
	if addErr := db.AddEvent(&event); addErr != nil {
		t.Fatal(addErr)
	}
	if closeErr := db.db.Close(); closeErr != nil {
		t.Fatal(closeErr)
	}

	// #nosec G304 -- path is controlled by this test's temporary directory.
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.HasPrefix(data, []byte("SQLite format 3\x00")) {
		t.Fatal("playlog database no longer uses SQLite file format")
	}

	reopened, err := openPlayLogDbAt(path)
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		if closeErr := reopened.db.Close(); closeErr != nil {
			t.Errorf("close reopened database: %v", closeErr)
		}
	}()
	fixed, err := reopened.FixPowerLoss()
	if err != nil {
		t.Fatal(err)
	}
	if !fixed {
		t.Fatal("expected recovery event for interrupted game session")
	}
	gotCore, err := reopened.GetCore(core.Name)
	if err != nil {
		t.Fatal(err)
	}
	if gotCore != core {
		t.Fatalf("core = %#v, want %#v", gotCore, core)
	}
	gotGame, err := reopened.GetGame(game.Id)
	if err != nil {
		t.Fatal(err)
	}
	if gotGame != game {
		t.Fatalf("game = %#v, want %#v", gotGame, game)
	}
}

func TestPlayLogReadsLegacyGoSQLiteTimestamp(t *testing.T) {
	t.Parallel()

	db, err := openPlayLogDbAt(filepath.Join(t.TempDir(), "playlog.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		if closeErr := db.db.Close(); closeErr != nil {
			t.Errorf("close database: %v", closeErr)
		}
	}()

	game := tracker.GameTime{Id: "legacy", Path: "SNES/game.sfc", Name: "Game", Folder: "SNES", Time: 10}
	if updateErr := db.UpdateGame(game); updateErr != nil {
		t.Fatal(updateErr)
	}
	_, err = db.db.ExecContext(
		context.Background(),
		"insert into events (timestamp, action, target, total_time) values (?, ?, ?, ?)",
		"2023-11-14 22:13:20+00:00",
		tracker.EventActionGameStart,
		game.Id,
		game.Time,
	)
	if err != nil {
		t.Fatal(err)
	}
	fixed, err := db.FixPowerLoss()
	if err != nil {
		t.Fatal(err)
	}
	if !fixed {
		t.Fatal("expected legacy timestamp event to be recovered")
	}
}
