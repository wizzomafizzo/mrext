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
	"context"
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/wizzomafizzo/mrext/pkg/config"
)

// daemon is the part of the service that a reset stops and starts again.
type daemon interface {
	Running() bool
	Stop() error
	Start() error
}

// errDaemonUnknown is a running PlayLog that this process cannot manage.
var errDaemonUnknown = errors.New("a PlayLog service is running that this process does not recognise as its own")

// resetStopTimeout bounds how long a reset waits for the daemon to exit.
// A variable so tests need not wait it out.
var resetStopTimeout = 20 * time.Second

// resetHistory clears the play history in the database at path, after
// copying it to backup.
//
// A running daemon is stopped first and started again after: it keeps the
// totals of every core and game it has seen in memory and writes them back
// as they run, so clearing the tables under it would be undone the next time
// one of them starts. Stopping it also saves the time of what is running
// now, so the backup holds it.
//
// running reports whether a PlayLog daemon runs, found without the app path
// the service takes from the environment. If it runs but svc does not see it,
// svc would neither stop it nor start the right binary, so nothing is done.
func resetHistory(svc daemon, running func() bool, path, backup string) (err error) {
	if running() && !svc.Running() {
		return errDaemonUnknown
	}
	if svc.Running() {
		if stopErr := svc.Stop(); stopErr != nil {
			return fmt.Errorf("stop playlog: %w", stopErr)
		}
		deadline := time.Now().Add(resetStopTimeout)
		for svc.Running() {
			if time.Now().After(deadline) {
				return errors.New("playlog did not stop")
			}
			time.Sleep(100 * time.Millisecond)
		}
		defer func() {
			if startErr := startAndWait(svc); startErr != nil {
				err = errors.Join(err, startErr)
			}
		}()
	}

	if _, statErr := os.Stat(path); statErr != nil {
		if errors.Is(statErr, os.ErrNotExist) {
			return nil
		}
		return fmt.Errorf("check play history: %w", statErr)
	}

	db, err := openPlayLogDbAt(path)
	if err != nil {
		return err
	}
	defer func() { err = errors.Join(err, db.db.Close()) }()

	// An empty history has nothing to keep: copying it would replace the
	// backup of the last history cleared, the only copy left of it.
	empty, err := db.empty()
	if err != nil {
		return err
	}
	if empty {
		return nil
	}

	if copyErr := copyFile(path, backup); copyErr != nil {
		return fmt.Errorf("back up play history: %w", copyErr)
	}
	return db.clear()
}

// empty reports whether no table holds a row.
func (p *playLogDb) empty() (bool, error) {
	ctx := context.Background()
	for _, table := range []string{"events", "core_times", "game_times"} {
		var found int
		// #nosec G202 -- the table names are the constants above.
		err := p.db.QueryRowContext(ctx, "select exists (select 1 from "+table+")").Scan(&found)
		if err != nil {
			return false, fmt.Errorf("check %s: %w", table, err)
		}
		if found != 0 {
			return false, nil
		}
	}
	return true, nil
}

// startAndWait starts the daemon and waits until it runs, so whoever asked
// for the reset sees PlayLog running again once it returns.
func startAndWait(svc daemon) error {
	if err := svc.Start(); err != nil {
		return fmt.Errorf("start playlog: %w", err)
	}
	deadline := time.Now().Add(resetStopTimeout)
	for !svc.Running() {
		if time.Now().After(deadline) {
			return errors.New("playlog did not start")
		}
		time.Sleep(100 * time.Millisecond)
	}
	return nil
}

// clear empties every table in one transaction and compacts the file.
func (p *playLogDb) clear() error {
	ctx := context.Background()
	tx, err := p.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin reset: %w", err)
	}
	for _, table := range []string{"events", "core_times", "game_times"} {
		// #nosec G202 -- the table names are the constants above.
		if _, err = tx.ExecContext(ctx, "delete from "+table); err != nil {
			return errors.Join(fmt.Errorf("clear %s: %w", table, err), tx.Rollback())
		}
	}
	if err = tx.Commit(); err != nil {
		return fmt.Errorf("commit reset: %w", err)
	}
	if _, err = p.db.ExecContext(ctx, "vacuum"); err != nil {
		return fmt.Errorf("compact play log database: %w", err)
	}
	return nil
}

// copyFile copies src to dst through a staged file, so dst is never left
// half written.
func copyFile(src, dst string) error {
	// #nosec G304 -- src is PlayLog's own database.
	data, err := os.ReadFile(src)
	if err != nil {
		return fmt.Errorf("read %s: %w", src, err)
	}
	staged := dst + ".tmp"
	// #nosec G306 G703 -- PlayLog's fixed backup path, same mode as the database.
	if err := os.WriteFile(staged, data, 0o644); err != nil {
		return fmt.Errorf("write %s: %w", staged, err)
	}
	// #nosec G703 -- dst is PlayLog's fixed backup path.
	if err := os.Rename(staged, dst); err != nil {
		return errors.Join(fmt.Errorf("replace %s: %w", dst, err), os.Remove(staged))
	}
	return nil
}

// playLogBackupFile is where a reset keeps the history it cleared.
const playLogBackupFile = config.PlayLogDbFile + ".bak"
