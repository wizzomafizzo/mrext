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

// Package retroachievements serves the hardcore setting of MiSTer's
// RetroAchievements build. Only the hardcore keys of retroachievements.cfg
// are ever read back or written: the same file holds the player's login.
package retroachievements

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"sync"

	"github.com/wizzomafizzo/mrext/pkg/config"
	"github.com/wizzomafizzo/mrext/pkg/mister"
	"github.com/wizzomafizzo/mrext/pkg/service"
)

// Status is the hardcore part of retroachievements.cfg.
type Status struct {
	// Present is false without a retroachievements.cfg, that is without the
	// RetroAchievements build or its login.
	Present bool `json:"present"`
	// Hardcore is the mode each core load starts in.
	Hardcore bool `json:"hardcore"`
	// Forced is force_hardcore, which turns hardcore on for every core.
	Forced bool `json:"forced"`
}

// Modes owns Remote's use of retroachievements.cfg.
type Modes struct {
	logger *service.Logger
	path   string
	mu     sync.Mutex
}

// New returns Modes for the MiSTer's retroachievements.cfg.
func New(logger *service.Logger) *Modes {
	return &Modes{logger: logger, path: config.RetroAchievementsFile}
}

func (m *Modes) status() (Status, error) {
	settings, err := mister.ReadRAHardcore(m.path)
	if errors.Is(err, os.ErrNotExist) {
		return Status{}, nil
	}
	if err != nil {
		return Status{}, fmt.Errorf("RetroAchievements status: %w", err)
	}
	return Status{Present: true, Hardcore: settings.Hardcore, Forced: settings.Forced}, nil
}

// HandleStatus answers GET /settings/retroachievements.
func (m *Modes) HandleStatus() http.HandlerFunc {
	return func(w http.ResponseWriter, _ *http.Request) {
		m.mu.Lock()
		status, err := m.status()
		m.mu.Unlock()
		if err != nil {
			http.Error(w, "cannot read RetroAchievements settings", http.StatusInternalServerError)
			m.logger.Error("retroachievements: %s", err)
			return
		}
		writeJSON(w, m.logger, status)
	}
}

// HandleSetHardcore answers PUT /settings/retroachievements: it sets the
// mode each core load starts in. The file is never created.
func (m *Modes) HandleSetHardcore() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var args struct {
			Hardcore *bool `json:"hardcore"`
		}
		if err := json.NewDecoder(r.Body).Decode(&args); err != nil || args.Hardcore == nil {
			http.Error(w, "hardcore must be true or false", http.StatusBadRequest)
			return
		}

		m.mu.Lock()
		defer m.mu.Unlock()
		changed, err := mister.SetRAHardcore(m.path, *args.Hardcore)
		switch {
		case errors.Is(err, os.ErrNotExist):
			http.Error(w, "retroachievements.cfg not found", http.StatusNotFound)
			return
		case err != nil:
			http.Error(w, "cannot save RetroAchievements settings", http.StatusInternalServerError)
			m.logger.Error("retroachievements: %s", err)
			return
		}
		if changed {
			m.logger.Info("retroachievements: hardcore set to %t", *args.Hardcore)
		}

		status, err := m.status()
		if err != nil {
			http.Error(w, "cannot read RetroAchievements settings", http.StatusInternalServerError)
			m.logger.Error("retroachievements: %s", err)
			return
		}
		writeJSON(w, m.logger, status)
	}
}

func writeJSON(w http.ResponseWriter, logger *service.Logger, value any) {
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(value); err != nil {
		logger.Error("retroachievements: encoding response: %s", err)
	}
}
