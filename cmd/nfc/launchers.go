package main

import (
	"encoding/csv"
	"errors"
	"fmt"
	"github.com/wizzomafizzo/mrext/pkg/config"
	"github.com/wizzomafizzo/mrext/pkg/mister"
	"os"
	"path/filepath"
)

func readCsvFile(filePath string) ([][]string, error) {
	f, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("unable to load fallback database file %s: %s", filePath, err)
	}
	defer func(f *os.File) {
		_ = f.Close()
	}(f)

	csvReader := csv.NewReader(f)
	records, err := csvReader.ReadAll()
	if err != nil {
		return nil, fmt.Errorf("unable to parse CSV file %s: %s", filePath, err)
	}

	return records, nil
}

func loadDatabase() (map[string]string, error) {
	database := make(map[string]string)

	data, err := readCsvFile(databaseFile)
	if err != nil {
		return nil, err
	}

	for _, row := range data {
		uid := row[0]
		value := row[1]
		database[uid] = value
	}

	return database, nil
}

func loadCoreFromFilename(cfg *config.UserConfig, filename string) error {
	// TODO: this will not work very well long term, full core filename changes each release
	//		 but it's ok, no problem using partial matches as mister does
	fullPath := filepath.Join(config.SdFolder, filename) // TODO: saves a few chars on the tag but is it worth it?

	if _, err := os.Stat(fullPath); errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("core does not exist: %s", fullPath)
	}

	return mister.LaunchGenericFile(cfg, fullPath)
}

func loadCoreFromCardUID(cfg *config.UserConfig, db map[string]string, cardId string) error {
	filename, ok := db[cardId]
	if !ok {
		return fmt.Errorf("no core mapped for card ID: %s", cardId)
	}

	return loadCoreFromFilename(cfg, filename)
}
