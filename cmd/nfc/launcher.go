package main

import (
	_ "embed"
	"errors"
	"fmt"
	"github.com/gocarina/gocsv"
	"github.com/wizzomafizzo/mrext/pkg/config"
	"github.com/wizzomafizzo/mrext/pkg/mister"
	"io"
	"os"
	"strings"
)

// Breviceps (https://freesound.org/people/Breviceps/sounds/445978/)
// Licence: CC0 1.0 Universal (CC0 1.0) Public Domain Dedication
//
//go:embed sounds/success.wav
var successSound []byte

// PaulMorek (https://freesound.org/people/PaulMorek/sounds/330046/)
// Licence: CC0 1.0 Universal (CC0 1.0) Public Domain Dedication
//
//go:embed sounds/fail.wav
var failSound []byte

type NfcMappingEntry struct {
	MatchUID  string `csv:"match_uid"`
	MatchText string `csv:"match_text"`
	Text      string `csv:"text"`
}

func loadDatabase(state *ServiceState) error {
	uids := make(map[string]string)
	texts := make(map[string]string)

	if _, err := os.Stat(config.NfcDatabaseFile); errors.Is(err, os.ErrNotExist) {
		logger.Info("no database file found, skipping")
		return nil
	}

	f, err := os.Open(config.NfcDatabaseFile)
	if err != nil {
		return err
	}
	defer func(c io.Closer) {
		_ = c.Close()
	}(f)

	entries := make([]NfcMappingEntry, 0)
	err = gocsv.Unmarshal(f, &entries)
	if err != nil {
		return err
	}

	count := 0
	for i, entry := range entries {
		if entry.MatchUID == "" && entry.MatchText == "" {
			logger.Warn("entry %d has no UID or text, skipping", i+1)
			continue
		}

		if entry.MatchUID != "" {
			uid := strings.TrimSpace(entry.MatchUID)
			uid = strings.ToLower(uid)
			uid = strings.ReplaceAll(uid, ":", "")
			uids[uid] = strings.TrimSpace(entry.Text)
		}

		if entry.MatchText != "" {
			text := strings.TrimSpace(entry.MatchText)
			texts[text] = strings.TrimSpace(entry.Text)
		}

		count++
	}
	logger.Info("loaded %d entries from database", count)

	state.SetDB(uids, texts)

	return nil
}

func launchCard(cfg *config.UserConfig, state *ServiceState) error {
	card := state.GetActiveCard()
	uidMap, textMap := state.GetDB()

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

	logger.Info("launching with text: %s", card.Text)
	err := mister.LaunchToken(cfg, cfg.Nfc.AllowCommands, card.Text)
	if err != nil {
		return err
	}

	return nil
}
