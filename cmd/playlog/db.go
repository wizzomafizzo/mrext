package main

import (
	"database/sql"
	"time"

	"github.com/wizzomafizzo/mrext/pkg/config"

	_ "github.com/mattn/go-sqlite3"
)

func noResults(err error) bool {
	return err == sql.ErrNoRows
}

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

func (p *playLogDb) getCore(name string) (coreTime, error) {
	var core coreTime

	err := p.db.QueryRow("select name, time from core_times where name = ?", name).Scan(&core.name, &core.time)

	if err != nil {
		return core, err
	}

	return core, nil
}

func (p *playLogDb) updateCore(core coreTime) error {
	_, err := p.db.Exec("insert or replace into core_times (name, time) values (?, ?)", core.name, core.time)
	return err
}

func (p *playLogDb) getGame(id string) (gameTime, error) {
	var game gameTime

	err := p.db.QueryRow("select id, path, name, folder, time from game_times where id = ?", id).Scan(&game.id, &game.path, &game.name, &game.folder, &game.time)

	if err != nil {
		return game, err
	}

	return game, nil
}

func (p *playLogDb) updateGame(game gameTime) error {
	_, err := p.db.Exec("insert or replace into game_times (id, path, name, folder, time) values (?, ?, ?, ?, ?)", game.id, game.path, game.name, game.folder, game.time)
	return err
}

func (p *playLogDb) addEvent(event eventAction) error {
	_, err := p.db.Exec("insert into events (timestamp, action, target, total_time) values (?, ?, ?, ?)", event.timestamp, event.action, event.target, event.totalTime)
	return err
}

func (p *playLogDb) fixPowerLoss() (bool, error) {
	// FIXME: repeating a lot of code here?
	var lastEvent eventAction
	fixed := false

	// cores
	err := p.db.QueryRow("select timestamp, action, target, total_time from events where action = ? or action = ? order by timestamp desc", eventActionCoreStart, eventActionCoreStop).Scan(&lastEvent.timestamp, &lastEvent.action, &lastEvent.target, &lastEvent.totalTime)
	if noResults(err) {
		// skip
	} else if err != nil {
		return fixed, err
	} else if lastEvent.action == eventActionCoreStart {
		newEvent := eventAction{
			timestamp: lastEvent.timestamp.Add(time.Second),
			action:    eventActionCoreStop,
			target:    lastEvent.target,
			totalTime: lastEvent.totalTime,
		}

		ct, err := p.getCore(lastEvent.target)
		if noResults(err) {
			err := p.addEvent(newEvent)
			if err != nil {
				return fixed, err
			}
			fixed = true
		} else if err != nil {
			return fixed, err
		} else {
			offset := ct.time - lastEvent.totalTime
			if offset > 0 {
				newEvent.totalTime = ct.time
				newEvent.timestamp = lastEvent.timestamp.Add(time.Second * time.Duration(offset))
			}
			err := p.addEvent(newEvent)
			if err != nil {
				return fixed, err
			}
			fixed = true
		}
	}

	// games
	err = p.db.QueryRow("select timestamp, action, target, total_time from events where action = ? or action = ? order by timestamp desc", eventActionGameStart, eventActionGameStop).Scan(&lastEvent.timestamp, &lastEvent.action, &lastEvent.target, &lastEvent.totalTime)
	if noResults(err) {
		// skip
	} else if err != nil {
		return fixed, err
	} else if lastEvent.action == eventActionGameStart {
		newEvent := eventAction{
			timestamp: lastEvent.timestamp.Add(time.Second),
			action:    eventActionGameStop,
			target:    lastEvent.target,
			totalTime: lastEvent.totalTime,
		}

		gt, err := p.getGame(lastEvent.target)
		if noResults(err) {
			err := p.addEvent(newEvent)
			if err != nil {
				return fixed, err
			}
			fixed = true
		} else if err != nil {
			return fixed, err
		} else {
			offset := gt.time - lastEvent.totalTime
			if offset > 0 {
				newEvent.totalTime = gt.time
				newEvent.timestamp = lastEvent.timestamp.Add(time.Second * time.Duration(offset))
			}
			err := p.addEvent(newEvent)
			if err != nil {
				return fixed, err
			}
			fixed = true
		}
	}

	return fixed, nil
}
