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
	"sync/atomic"
	"time"

	"github.com/wizzomafizzo/mrext/pkg/config"
	"github.com/wizzomafizzo/mrext/pkg/mister"
	"github.com/wizzomafizzo/mrext/pkg/service"
)

// Status is the hardcore part of retroachievements.cfg, and the mode the
// running core started in.
type Status struct {
	// Session is the mode the running core started in: "hardcore",
	// "softcore", or "unknown" when Remote cannot tell.
	Session string `json:"session"`
	// Present is false without a retroachievements.cfg, that is without the
	// RetroAchievements build or its login.
	Present bool `json:"present"`
	// Hardcore is the mode each core load starts in.
	Hardcore bool `json:"hardcore"`
	// Forced is force_hardcore, which turns hardcore on for every core.
	Forced bool `json:"forced"`
}

const (
	sessionHardcore = "hardcore"
	sessionSoftcore = "softcore"
	sessionUnknown  = "unknown"
)

// fileMargin is how much older than the running core retroachievements.cfg
// must be for its value to count as the one Main read.
const fileMargin = 2 * time.Second

// coreRewrite is how long after a launch's core first starts a new CORENAME
// time still belongs to that core.
const coreRewrite = time.Second

// launchRecord is a launch that asked for a mode, and when its core started.
type launchRecord struct {
	before  time.Time
	started time.Time
	// until is CORENAME's time when a launch without a mode followed this
	// one; zero until one does. A core after it is the later launch's.
	until    time.Time
	hardcore bool
}

// owns reports whether the core that started at core is this launch's: the
// first one seen after it, or that one written again. A RetroAchievements
// build rewrites CORENAME about 40 ms after the first write (SuperStation,
// RA_SNES), so the time first seen may not be the one that stays.
func (r *launchRecord) owns(core time.Time) bool {
	return !r.started.IsZero() && !core.Before(r.started) && core.Sub(r.started) <= coreRewrite &&
		r.ran(core)
}

// ran reports whether a core that started at core can be this launch's:
// one started after a later launch without a mode is that launch's.
func (r *launchRecord) ran(core time.Time) bool {
	return r.until.IsZero() || !core.After(r.until)
}

// adopt takes core as the launch's own when none is yet and core started
// after the launch, and before any later launch without a mode.
func (r *launchRecord) adopt(core time.Time) {
	if r.started.IsZero() && !core.IsZero() && !core.Equal(r.before) && r.ran(core) {
		r.started = core
	}
}

// Main writes CORENAME as a new core starts, about 0.1 s before it reads
// retroachievements.cfg (measured on a SuperStation, MENU, SNES and PSX).
// The default goes back settleDelay after that, or startTimeout after the
// launch if CORENAME never changes.
const (
	pollInterval = 250 * time.Millisecond
	settleDelay  = 3 * time.Second
	startTimeout = 30 * time.Second
)

// Modes owns Remote's use of retroachievements.cfg.
type Modes struct {
	logger *service.Logger
	// restore is the default to put back while a launch in the other mode
	// waits for its core to start.
	restore *bool
	// last is the latest launch that asked for a mode.
	last     *launchRecord
	path     string
	coreName string

	pollInterval time.Duration
	settleDelay  time.Duration
	startTimeout time.Duration

	// generation tells a waiting restore that a later launch took over.
	generation uint64
	// mu also covers each launch, so no launch can write a mode between
	// another launch putting the default back and starting its game.
	mu sync.Mutex
	// pending mirrors restore != nil, so a waiting restore can poll it.
	pending atomic.Bool
	// closed is set as Remote stops: no launch writes a mode after that.
	closed bool
}

// New returns Modes for the MiSTer's retroachievements.cfg.
func New(logger *service.Logger) *Modes {
	return &Modes{
		logger:       logger,
		path:         config.RetroAchievementsFile,
		coreName:     config.CoreNameFile,
		pollInterval: pollInterval,
		settleDelay:  settleDelay,
		startTimeout: startTimeout,
	}
}

// status reads the file, showing the default rather than a launch's mode
// while that launch waits for its core.
func (m *Modes) status() (Status, error) {
	settings, err := mister.ReadRAHardcore(m.path)
	if errors.Is(err, os.ErrNotExist) {
		return Status{Session: sessionUnknown}, nil
	}
	if err != nil {
		return Status{}, fmt.Errorf("RetroAchievements status: %w", err)
	}
	status := Status{
		Present:  true,
		Hardcore: settings.Hardcore,
		Forced:   settings.Forced,
		Session:  m.sessionLocked(settings),
	}
	if m.restore != nil {
		status.Hardcore = *m.restore
	}
	return status, nil
}

// sessionLocked works out the mode the running core started in, only when
// asked, so core changes cost nothing. It is the mode a launch asked for
// when the running core is that launch's; otherwise the file's value when
// the file is older than the core; otherwise unknown. With no clock set at
// boot, CORENAME can be dated 1970 and the answer is unknown, never a guess.
func (m *Modes) sessionLocked(file mister.RAHardcore) string {
	if file.Forced {
		return sessionHardcore
	}
	core := m.coreStarted()
	if core.IsZero() {
		return sessionUnknown
	}
	if last := m.last; last != nil {
		last.adopt(core)
		if last.owns(core) {
			return sessionName(last.hardcore)
		}
	}
	if m.restore != nil {
		// The file holds a launch's mode, not the running core's.
		return sessionUnknown
	}
	info, err := os.Stat(m.path)
	if err != nil || !info.ModTime().Add(fileMargin).Before(core) {
		return sessionUnknown
	}
	return sessionName(file.Hardcore)
}

