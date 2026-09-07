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

package bgm

import (
	"context"
	"fmt"
	"os"
	"strings"
	"sync"
	"time"
)

// Service is the long-running "exec" process: boot sound, remote socket and
// the core-change loop that starts and stops music.
type Service struct {
	logger      *Logger
	player      *Player
	remote      *Remote
	watcher     *CoreWatcher
	shutdown    chan struct{}
	paths       Paths
	cleanupOnce sync.Once
}

// NewService builds the player from the current configuration.
func NewService(paths *Paths, logger *Logger) *Service {
	cfg, _ := LoadConfig(paths)
	return &Service{
		paths:    *paths,
		logger:   logger,
		player:   NewPlayer(paths, logger, &cfg),
		watcher:  NewCoreWatcher(paths, logger),
		shutdown: make(chan struct{}),
	}
}

// Player exposes the service player for tests.
func (s *Service) Player() *Player { return s.player }

// Watcher exposes the core watcher so tests can shorten its timings.
func (s *Service) Watcher() *CoreWatcher { return s.watcher }

// Run mirrors start_service(). It returns when CORENAME disappears, the
// watch fails, or ctx ends.
func (s *Service) Run(ctx context.Context) error {
	s.logger.Log("Starting service...")
	s.logger.Logf("Playlist folder: %s", s.player.PlaylistPath(s.player.Playlist()))

	cfg, _ := LoadConfig(&s.paths)

	remote, err := StartRemote(&s.paths, s.logger, s.player)
	if err != nil {
		return err
	}
	s.remote = remote

	if err := s.waitBootDelay(ctx, cfg.BootDelay); err != nil {
		return err
	}
	if cfg.ShouldChangeVolume() {
		VolumeSet(&s.paths, s.logger, cfg.MenuVolume)
		s.player.PlayBoot()
		VolumeSet(&s.paths, s.logger, cfg.DefaultVolume)
	} else {
		s.player.PlayBoot()
	}

	core, hasCore := GetCore(&s.paths)
	if core == MenuCore || !hasCore || s.player.PlayInCore() {
		if cfg.ShouldChangeVolume() && core == MenuCore {
			VolumeSet(&s.paths, s.logger, cfg.MenuVolume)
		}
		s.player.StartPlaylist(cfg.Playback)
	}

	for {
		newCore, ok := s.watcher.WaitCoreChange(ctx)
		if err := ctx.Err(); err != nil {
			return fmt.Errorf("BGM service stopped: %w", err)
		}
		cfg, _ = LoadConfig(&s.paths)
		s.player.SetPlayInCore(cfg.PlayInCore)

		if !ok {
			s.logger.Log("CORENAME file is missing, exiting...")
			return nil
		}

		switch {
		case hasCore && strings.EqualFold(core, newCore):
			s.logger.Log("CORENAME file changed, but core is the same")
		case s.player.PlayInCore():
			s.logger.Log("playincore is enabled")
		case newCore == MenuCore:
			s.enterMenu(&cfg)
		default:
			s.enterCore(&cfg, newCore)
		}

		core, hasCore = newCore, true
	}
}

func (s *Service) waitBootDelay(ctx context.Context, seconds float64) error {
	delay, err := bootDelayDuration(seconds)
	if err != nil {
		return err
	}
	if delay > 0 {
		s.logger.Logf("Waiting %s before startup audio", delay)
	}
	timer := time.NewTimer(delay)
	defer timer.Stop()
	select {
	case <-ctx.Done():
		return fmt.Errorf("BGM startup stopped: %w", ctx.Err())
	case <-s.shutdown:
		return fmt.Errorf("BGM startup stopped: %w", context.Canceled)
	case <-timer.C:
	}
	// Cancellation takes precedence when both timer and shutdown are ready.
	select {
	case <-s.shutdown:
		return fmt.Errorf("BGM startup stopped: %w", context.Canceled)
	default:
	}
	if err := ctx.Err(); err != nil {
		return fmt.Errorf("BGM startup stopped: %w", err)
	}
	return nil
}

func (s *Service) enterMenu(cfg *Config) {
	s.logger.Log("Switched to menu core, starting playlist...")
	s.logger.Log("Grabbing mutex")
	s.player.CmdMu.Lock()
	defer func() {
		s.logger.Log("Releasing mutex")
		s.player.CmdMu.Unlock()
	}()
	if cfg.ShouldChangeVolume() {
		s.logger.Log("Changing volume to menu volume")
		VolumeSet(&s.paths, s.logger, cfg.MenuVolume)
	}
	s.logger.Log("Starting playlist")
	s.player.StartCurrentPlaylist()
}

func (s *Service) enterCore(cfg *Config, core string) {
	s.logger.Log("Exited menu core, stopping playlist...")
	s.logger.Log("Grabbing mutex")
	s.player.CmdMu.Lock()
	defer func() {
		s.logger.Log("Releasing mutex")
		s.player.CmdMu.Unlock()
	}()
	s.logger.Log("Stopping playlist")
	s.player.StopPlaylist()
	s.logger.Log("Playing core boot")
	s.player.PlayCoreBoot(core)
	if cfg.ShouldChangeVolume() {
		s.logger.Log("Changing volume to default volume")
		VolumeSet(&s.paths, s.logger, cfg.DefaultVolume)
	}
}

// Cleanup mirrors cleanup(): stop music, stop the remote, remove the socket.
// It is safe to call more than once and from a signal handler.
func (s *Service) Cleanup() {
	s.cleanupOnce.Do(func() {
		close(s.shutdown)
		s.player.StopPlaylist()
		if s.remote != nil {
			s.remote.Close()
		}
		if SocketExists(&s.paths) {
			_ = os.Remove(s.paths.SocketFile)
		}
	})
}
