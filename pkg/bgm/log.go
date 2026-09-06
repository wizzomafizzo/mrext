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
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// Logger mirrors the Python log() helper: the debug flag is re-read from
// bgm.ini on every call so it can be toggled while the service runs, debug
// output goes to stdout and /tmp/bgm.log, and Print() always reaches stdout.
type Logger struct {
	out   io.Writer
	paths Paths
	mu    sync.Mutex
}

// NewLogger writes console output to stdout.
func NewLogger(paths *Paths) *Logger {
	return &Logger{out: os.Stdout, paths: *paths}
}

// NewLoggerTo writes console output to out; used by tests.
func NewLoggerTo(paths *Paths, out io.Writer) *Logger {
	return &Logger{out: out, paths: *paths}
}

// Log records a debug message.
func (l *Logger) Log(msg string) { l.write(msg, false, false) }

// Logf records a formatted debug message.
func (l *Logger) Logf(format string, args ...any) {
	l.write(fmt.Sprintf(format, args...), false, false)
}

// Print records a message that is always shown to the user.
func (l *Logger) Print(msg string) { l.write(msg, true, false) }

// Error records actionable failures even when daemon stdout is discarded.
func (l *Logger) Error(msg string) { l.write(msg, true, true) }

func (l *Logger) write(msg string, always, persist bool) {
	cfg, _ := LoadConfig(&l.paths)
	if msg == "" {
		return
	}
	l.mu.Lock()
	defer l.mu.Unlock()
	if always || cfg.Debug {
		_, _ = fmt.Fprintln(l.out, msg)
	}
	if !cfg.Debug && !persist {
		return
	}
	// #nosec G302,G304 -- the log lives in MiSTer's world-readable /tmp.
	file, err := os.OpenFile(filepath.Clean(l.paths.LogFile), os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		return
	}
	defer func() { _ = file.Close() }()
	_, _ = fmt.Fprintf(file, "[%s] %s\n", isoformat(time.Now()), msg)
}

// isoformat matches datetime.isoformat(): the microsecond fraction is
// omitted when zero and otherwise always six digits.
func isoformat(now time.Time) string {
	stamp := now.Format("2006-01-02T15:04:05")
	if micros := now.Nanosecond() / 1000; micros != 0 {
		stamp += fmt.Sprintf(".%06d", micros)
	}
	return stamp
}
