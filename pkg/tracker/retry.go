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

package tracker

import (
	"errors"
	"fmt"
	"io/fs"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/wizzomafizzo/mrext/pkg/config"
)

// StartFileWatchWithRetry tolerates MiSTer state files appearing late in boot.
// Each failed StartFileWatch closes its watcher before the next attempt.
func StartFileWatchWithRetry(tr *Tracker) (*fsnotify.Watcher, error) {
	return retryFileWatch(func() (*fsnotify.Watcher, error) { return StartFileWatch(tr) },
		time.Sleep, tr.Logger.Warn)
}

func retryFileWatch(
	start func() (*fsnotify.Watcher, error), pause func(time.Duration), warn func(string, ...any),
) (*fsnotify.Watcher, error) {
	for attempt := 1; ; attempt++ {
		watcher, err := start()
		if err == nil {
			return watcher, nil
		}
		if !errors.Is(err, fs.ErrNotExist) || attempt >= config.RemoteWatchAttempts {
			return nil, fmt.Errorf("tracker watch setup failed after %d attempt(s): %w", attempt, err)
		}
		warn("tracker files not ready (attempt %d/%d), retrying: %s", attempt, config.RemoteWatchAttempts, err)
		pause(time.Duration(config.RemoteWatchRetrySeconds) * time.Second)
	}
}
