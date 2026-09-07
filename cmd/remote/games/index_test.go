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
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/wizzomafizzo/mrext/pkg/gamesdb"
)

func nextIndexStatus(t *testing.T, updates <-chan string) string {
	t.Helper()
	select {
	case status := <-updates:
		return status
	case <-time.After(5 * time.Second):
		t.Fatal("index status was not published")
		return ""
	}
}

func TestIndexReportsFailureAndRejectsOverlappingBuild(t *testing.T) {
	index := NewIndex()
	updates := make(chan string, 10)
	release := make(chan struct{})
	failure := errors.New("fixture, scan\nfailed")
	logged := make(chan error, 1)
	publish := func() { updates <- index.status(true) }
	if !index.start(func(update func(gamesdb.IndexStatus)) (int, error) {
		update(gamesdb.IndexStatus{Total: 3, Step: 2, SystemID: "NES"})
		<-release
		return 0, failure
	}, publish, func(err error) { logged <- err }) {
		t.Fatal("first build did not start")
	}
	if got := nextIndexStatus(t, updates); !strings.HasPrefix(got, "indexStatus:y,y,") {
		t.Fatalf("initial status = %q", got)
	}
	if got := nextIndexStatus(t, updates); !strings.Contains(got, ",3,2,Indexing ") {
		t.Fatalf("progress status = %q", got)
	}
	if index.start(nil, publish, nil) {
		t.Fatal("overlapping build was accepted")
	}
	close(release)
	want := "indexStatus:y,n,0,0,Index failed: fixture; scan failed"
	if got := nextIndexStatus(t, updates); got != want {
		t.Fatalf("failure status = %q, want %q", got, want)
	}
	if err := <-logged; !errors.Is(err, failure) {
		t.Fatalf("logged error = %v", err)
	}
	if got := index.status(true); got != want {
		t.Fatalf("reconnecting client lost failure: %q", got)
	}
	if !index.start(func(update func(gamesdb.IndexStatus)) (int, error) {
		update(gamesdb.IndexStatus{Total: 2, Step: 2, Files: 4})
		return 4, nil
	}, publish, func(err error) { logged <- err }) {
		t.Fatal("retry did not start")
	}
	if got := nextIndexStatus(t, updates); strings.Contains(got, "failed") {
		t.Fatalf("retry retained previous error: %q", got)
	}
	if got := nextIndexStatus(t, updates); !strings.Contains(got, "Writing database... (4 games)") {
		t.Fatalf("writing status = %q", got)
	}
	if got := nextIndexStatus(t, updates); got != "indexStatus:y,n,0,0," {
		t.Fatalf("success status = %q", got)
	}
}
