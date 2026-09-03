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

package games

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/wizzomafizzo/mrext/cmd/remote/websocket"
	"github.com/wizzomafizzo/mrext/pkg/config"
	"github.com/wizzomafizzo/mrext/pkg/mister"
	"github.com/wizzomafizzo/mrext/pkg/service"
	"github.com/wizzomafizzo/mrext/pkg/tracker"
)

type fakeDb struct {
	logger *service.Logger
	cfg    *config.UserConfig
}

func (*fakeDb) FixPowerLoss() (bool, error) {
	return false, nil
}

func (f *fakeDb) AddEvent(ev *tracker.EventAction) error {
	switch ev.Action {
	case tracker.EventActionCoreStart:
		websocket.Broadcast(f.logger, "coreRunning:"+ev.Target)
		SendAnnounceGame(f.cfg, f.logger, ev)
	case tracker.EventActionCoreStop:
		websocket.Broadcast(f.logger, "coreRunning:")
		SendAnnounceGame(f.cfg, f.logger, ev)
	case tracker.EventActionGameStart:
		websocket.Broadcast(f.logger, "gameRunning:"+ev.Target)
		SendAnnounceGame(f.cfg, f.logger, ev)
	case tracker.EventActionGameStop:
		websocket.Broadcast(f.logger, "gameRunning:")
		SendAnnounceGame(f.cfg, f.logger, ev)
	case tracker.EventActionMenuNavigation:
		websocket.Broadcast(f.logger, "menuNavigation:"+ev.Target)
	}
	return nil
}

func (*fakeDb) UpdateCore(_ tracker.CoreTime) error {
	return nil
}

func (*fakeDb) GetCore(_ string) (tracker.CoreTime, error) {
	return tracker.CoreTime{}, nil
}

func (*fakeDb) UpdateGame(_ tracker.GameTime) error {
	return nil
}

func (*fakeDb) GetGame(_ string) (tracker.GameTime, error) {
	return tracker.GameTime{}, nil
}

func (*fakeDb) NoResults(_ error) bool {
	return true
}

func StartTracker(logger *service.Logger, cfg *config.UserConfig) (*tracker.Tracker, func() error, error) {
	tr, err := tracker.NewTracker(logger, cfg, &fakeDb{
		logger: logger,
		cfg:    cfg,
	})
	if err != nil {
		logger.Error("failed to start tracker: %s", err)
		return nil, nil, fmt.Errorf("create tracker: %w", err)
	}

	tr.LoadCore()
	if !mister.ActiveGameEnabled() {
		if activeErr := mister.SetActiveGame(""); activeErr != nil {
			tr.Logger.Error("error setting active game: %s", activeErr)
		}
	}

	watcher, err := tracker.StartFileWatch(tr)
	if err != nil {
		tr.Logger.Error("error starting file watch: %s", err)
		return nil, nil, fmt.Errorf("start tracker file watch: %w", err)
	}

	tr.StartTicker(0)

	return tr, func() error {
		err := watcher.Close()
		if err != nil {
			tr.Logger.Error("error closing file watcher: %s", err)
		}
		tr.StopAll()
		return nil
	}, nil
}

type PlayingPayload struct {
	Core       string `json:"core"`
	System     string `json:"system"`
	SystemName string `json:"systemName"`
	Game       string `json:"game"`
	GameName   string `json:"gameName"`
}

func HandlePlaying(tr *tracker.Tracker) http.HandlerFunc {
	return func(w http.ResponseWriter, _ *http.Request) {
		playing := PlayingPayload{
			Core:       tr.ActiveCore,
			System:     tr.ActiveSystem,
			SystemName: tr.ActiveSystemName,
			Game:       tr.ActiveGame,
			GameName:   tr.ActiveGameName,
		}

		err := json.NewEncoder(w).Encode(playing)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
}

type AnnounceGamePayload struct {
	Platform     string `json:"platform"`
	Hostname     string `json:"hostname"`
	Core         string `json:"core"`
	System       string `json:"system"`
	SystemName   string `json:"systemName"`
	GamePath     string `json:"gamePath"`
	GameFilename string `json:"gameFilename"`
	GameName     string `json:"gameName"`
}

func SendAnnounceGame(cfg *config.UserConfig, logger *service.Logger, ev *tracker.EventAction) {
	hostname, err := os.Hostname()
	if err != nil {
		hostname = ""
	}

	announce := AnnounceGamePayload{
		Platform:     "MiSTer",
		Hostname:     hostname,
		Core:         ev.ActiveCore.Core,
		System:       ev.ActiveCore.System,
		SystemName:   ev.ActiveCore.SystemName,
		GamePath:     ev.ActiveGame.Path,
		GameFilename: filepath.Base(ev.ActiveGame.Path),
		GameName:     ev.ActiveGame.Name,
	}

	url := cfg.Remote.AnnounceGameURL
	data, err := json.Marshal(announce)
	if err != nil {
		logger.Error("error marshalling announce payload: %s", err)
		return
	}

	if url != "" {
		req, requestErr := http.NewRequestWithContext(
			context.Background(), http.MethodPost, url, bytes.NewReader(data),
		)
		if requestErr != nil {
			logger.Error("error creating announce request: %s", requestErr)
			return
		}
		req.Header.Set("Content-Type", "application/json")
		client := &http.Client{Timeout: 15 * time.Second}
		resp, requestErr := client.Do(req)
		if requestErr != nil {
			logger.Error("error sending announce payload: %s", requestErr)
			return
		}
		_ = resp.Body.Close()
	}
}
