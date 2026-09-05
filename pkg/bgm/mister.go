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
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/fsnotify/fsnotify"
)

// RunCmd sends a command to MiSTer exactly like the Python script's
// os.system('echo "<cmd>" > /dev/MiSTer_cmd'), which worked around hangs seen
// when opening the command interface in-process.
func RunCmd(paths *Paths, cmd string) error {
	script := fmt.Sprintf("echo %q > %s", cmd, shellQuote(paths.CmdInterface))
	// #nosec G204 -- the command text is a fixed MiSTer volume instruction.
	if err := exec.CommandContext(context.Background(), "/bin/sh", "-c", script).Run(); err != nil {
		return fmt.Errorf("send MiSTer command %q: %w", cmd, err)
	}
	return nil
}

func shellQuote(value string) string {
	return "'" + strings.ReplaceAll(value, "'", `'\''`) + "'"
}

// VolumeSet clamps volume to MiSTer's 0-7 range and applies it.
func VolumeSet(paths *Paths, logger *Logger, volume int) {
	volume = max(0, min(7, volume))
	logger.Logf("Setting volume to %d", volume)
	if err := RunCmd(paths, fmt.Sprintf("volume %d", volume)); err != nil {
		logger.Log(err.Error())
	}
}

// GetCore reads /tmp/CORENAME. ok is false when the file is missing.
func GetCore(paths *Paths) (core string, ok bool) {
	// #nosec G304 -- CORENAME is a fixed MiSTer runtime path.
	data, err := os.ReadFile(filepath.Clean(paths.CoreNameFile))
	if err != nil {
		return "", false
	}
	return strings.TrimSpace(string(data)), true
}

// CoreWatcher blocks until MiSTer rewrites /tmp/CORENAME, replacing the
// one-shot `inotifywait -e modify` the Python script spawned.
type CoreWatcher struct {
	logger        *Logger
	paths         Paths
	RetryInterval time.Duration
	Settle        time.Duration
	RetryCount    int
}

// NewCoreWatcher uses the Python script's timings: sixteen one-second
// retries while the file is missing.
func NewCoreWatcher(paths *Paths, logger *Logger) *CoreWatcher {
	return &CoreWatcher{
		logger:        logger,
		paths:         *paths,
		RetryInterval: time.Second,
		Settle:        100 * time.Millisecond,
		RetryCount:    16,
	}
}

// WaitCoreChange returns the new core name after the next modification, or
// ok=false when the file never appears, the watch fails, or ctx ends.
func (w *CoreWatcher) WaitCoreChange(ctx context.Context) (core string, ok bool) {
	if _, exists := GetCore(&w.paths); !exists {
		w.logger.Log("CORENAME file does not exist, retrying...")
		for range w.RetryCount {
			if !sleepContext(ctx, w.RetryInterval) {
				return "", false
			}
			if _, exists := GetCore(&w.paths); exists {
				break
			}
		}
		if _, exists := GetCore(&w.paths); !exists {
			w.logger.Log("No CORENAME file found")
			return "", false
		}
	}

	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		w.logger.Log("Error when running inotify watch")
		return "", false
	}
	defer func() { _ = watcher.Close() }()
	if err := watcher.Add(w.paths.CoreNameFile); err != nil {
		w.logger.Log("Error when running inotify watch")
		return "", false
	}
	if !w.awaitModify(ctx, watcher) {
		return "", false
	}
	// MiSTer truncates then writes; give the write a moment so the file is
	// never read empty in between, as the slower Python path did in practice.
	w.drain(ctx, watcher)

	core, exists := GetCore(&w.paths)
	if !exists {
		w.logger.Log("Core change to: None")
		return "", false
	}
	w.logger.Logf("Core change to: %s", core)
	return core, true
}

func (w *CoreWatcher) awaitModify(ctx context.Context, watcher *fsnotify.Watcher) bool {
	for {
		select {
		case <-ctx.Done():
			return false
		case event, open := <-watcher.Events:
			if !open {
				w.logger.Log("Error when running inotify watch")
				return false
			}
			if event.Has(fsnotify.Write) {
				return true
			}
		case <-watcher.Errors:
			w.logger.Log("Error when running inotify watch")
			return false
		}
	}
}

func (w *CoreWatcher) drain(ctx context.Context, watcher *fsnotify.Watcher) {
	if w.Settle <= 0 {
		return
	}
	timer := time.NewTimer(w.Settle)
	defer timer.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-timer.C:
			return
		case <-watcher.Events:
		case <-watcher.Errors:
		}
	}
}

// sleepContext waits for duration unless ctx ends first.
func sleepContext(ctx context.Context, duration time.Duration) bool {
	timer := time.NewTimer(duration)
	defer timer.Stop()
	select {
	case <-ctx.Done():
		return false
	case <-timer.C:
		return true
	}
}
