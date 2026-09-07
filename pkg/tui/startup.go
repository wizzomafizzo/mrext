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
	"sync"
	"sync/atomic"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

// RunWhenStarted schedules work to run once the application's event loop is
// live, and never if it is not. The returned flag reports whether that
// happened, which is how a caller tells a real failure from a headless run
// where no terminal was ever available.
//
// This exists because BuildAndRetry calls its builder a second time to retry
// on /dev/tty2. Work started directly inside a builder therefore runs twice,
// concurrently, against the same files.
//
// QueueUpdateDraw cannot be used from the builder either: it blocks until the
// event loop executes the callback, so calling it before Run deadlocks. The
// after-draw hook only fires when the application is really drawing, which is
// exactly the signal needed, and the work is then queued from a goroutine so
// it mutates the UI on the event loop rather than mid-draw.
//
// It installs an after-draw handler, so it must not be combined with a caller
// that sets its own.
func RunWhenStarted(app *tview.Application, work func()) *atomic.Bool {
	var started atomic.Bool
	var once sync.Once
	app.SetAfterDrawFunc(func(tcell.Screen) {
		once.Do(func() {
			started.Store(true)
			go app.QueueUpdateDraw(work)
		})
	})
	return &started
}
