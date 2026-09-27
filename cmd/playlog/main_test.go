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
package main

import (
	"errors"
	"reflect"
	"testing"
)

func TestStartTrackingReadsARunningGame(t *testing.T) {
	t.Parallel()

	var calls []string
	err := startTracking(
		func() bool { return true },
		func() { calls = append(calls, "core") },
		func() { calls = append(calls, "game") },
		func() { calls = append(calls, "clear") },
		func() error { calls = append(calls, "watch"); return nil },
	)
	if err != nil {
		t.Fatal(err)
	}
	want := []string{"core", "game", "watch", "core", "game"}
	if !reflect.DeepEqual(calls, want) {
		t.Errorf("calls = %v, want %v", calls, want)
	}
}

func TestStartTrackingWithoutActiveGameFile(t *testing.T) {
	t.Parallel()

	var calls []string
	watchErr := errors.New("no watch")
	err := startTracking(
		func() bool { return false },
		func() { calls = append(calls, "core") },
		func() { calls = append(calls, "game") },
		func() { calls = append(calls, "clear") },
		func() error { calls = append(calls, "watch"); return watchErr },
	)
	if !errors.Is(err, watchErr) {
		t.Errorf("err = %v, want %v", err, watchErr)
	}
	want := []string{"core", "clear", "watch"}
	if !reflect.DeepEqual(calls, want) {
		t.Errorf("calls = %v, want %v", calls, want)
	}
}