func sessionName(hardcore bool) string {
	if hardcore {
		return sessionHardcore
	}
	return sessionSoftcore
}

// coreStarted is when the running core started, from CORENAME's time.
func (m *Modes) coreStarted() time.Time {
	info, err := os.Stat(m.coreName)
	if err != nil {
		return time.Time{}
	}
	return info.ModTime()
}

// Launch runs launch so that the new core starts in the asked mode: the
// mode is written first and the default put back once the core is up. A nil
// mode, or no retroachievements.cfg, is a plain launch.
func (m *Modes) Launch(hardcore *bool, launch func() error) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if hardcore == nil {
		m.plainLocked()
		return launch()
	}
	if m.closed {
		return errors.New("set RetroAchievements mode: Remote is stopping")
	}
	if err := m.restoreLocked(); err != nil {
		return fmt.Errorf("set RetroAchievements mode: %w", err)
	}

	settings, err := mister.ReadRAHardcore(m.path)
	if errors.Is(err, os.ErrNotExist) {
		return launch()
	}
	if err != nil {
		return fmt.Errorf("set RetroAchievements mode: %w", err)
	}

	record := &launchRecord{before: m.coreStarted(), hardcore: *hardcore}
	if settings.Hardcore != *hardcore {
		if _, err := mister.SetRAHardcore(m.path, *hardcore); err != nil {
			return fmt.Errorf("set RetroAchievements mode: %w", err)
		}
		previous := settings.Hardcore
		m.restore = &previous
		m.pending.Store(true)
		m.generation++
		go m.restoreAfterStart(m.generation, record)
		m.logger.Info("retroachievements: launching with hardcore %t", *hardcore)
	}

	if err := launch(); err != nil {
		if restoreErr := m.restoreLocked(); restoreErr != nil {
			m.logger.Error("retroachievements: putting the default back: %s", restoreErr)
		}
		return err
	}
	m.last = record
	return nil
}

// plainLocked readies a launch that asks for no mode: a mode still waiting
// for its core is put back, so the launch starts in the default, and the
// launch before keeps only the core running now, so the new core is never
// taken for it. A launch that fails leaves that core, and its mode, running.
func (m *Modes) plainLocked() {
	if err := m.restoreLocked(); err != nil {
		m.logger.Error("retroachievements: putting the default back: %s", err)
	}
	if last := m.last; last != nil && last.until.IsZero() {
		last.until = m.coreStarted()
	}
}

// Settled readies a launch without a mode before next, and runs next under
// the same lock, for routes that launch without a mode.
func (m *Modes) Settled(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		m.mu.Lock()
		defer m.mu.Unlock()
		m.plainLocked()
		next(w, r)
	}
}

// Close puts the default back now if a launch in the other mode is still
// waiting, and refuses mode launches from then on. Remote calls it as it
// stops.
func (m *Modes) Close() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.closed = true
	if err := m.restoreLocked(); err != nil {
		m.logger.Error("retroachievements: putting the default back: %s", err)
	}
}

// restoreLocked puts the default back. On failure it is kept, so the next
// launch, Close or waiting restore tries again.
func (m *Modes) restoreLocked() error {
	if m.restore == nil {
		return nil
	}
	if _, err := mister.SetRAHardcore(m.path, *m.restore); err != nil {
		return fmt.Errorf("put the default back: %w", err)
	}
	m.restore = nil
	m.pending.Store(false)
	return nil
}

// restoreAfterStart waits for the launched core to start, then puts the
// default back, unless a later launch has taken over by then.
func (m *Modes) restoreAfterStart(generation uint64, record *launchRecord) {
	var started time.Time
	deadline := time.Now().Add(m.startTimeout)
	for time.Now().Before(deadline) {
		time.Sleep(m.pollInterval)
		if !m.pending.Load() {
			return // put back already, by a later launch
		}
		if core := m.coreStarted(); !core.Equal(record.before) {
			started = core
			time.Sleep(m.settleDelay)
			break
		}
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	record.adopt(started)
	if started.IsZero() && m.last == record {
		// The core never started: a later core is not this launch's.
		m.last = nil
	}
	if m.generation == generation {
		if err := m.restoreLocked(); err != nil {
			m.logger.Error("retroachievements: putting the default back: %s", err)
		}
	}
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
		if m.restore != nil {
			// A launch in the other mode is waiting for its core: the new
			// default goes in when that launch's mode comes out.
			value := *args.Hardcore
			m.restore = &value
		}
		changed, err := m.setDefaultLocked(*args.Hardcore)
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

// setDefaultLocked writes the default, unless a launch's mode is in the
// file for now.
func (m *Modes) setDefaultLocked(hardcore bool) (bool, error) {
	if m.restore != nil {
		if _, err := os.Stat(m.path); err != nil {
			return false, fmt.Errorf("stat RetroAchievements settings: %w", err)
		}
		return false, nil
	}
	changed, err := mister.SetRAHardcore(m.path, hardcore)
	if err != nil {
		return false, fmt.Errorf("set RetroAchievements default: %w", err)
	}
	return changed, nil
}

func writeJSON(w http.ResponseWriter, logger *service.Logger, value any) {
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(value); err != nil {
		logger.Error("retroachievements: encoding response: %s", err)
	}
}
