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

//nolint:gosec // Tests operate only on paths rooted in temporary directories.
package bgm

import (
	"os"
	"regexp"
	"strings"
	"testing"
	"time"
)

func TestLoggerHonoursDebugFlagPerCall(t *testing.T) {
	paths := newTestPaths(t)
	logger, out := newTestLogger(&paths)
	writeINI(t, &paths, DefaultINI)

	logger.Log("hidden")
	logger.Print("shown")
	if out.String() != "shown\n" {
		t.Fatalf("output %q", out.String())
	}
	if _, err := os.Stat(paths.LogFile); !os.IsNotExist(err) {
		t.Fatal("log file must not be written without debug")
	}

	writeINI(t, &paths, "[bgm]\ndebug = yes\n")
	logger.Log("")
	logger.Log("visible")
	if out.String() != "shown\nvisible\n" {
		t.Fatalf("output %q", out.String())
	}
	line := strings.TrimSpace(readFile(t, paths.LogFile))
	if !regexp.MustCompile(`^\[\d{4}-\d\d-\d\dT\d\d:\d\d:\d\d(\.\d{6})?\] visible$`).MatchString(line) {
		t.Fatalf("log line %q", line)
	}
}

func TestLoggerRecreatesMissingINI(t *testing.T) {
	paths := newTestPaths(t)
	logger, _ := newTestLogger(&paths)
	logger.Log("anything")
	if got := readFile(t, paths.IniFile); got != DefaultINI {
		t.Fatalf("ini = %q", got)
	}
}

func TestIsoformatOmitsZeroMicroseconds(t *testing.T) {
	base := time.Date(2026, 9, 5, 17, 25, 19, 0, time.UTC)
	if got := isoformat(base); got != "2026-09-05T17:25:19" {
		t.Fatalf("isoformat = %q", got)
	}
	if got := isoformat(base.Add(837446 * time.Microsecond)); got != "2026-09-05T17:25:19.837446" {
		t.Fatalf("isoformat = %q", got)
	}
	if got := isoformat(base.Add(5 * time.Microsecond)); got != "2026-09-05T17:25:19.000005" {
		t.Fatalf("isoformat = %q", got)
	}
}
