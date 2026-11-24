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

//nolint:gosec // Tests only operate on temporary fixture paths.
package main

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/wizzomafizzo/mrext/pkg/tracker"
)

type fakeDaemon struct {
	stopErr    error
	startErr   error
	onStart    func()
	calls      []string
	startPolls int
	running    bool
	stuck      bool
}

func (d *fakeDaemon) Running() bool {
	if d.startPolls > 0 {
		d.startPolls--
		d.running = d.startPolls == 0
	}
	return d.running
}

func (d *fakeDaemon) Stop() error {
	d.calls = append(d.calls, "stop")
	if d.stopErr != nil {
		return d.stopErr
	}
	if !d.stuck {
		d.running = false
	}
	return nil
}

func (d *fakeDaemon) Start() error {
	d.calls = append(d.calls, "start")
	if d.onStart != nil {
		d.onStart()
	}
	// A real daemon writes its PID file a moment after Start returns.
	d.startPolls = 3
	return d.startErr
}

// svcRuns is the independent check agreeing with the service's own.
func svcRuns(svc daemon) func() bool { return svc.Running }

// historyFixture writes one core, one game and one event.
func historyFixture(t *testing.T) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "playlog.db")
	db, err := openPlayLogDbAt(path)
	if err != nil {
		t.Fatal(err)
	}
	if err := db.UpdateCore(tracker.CoreTime{Name: "PSX", Time: 600}); err != nil {
		t.Fatal(err)
	}
	game := tracker.GameTime{
		Id: "PSX/Ape Escape.cue", Path: "PSX/Ape Escape.cue", Name: "Ape Escape", Folder: "PSX", Time: 540,
	}
	if err := db.UpdateGame(game); err != nil {
		t.Fatal(err)
	}
	event := tracker.EventAction{Timestamp: time.Now(), Action: tracker.EventActionGameStart, Target: game.Id}
	if err := db.AddEvent(&event); err != nil {
		t.Fatal(err)
	}
	if err := db.db.Close(); err != nil {
		t.Fatal(err)
	}
	return path
}

// rows counts the rows of each table, in the order events, cores, games.
func rows(t *testing.T, path string) [3]int {
	t.Helper()
	db, err := openPlayLogDbAt(path)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = db.db.Close() }()
	var counts [3]int
	for i, table := range []string{"events", "core_times", "game_times"} {
		row := db.db.QueryRowContext(context.Background(), "select count(*) from "+table)
		if err := row.Scan(&counts[i]); err != nil {
			t.Fatal(err)
		}
	}
	return counts
}

func TestResetStopsPlayLogClearsAndStartsItAgain(t *testing.T) {
	path := historyFixture(t)
	backup := path + ".bak"
	var atStart [3]int
	svc := &fakeDaemon{running: true}
	svc.onStart = func() { atStart = rows(t, path) }

	if err := resetHistory(svc, svcRuns(svc), path, backup, t.Logf); err != nil {
		t.Fatal(err)
	}
	if svc.startPolls != 0 || !svc.running {
		t.Fatal("the reset returned before PlayLog was running again")
	}
	if got := svc.calls; len(got) != 2 || got[0] != "stop" || got[1] != "start" {
		t.Fatalf("calls = %v, want stop then start", got)
	}
	if atStart != [3]int{} {
		t.Fatalf("rows when PlayLog started again = %v, want none", atStart)
	}
	if got := rows(t, path); got != [3]int{} {
		t.Fatalf("rows after reset = %v, want none", got)
	}
	if got := rows(t, backup); got != [3]int{1, 1, 1} {
		t.Fatalf("backup rows = %v, want the history before the reset", got)
	}
}

func TestResetWhilePlayLogIsStopped(t *testing.T) {
	path := historyFixture(t)
	svc := &fakeDaemon{}
	if err := resetHistory(svc, svcRuns(svc), path, path+".bak", t.Logf); err != nil {
		t.Fatal(err)
	}
	if len(svc.calls) != 0 {
		t.Fatalf("calls = %v, want PlayLog left stopped", svc.calls)
	}
	if got := rows(t, path); got != [3]int{} {
		t.Fatalf("rows after reset = %v, want none", got)
	}
}

