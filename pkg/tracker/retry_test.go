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
	"testing"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/wizzomafizzo/mrext/pkg/config"
)

func TestWatchRetriesMissingPathsUntilReady(t *testing.T) {
	calls, waits, logs := 0, 0, 0
	want := &fsnotify.Watcher{}
	got, err := retryFileWatch(func() (*fsnotify.Watcher, error) {
		calls++
		if calls < 3 {
			return nil, fmt.Errorf("watch core name: %w", fs.ErrNotExist)
		}
		return want, nil
	}, func(delay time.Duration) {
		waits++
		if delay != time.Duration(config.RemoteWatchRetrySeconds)*time.Second {
			t.Fatalf("unexpected retry delay: %v", delay)
		}
	}, func(string, ...any) { logs++ })
	if err != nil || got != want || calls != 3 || waits != 2 || logs != 2 {
		t.Fatalf("result=%p err=%v calls=%d waits=%d logs=%d", got, err, calls, waits, logs)
	}
}

func TestWatchRetryIsBoundedAndOnlyForMissingPaths(t *testing.T) {
	for _, failure := range []error{fs.ErrNotExist, fs.ErrPermission} {
		calls, waits := 0, 0
		_, err := retryFileWatch(func() (*fsnotify.Watcher, error) {
			calls++
			return nil, fmt.Errorf("watch setup: %w", failure)
		}, func(time.Duration) { waits++ }, func(string, ...any) {})
		wantCalls := 1
		if errors.Is(failure, fs.ErrNotExist) {
			wantCalls = config.RemoteWatchAttempts
		}
		if !errors.Is(err, failure) || calls != wantCalls || waits != wantCalls-1 {
			t.Fatalf("failure=%v calls=%d waits=%d err=%v", failure, calls, waits, err)
		}
	}
}
