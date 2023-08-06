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

func launchCard(cfg *config.UserConfig, card Card) error {
	if card.Text != "" {
		err := loadCoreFromFilename(cfg, card.Text)
		if err != nil {
			return fmt.Errorf("error loading core: %s", err)
		}
		// TODO: if string is in special format
		//       e.g. !!GBA:abad9c764c35b8202e3d9e5915ca7007bdc7cc62 try to load that way.
	} else {
		logger.Info("no text NDEF found, falling back to UID mapping in CSV file")

		db, err := loadDatabase()
		if err != nil {
			logger.Error("error loading database: %s", err)
			return err
		}

		err = loadCoreFromCardUID(cfg, db, card.UID)
		if err != nil {
			logger.Error("error loading core: %s", err)
		}
	}

	return nil
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
	database, err := loadDatabase()
	if err != nil {
		logger.Error("error loading database: %s", err)
	} else {
		logger.Info("loaded %d mappings", len(database))
	}

	filename, ok := db[cardId]
	if !ok {
		return fmt.Errorf("no core mapped for card ID: %s", cardId)
	}

	return loadCoreFromFilename(cfg, filename)
}
