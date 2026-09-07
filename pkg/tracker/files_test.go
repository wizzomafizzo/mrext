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

//nolint:gosec // Tests only operate on temporary fixture paths.
package tracker

import (
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"github.com/fsnotify/fsnotify"
)

type recordingLogger struct {
	messages []string
	mu       sync.Mutex
}

func (l *recordingLogger) record(format string, _ ...any) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.messages = append(l.messages, format)
}

func (l *recordingLogger) Info(format string, v ...any)  { l.record(format, v...) }
func (l *recordingLogger) Warn(format string, v ...any)  { l.record(format, v...) }
func (l *recordingLogger) Error(format string, v ...any) { l.record(format, v...) }

// waitFor polls until condition holds, so the tests do not depend on inotify
// delivery timing.
func waitFor(t *testing.T, what string, condition func() bool) {
	t.Helper()
	deadline := time.Now().Add(5 * time.Second)
	for time.Now().Before(deadline) {
		if condition() {
			return
		}
		time.Sleep(5 * time.Millisecond)
	}
	t.Fatalf("timed out waiting for %s", what)
}

func newWatchFixture(t *testing.T) (*fileWatch, string, *int32Counter) {
	t.Helper()
	dir := t.TempDir()
	tracked := filepath.Join(dir, "CORENAME")
	if err := os.WriteFile(tracked, []byte("MENU"), 0o600); err != nil {
		t.Fatal(err)
	}

	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = watcher.Close() })
	if err := watcher.Add(tracked); err != nil {
		t.Fatal(err)
	}

	hits := &int32Counter{}
	watch := &fileWatch{
		watcher:  watcher,
		logger:   &recordingLogger{},
		files:    map[string]func(){tracked: hits.inc},
		attempts: 20,
		delay:    10 * time.Millisecond,
	}
	go watch.run()
	return watch, tracked, hits
}

type int32Counter struct {
	count int
	mu    sync.Mutex
}

func (c *int32Counter) inc() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.count++
}

func (c *int32Counter) value() int {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.count
}

func TestFileWatchReactsToWrites(t *testing.T) {
	_, tracked, hits := newWatchFixture(t)

	if err := os.WriteFile(tracked, []byte("NES"), 0o600); err != nil {
		t.Fatal(err)
	}
	waitFor(t, "a write to be handled", func() bool { return hits.value() > 0 })
}

func TestFileWatchSurvivesTheFileBeingReplaced(t *testing.T) {
	_, tracked, hits := newWatchFixture(t)

	// inotify watches follow the inode. Another tool replacing the file rather
	// than truncating it used to leave the watch on a dead inode, and the
	// tracker went silent until the service was restarted.
	if err := os.Remove(tracked); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(tracked, []byte("SNES"), 0o600); err != nil {
		t.Fatal(err)
	}
	waitFor(t, "the replaced file to be picked up", func() bool { return hits.value() > 0 })

	before := hits.value()
	if err := os.WriteFile(tracked, []byte("GBA"), 0o600); err != nil {
		t.Fatal(err)
	}
	waitFor(t, "a write after replacement to be handled", func() bool { return hits.value() > before })
}
