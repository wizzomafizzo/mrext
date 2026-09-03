package main

import (
	"bytes"
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
	event := tracker.EventAction{Timestamp: time.Unix(1_700_000_000, 0).UTC(), Action: tracker.EventActionGameStart, Target: game.Id, TotalTime: 21}
	if err := db.UpdateCore(core); err != nil {
		t.Fatal(err)
	}
	if err := db.UpdateGame(game); err != nil {
		t.Fatal(err)
	}
	if err := db.AddEvent(event); err != nil {
		t.Fatal(err)
	}
	if err := db.db.Close(); err != nil {
		t.Fatal(err)
	}

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
	defer reopened.db.Close()
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
	defer db.db.Close()

	game := tracker.GameTime{Id: "legacy", Path: "SNES/game.sfc", Name: "Game", Folder: "SNES", Time: 10}
	if err := db.UpdateGame(game); err != nil {
		t.Fatal(err)
	}
	_, err = db.db.Exec(
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
