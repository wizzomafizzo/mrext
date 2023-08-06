package main

import (
	"errors"
	"fmt"
	"github.com/gocarina/gocsv"
	"github.com/wizzomafizzo/mrext/pkg/config"
	"github.com/wizzomafizzo/mrext/pkg/mister"
	"io"
	"os"
)

type NfcMappingEntry struct {
	UID      string `csv:"uid"`
	Text     string `csv:"text"`
	Override string `csv:"override"`
}

func loadDatabase() (map[string]string, map[string]string, error) {
	uids := make(map[string]string)
	texts := make(map[string]string)

	if _, err := os.Stat(databaseFile); errors.Is(err, os.ErrNotExist) {
		logger.Info("no database file found, skipping")
		return uids, texts, nil
	}

	f, err := os.Open(databaseFile)
	if err != nil {
		return uids, texts, err
	}
	defer func(c io.Closer) {
		_ = c.Close()
	}(f)

	entries := make([]NfcMappingEntry, 0)
	err = gocsv.Unmarshal(f, &entries)
	if err != nil {
		return uids, texts, err
	}

	count := 0
	for i, entry := range entries {
		if entry.UID == "" && entry.Text == "" {
			logger.Warn("entry %d has no UID or text, skipping", i+1)
			continue
		}

		if entry.UID != "" {
			uids[entry.UID] = entry.Override
		}

		if entry.Text != "" {
			texts[entry.Text] = entry.Override
		}

		count++
	}
	logger.Info("loaded %d entries from database", count)

	return uids, texts, nil
}

func launchCard(cfg *config.UserConfig, card Card) error {
	uidMap, textMap, err := loadDatabase()
	if err != nil {
		return err
	}

	if override, ok := uidMap[card.UID]; ok {
		logger.Info("launching with uid match override: %s", override)
		return mister.LaunchToken(cfg, true, override)
	}

	if override, ok := textMap[card.Text]; ok {
		logger.Info("launching with text match override: %s", override)
		return mister.LaunchToken(cfg, true, override)
	}

	if card.Text == "" {
		return fmt.Errorf("no text NDEF found in card or database")
	}

	err = mister.LaunchToken(cfg, false, card.Text)
	if err != nil {
		return err
	}

	return nil
}
