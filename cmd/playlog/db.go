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
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/wizzomafizzo/mrext/pkg/config"
	"github.com/wizzomafizzo/mrext/pkg/tracker"
	_ "modernc.org/sqlite"
)

type playLogDb struct {
	db *sql.DB
}

func openPlayLogDb() (*playLogDb, error) {
	return openPlayLogDbAt(config.PlayLogDbFile)
}

func openPlayLogDbAt(path string) (*playLogDb, error) {
	db, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, fmt.Errorf("open play log database: %w", err)
	}
	pldb := &playLogDb{db: db}
	if err := pldb.setupDb(); err != nil {
		return nil, errors.Join(err, db.Close())
	}
	return pldb, nil
}

func (*playLogDb) NoResults(err error) bool {
	return errors.Is(err, sql.ErrNoRows)
}

func (p *playLogDb) setupDb() error {
	sqlEvents := `create table if not exists events (
		timestamp timestamp not null,
		action integer not null,
		target text not null,
		total_time integer not null
	)`
	_, err := p.db.ExecContext(context.Background(), sqlEvents)
	if err != nil {
		return fmt.Errorf("create events table: %w", err)
	}

	sqlCoreTimes := `create table if not exists core_times (
		name integer not null unique,
		time integer not null
	)`
	_, err = p.db.ExecContext(context.Background(), sqlCoreTimes)
	if err != nil {
		return fmt.Errorf("create core-times table: %w", err)
	}

	sqlGameTimes := `create table if not exists game_times (
		id text not null unique,
		path text not null,
		name text not null,
		folder text not null,
		time integer not null
	)`
	_, err = p.db.ExecContext(context.Background(), sqlGameTimes)
	if err != nil {
		return fmt.Errorf("create game-times table: %w", err)
	}

	return nil
}

func (p *playLogDb) GetCore(name string) (tracker.CoreTime, error) {
	var core tracker.CoreTime

	err := p.db.QueryRowContext(
		context.Background(),
		"select name, time from core_times where name = ?",
		name,
	).Scan(&core.Name, &core.Time)
	if err != nil {
		return core, fmt.Errorf("query core time: %w", err)
	}

	return core, nil
}

func (p *playLogDb) UpdateCore(core tracker.CoreTime) error {
	_, err := p.db.ExecContext(
		context.Background(),
		"insert or replace into core_times (name, time) values (?, ?)",
		core.Name,
		core.Time,
	)
	if err != nil {
		return fmt.Errorf("update core time: %w", err)
	}
	return nil
}

func (p *playLogDb) GetGame(id string) (tracker.GameTime, error) {
	var game tracker.GameTime

	err := p.db.QueryRowContext(
		context.Background(),
		"select id, path, name, folder, time from game_times where id = ?",
		id,
	).Scan(
		&game.Id,
		&game.Path,
		&game.Name,
		&game.Folder,
		&game.Time,
	)
	if err != nil {
		return game, fmt.Errorf("query game time: %w", err)
	}

	return game, nil
}

func (p *playLogDb) UpdateGame(game tracker.GameTime) error {
	_, err := p.db.ExecContext(
		context.Background(),
		"insert or replace into game_times (id, path, name, folder, time) values (?, ?, ?, ?, ?)",
		game.Id,
		game.Path,
		game.Name,
		game.Folder,
		game.Time,
	)
	if err != nil {
		return fmt.Errorf("update game time: %w", err)
	}
	return nil
}

func (p *playLogDb) AddEvent(event *tracker.EventAction) error {
	_, err := p.db.ExecContext(
		context.Background(),
		"insert into events (timestamp, action, target, total_time) values (?, ?, ?, ?)",
		event.Timestamp,
		event.Action,
		event.Target,
		event.TotalTime,
	)
	if err != nil {
		return fmt.Errorf("add play event: %w", err)
	}
	return nil
}

