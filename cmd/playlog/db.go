package main

import (
	"database/sql"
	"time"

	"github.com/wizzomafizzo/mrext/pkg/config"
	"github.com/wizzomafizzo/mrext/pkg/tracker"

	_ "github.com/mattn/go-sqlite3"
)

type playLogDb struct {
	db *sql.DB
}

func openPlayLogDb() (*playLogDb, error) {
	pldb := &playLogDb{}

	db, err := sql.Open("sqlite3", config.PlayLogDbFile)
	if err != nil {
		return nil, err
	}

	pldb.db = db
	pldb.setupDb()

	return pldb, nil
}

func (p *playLogDb) NoResults(err error) bool {
	return err == sql.ErrNoRows
}

func (p *playLogDb) setupDb() error {
	sqlEvents := `create table if not exists events (
		timestamp timestamp not null,
		action integer not null,
		target text not null,
		total_time integer not null
	)`
	_, err := p.db.Exec(sqlEvents)
	if err != nil {
		return err
	}

	sqlCoreTimes := `create table if not exists core_times (
		name integer not null unique,
		time integer not null
	)`
	_, err = p.db.Exec(sqlCoreTimes)
	if err != nil {
		return err
	}

	sqlGameTimes := `create table if not exists game_times (
		id text not null unique,
		path text not null,
		name text not null,
		folder text not null,
		time integer not null
	)`
	_, err = p.db.Exec(sqlGameTimes)
	if err != nil {
		return err
	}

	return nil
}

func (p *playLogDb) GetCore(name string) (tracker.CoreTime, error) {
	var core tracker.CoreTime

	err := p.db.QueryRow("select name, time from core_times where name = ?", name).Scan(&core.Name, &core.Time)

	if err != nil {
		return core, err
	}

	return core, nil
}

func (p *playLogDb) UpdateCore(core tracker.CoreTime) error {
	_, err := p.db.Exec("insert or replace into core_times (name, time) values (?, ?)", core.Name, core.Time)
	return err
}

func (p *playLogDb) GetGame(id string) (tracker.GameTime, error) {
	var game tracker.GameTime

	err := p.db.QueryRow(
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
		return game, err
	}

	return game, nil
}

func (p *playLogDb) UpdateGame(game tracker.GameTime) error {
	_, err := p.db.Exec(
		"insert or replace into game_times (id, path, name, folder, time) values (?, ?, ?, ?, ?)",
		game.Id,
		game.Path,
		game.Name,
		game.Folder,
		game.Time,
	)
	return err
}

func (p *playLogDb) AddEvent(event tracker.EventAction) error {
	_, err := p.db.Exec(
		"insert into events (timestamp, action, target, total_time) values (?, ?, ?, ?)",
		event.Timestamp,
		event.Action,
		event.Target,
		event.TotalTime,
	)
	return err
}

func (p *playLogDb) topCores(n int) ([]tracker.CoreTime, error) {
	rows, err := p.db.Query("select name, time from core_times order by time desc limit ?", n)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var cores []tracker.CoreTime
	for rows.Next() {
		var core tracker.CoreTime
		err = rows.Scan(&core.Name, &core.Time)
		if err != nil {
			return nil, err
		}

		cores = append(cores, core)
	}

	return cores, nil
}

func (p *playLogDb) topGames(n int) ([]tracker.GameTime, error) {
	rows, err := p.db.Query("select id, path, name, folder, time from game_times order by time desc limit ?", n)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var games []tracker.GameTime
	for rows.Next() {
		var game tracker.GameTime
		err = rows.Scan(&game.Id, &game.Path, &game.Name, &game.Folder, &game.Time)
		if err != nil {
			return nil, err
		}

		games = append(games, game)
	}

	return games, nil
}

func (p *playLogDb) FixPowerLoss() (bool, error) {
	// FIXME: repeating a lot of code here?
	var lastEvent tracker.EventAction
	fixed := false

	// cores
	err := p.db.QueryRow(
		"select timestamp, action, target, total_time from events where action = ? or action = ? order by timestamp desc",
		tracker.EventActionCoreStart,
		tracker.EventActionCoreStop,
	).Scan(
		&lastEvent.Timestamp,
		&lastEvent.Action,
		&lastEvent.Target,
		&lastEvent.TotalTime,
	)
	if p.NoResults(err) {
		// skip
	} else if err != nil {
		return fixed, err
	} else if lastEvent.Action == tracker.EventActionCoreStart {
		newEvent := tracker.EventAction{
			Timestamp: lastEvent.Timestamp.Add(time.Second),
			Action:    tracker.EventActionCoreStop,
			Target:    lastEvent.Target,
			TotalTime: lastEvent.TotalTime,
		}

		ct, err := p.GetCore(lastEvent.Target)
		if p.NoResults(err) {
			err := p.AddEvent(newEvent)
			if err != nil {
				return fixed, err
			}
			fixed = true
		} else if err != nil {
			return fixed, err
		} else {
			offset := ct.Time - lastEvent.TotalTime
			if offset > 0 {
				newEvent.TotalTime = ct.Time
				newEvent.Timestamp = lastEvent.Timestamp.Add(time.Second * time.Duration(offset))
			}
			err := p.AddEvent(newEvent)
			if err != nil {
				return fixed, err
			}
			fixed = true
		}
	}

	// games
	err = p.db.QueryRow(
		"select timestamp, action, target, total_time from events where action = ? or action = ? order by timestamp desc",
		tracker.EventActionGameStart,
		tracker.EventActionGameStop,
	).Scan(
		&lastEvent.Timestamp,
		&lastEvent.Action,
		&lastEvent.Target,
		&lastEvent.TotalTime,
	)
	if p.NoResults(err) {
		// skip
	} else if err != nil {
		return fixed, err
	} else if lastEvent.Action == tracker.EventActionGameStart {
		newEvent := tracker.EventAction{
			Timestamp: lastEvent.Timestamp.Add(time.Second),
			Action:    tracker.EventActionGameStop,
			Target:    lastEvent.Target,
			TotalTime: lastEvent.TotalTime,
		}

		gt, err := p.GetGame(lastEvent.Target)
		if p.NoResults(err) {
			err := p.AddEvent(newEvent)
			if err != nil {
				return fixed, err
			}
			fixed = true
		} else if err != nil {
			return fixed, err
		} else {
			offset := gt.Time - lastEvent.TotalTime
			if offset > 0 {
				newEvent.TotalTime = gt.Time
				newEvent.Timestamp = lastEvent.Timestamp.Add(time.Second * time.Duration(offset))
			}
			err := p.AddEvent(newEvent)
			if err != nil {
				return fixed, err
			}
			fixed = true
		}
	}

	return fixed, nil
}