func TestResetWithoutADatabaseCreatesNothing(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "playlog.db")
	if err := resetHistory(&fakeDaemon{}, svcRuns(&fakeDaemon{}), path, path+".bak", t.Logf); err != nil {
		t.Fatal(err)
	}
	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 0 {
		t.Fatalf("folder holds %d entries, want none", len(entries))
	}
}

func TestResetKeepsTheHistoryWhenPlayLogDoesNotStop(t *testing.T) {
	resetStopTimeout = 50 * time.Millisecond
	t.Cleanup(func() { resetStopTimeout = 20 * time.Second })

	for name, svc := range map[string]*fakeDaemon{
		"stop fails":  {running: true, stopErr: errors.New("no pid")},
		"never stops": {running: true, stuck: true},
	} {
		t.Run(name, func(t *testing.T) {
			path := historyFixture(t)
			if err := resetHistory(svc, svcRuns(svc), path, path+".bak", t.Logf); err == nil {
				t.Fatal("reset succeeded")
			}
			if got := rows(t, path); got != [3]int{1, 1, 1} {
				t.Fatalf("rows = %v, want the history untouched", got)
			}
			if _, err := os.Stat(path + ".bak"); !os.IsNotExist(err) {
				t.Fatalf("a backup was written: %v", err)
			}
		})
	}
}

func TestResetStartsPlayLogAgainWhenClearingFails(t *testing.T) {
	path := filepath.Join(t.TempDir(), "playlog.db")
	if err := os.WriteFile(path, []byte("not a database"), 0o644); err != nil {
		t.Fatal(err)
	}
	svc := &fakeDaemon{running: true}
	if err := resetHistory(svc, svcRuns(svc), path, path+".bak", t.Logf); err == nil {
		t.Fatal("reset succeeded")
	}
	if !svc.Running() {
		t.Fatal("PlayLog was left stopped")
	}
}

func TestResetRefusesADaemonItCannotManage(t *testing.T) {
	path := historyFixture(t)
	svc := &fakeDaemon{}
	err := resetHistory(svc, func() bool { return true }, path, path+".bak", t.Logf)
	if !errors.Is(err, errDaemonUnknown) {
		t.Fatalf("err = %v, want errDaemonUnknown", err)
	}
	if len(svc.calls) != 0 {
		t.Fatalf("calls = %v, want none", svc.calls)
	}
	if got := rows(t, path); got != [3]int{1, 1, 1} {
		t.Fatalf("rows = %v, want the history untouched", got)
	}
}

func TestResetTwiceKeepsTheBackup(t *testing.T) {
	path := historyFixture(t)
	backup := path + ".bak"
	for range 2 {
		svc := &fakeDaemon{running: true}
		if err := resetHistory(svc, svcRuns(svc), path, backup, t.Logf); err != nil {
			t.Fatal(err)
		}
		if !svc.running {
			t.Fatal("PlayLog was not started again")
		}
	}
	if got := rows(t, backup); got != [3]int{1, 1, 1} {
		t.Fatalf("backup rows = %v, want the history before the first reset", got)
	}
}

func TestPlayLogDbWaitsForReaders(t *testing.T) {
	path := filepath.Join(t.TempDir(), "playlog.db")
	db, err := openPlayLogDbAt(path)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = db.db.Close() }()
	var timeout int
	if err := db.db.QueryRowContext(context.Background(), "pragma busy_timeout").Scan(&timeout); err != nil {
		t.Fatal(err)
	}
	if timeout != 5000 {
		t.Fatalf("busy_timeout = %d, want 5000", timeout)
	}
	if _, err := os.Stat(path); err != nil {
		t.Fatalf("database not at its path: %s", err)
	}
}
