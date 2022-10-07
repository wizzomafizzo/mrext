package main

import (
	"database/sql"

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
		timestamp integer not null,
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
