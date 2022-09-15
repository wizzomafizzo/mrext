package sqlindex

import (
	"database/sql"
	"os"
	"path/filepath"
	s "strings"

	"github.com/wizzomafizzo/mrext/pkg/config"
	"github.com/wizzomafizzo/mrext/pkg/utils"

	_ "github.com/mattn/go-sqlite3"
)

func setupDb(db *sql.DB) error {
	sqlStmt := `create table if not exists games (
		path text not null,
		system text not null,
		name text not null
	)`
	_, err := db.Exec(sqlStmt)
	if err != nil {
		return err
	}

	db.Exec("pragma synchronous = normal")
	db.Exec("pragma journal_mode = off")

	return nil
}

func GetIndexPath() string {
	if _, err := os.Stat(config.SdFolder); err == nil {
		return filepath.Join(config.SdFolder, config.IndexName+".db")
	} else {
		return config.IndexName + ".db"
	}
}

func getDb() (*sql.DB, error) {
	db, err := sql.Open("sqlite3", GetIndexPath())
	if err != nil {
		return nil, err
	}

	return db, nil
}

func Generate(files [][2]string, statusFn func(count int)) error {
	tmpDbPath := filepath.Join(os.TempDir(), config.IndexName+".db")
	if err := os.Remove(tmpDbPath); err != nil && !os.IsNotExist(err) {
		return err
	}

	db, err := sql.Open("sqlite3", tmpDbPath)
	if err != nil {
		return err
	}

	if err := setupDb(db); err != nil {
		return err
	}

	tx, err := db.Begin()
	if err != nil {
		return err
	}

	insertStmt, err := db.Prepare("insert into games (path, system, name) values (?, ?, ?)")
	if err != nil {
		return err
	}

	for i, file := range files {
		basename := filepath.Base(file[1])
		name := s.TrimSuffix(basename, filepath.Ext(basename))
		_, err := insertStmt.Exec(file[1], file[0], name)
		if err != nil {
			return err
		}
		statusFn(i)
	}

	tx.Commit()
	insertStmt.Close()
	db.Close()

	dbPath := GetIndexPath()
	if err := os.Remove(dbPath); err != nil && !os.IsNotExist(err) {
		return err
	}

	utils.MoveFile(tmpDbPath, dbPath)

	return nil
}

type Game struct {
	Path   string
	System string
	Name   string
}

func GamesInSystem(id string) ([]Game, error) {
	db, err := getDb()
	if err != nil {
		return nil, err
	}

	rows, err := db.Query("select path, system, name from games where system = ?", id)
	if err != nil {
		return nil, err
	}

	games := []Game{}
	for rows.Next() {
		var game Game
		err := rows.Scan(&game.Path, &game.System, &game.Name)
		if err != nil {
			return nil, err
		}
		games = append(games, game)
	}

	return games, nil
}

func SearchGames(name string) ([]Game, error) {
	db, err := getDb()
	if err != nil {
		return nil, err
	}

	rows, err := db.Query("select path, system, name from games where name like ?", "%"+name+"%")
	if err != nil {
		return nil, err
	}

	games := []Game{}
	for rows.Next() {
		var game Game
		err := rows.Scan(&game.Path, &game.System, &game.Name)
		if err != nil {
			return nil, err
		}
		games = append(games, game)
	}

	return games, nil
}