func (p *playLogDb) topCores(n int) ([]tracker.CoreTime, error) {
	rows, err := p.db.QueryContext(
		context.Background(),
		"select name, time from core_times order by time desc limit ?",
		n,
	)
	if err != nil {
		return nil, fmt.Errorf("query top cores: %w", err)
	}
	defer func() { _ = rows.Close() }()

	var cores []tracker.CoreTime
	for rows.Next() {
		var core tracker.CoreTime
		err = rows.Scan(&core.Name, &core.Time)
		if err != nil {
			return nil, fmt.Errorf("scan core time: %w", err)
		}

		cores = append(cores, core)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate core times: %w", err)
	}

	return cores, nil
}

func (p *playLogDb) topGames(n int) ([]tracker.GameTime, error) {
	rows, err := p.db.QueryContext(
		context.Background(),
		"select id, path, name, folder, time from game_times order by time desc limit ?",
		n,
	)
	if err != nil {
		return nil, fmt.Errorf("query top games: %w", err)
	}
	defer func() { _ = rows.Close() }()

	var games []tracker.GameTime
	for rows.Next() {
		var game tracker.GameTime
		err = rows.Scan(&game.Id, &game.Path, &game.Name, &game.Folder, &game.Time)
		if err != nil {
			return nil, fmt.Errorf("scan game time: %w", err)
		}

		games = append(games, game)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate game times: %w", err)
	}

	return games, nil
}

func (p *playLogDb) lastEvent(startAction, stopAction int) (tracker.EventAction, error) {
	var event tracker.EventAction
	err := p.db.QueryRowContext(
		context.Background(),
		"select timestamp, action, target, total_time from events "+
			"where action = ? or action = ? order by timestamp desc",
		startAction,
		stopAction,
	).Scan(&event.Timestamp, &event.Action, &event.Target, &event.TotalTime)
	if err != nil {
		return event, fmt.Errorf("query last event: %w", err)
	}
	return event, nil
}

func missingStopEvent(lastEvent *tracker.EventAction, stopAction, total int) tracker.EventAction {
	event := tracker.EventAction{
		Timestamp: lastEvent.Timestamp.Add(time.Second),
		Action:    stopAction,
		Target:    lastEvent.Target,
		TotalTime: lastEvent.TotalTime,
	}
	offset := total - lastEvent.TotalTime
	if offset > 0 {
		event.TotalTime = total
		event.Timestamp = lastEvent.Timestamp.Add(time.Second * time.Duration(offset))
	}
	return event
}

func (p *playLogDb) FixPowerLoss() (bool, error) {
	fixed := false

	lastCore, err := p.lastEvent(tracker.EventActionCoreStart, tracker.EventActionCoreStop)
	switch {
	case p.NoResults(err):
		// No core event needs repair.
	case err != nil:
		return fixed, err
	case lastCore.Action == tracker.EventActionCoreStart:
		core, coreErr := p.GetCore(lastCore.Target)
		if coreErr != nil && !p.NoResults(coreErr) {
			return fixed, coreErr
		}
		total := lastCore.TotalTime
		if coreErr == nil {
			total = core.Time
		}
		event := missingStopEvent(&lastCore, tracker.EventActionCoreStop, total)
		if addErr := p.AddEvent(&event); addErr != nil {
			return fixed, addErr
		}
		fixed = true
	}

	lastGame, err := p.lastEvent(tracker.EventActionGameStart, tracker.EventActionGameStop)
	switch {
	case p.NoResults(err):
		// No game event needs repair.
	case err != nil:
		return fixed, err
	case lastGame.Action == tracker.EventActionGameStart:
		game, gameErr := p.GetGame(lastGame.Target)
		if gameErr != nil && !p.NoResults(gameErr) {
			return fixed, gameErr
		}
		total := lastGame.TotalTime
		if gameErr == nil {
			total = game.Time
		}
		event := missingStopEvent(&lastGame, tracker.EventActionGameStop, total)
		if addErr := p.AddEvent(&event); addErr != nil {
			return fixed, addErr
		}
		fixed = true
	}

	return fixed, nil
}
