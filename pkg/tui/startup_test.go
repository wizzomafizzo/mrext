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

package tui

import (
	"testing"
	"time"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

func TestRunWhenStartedIgnoresAnApplicationThatNeverDraws(t *testing.T) {
	// BuildAndRetry builds a second application to retry on /dev/tty2, so work
	// started directly in a builder runs twice against the same files. The
	// abandoned application never draws, so it must never start anything.
	abandoned := tview.NewApplication()
	ran := false
	started := RunWhenStarted(abandoned, func() { ran = true })

	time.Sleep(50 * time.Millisecond)

	if started.Load() {
		t.Error("reported started for an application that never ran")
	}
	if ran {
		t.Error("work ran on an application that never drew")
	}
}

func TestRunWhenStartedRunsOnceOnTheApplicationThatDraws(t *testing.T) {
	screen := tcell.NewSimulationScreen("UTF-8")
	if err := screen.Init(); err != nil {
		t.Fatal(err)
	}
	app := tview.NewApplication().SetScreen(screen)
	app.SetRoot(tview.NewBox(), true)

	runs := make(chan struct{}, 8)
	started := RunWhenStarted(app, func() { runs <- struct{}{} })

	runErr := make(chan error, 1)
	go func() { runErr <- app.Run() }()
	t.Cleanup(func() {
		app.Stop()
		select {
		case <-runErr:
		case <-time.After(5 * time.Second):
			t.Error("application did not stop")
		}
	})

	select {
	case <-runs:
	case <-time.After(5 * time.Second):
		t.Fatal("work never ran on a drawing application")
	}
	if !started.Load() {
		t.Error("started flag not set")
	}

	// Redraws must not start it again.
	app.QueueUpdateDraw(func() {})
	app.QueueUpdateDraw(func() {})
	select {
	case <-runs:
		t.Error("work ran more than once")
	case <-time.After(100 * time.Millisecond):
	}
}
