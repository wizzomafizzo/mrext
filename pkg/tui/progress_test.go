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
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

func TestProgressDrawsTextAndHalfFilledBar(t *testing.T) {
	screen := tcell.NewSimulationScreen("UTF-8")
	if err := screen.Init(); err != nil {
		t.Fatal(err)
	}
	defer screen.Fini()
	screen.SetSize(60, 7)
	progress := NewProgress("Creating MGL files...").
		SetUpdate(ProgressUpdate{Text: "Scanning SNES (/media/fat/games/SNES)", Current: 5, Total: 10})
	progress.SetRect(0, 0, 60, 7)
	progress.Draw(screen)
	var text strings.Builder
	for x := range 60 {
		value, _, _ := screen.Get(x, 1)
		_, _ = text.WriteString(value)
	}
	if !strings.Contains(text.String(), "Scanning SNES") {
		t.Fatalf("progress text missing: %q", text.String())
	}
	filled, empty := 0, 0
	for x := range 60 {
		value, _, _ := screen.Get(x, 5)
		switch value {
		case "#":
			filled++
		case "-":
			empty++
		}
	}
	if filled != 25 || empty != 25 {
		t.Fatalf("bar filled=%d empty=%d", filled, empty)
	}
	progress.InputHandler()(tcell.NewEventKey(tcell.KeyEscape, 0, tcell.ModNone), func(tview.Primitive) {
		t.Fatal("progress box delegated focus")
	})
}

func TestShowProgressModalRunsTaskAndThrottlesUpdates(t *testing.T) {
	screen := tcell.NewSimulationScreen("UTF-8")
	if err := screen.Init(); err != nil {
		t.Fatal(err)
	}
	app := tview.NewApplication().SetScreen(screen)
	base := tview.NewBox()
	pages := tview.NewPages().AddPage("base", base, true, true)
	app.SetRoot(pages, true).SetFocus(base)
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

	sentinel := errors.New("task finished")
	done := make(chan error, 1)
	var progress *Progress
	app.QueueUpdate(func() {
		progress = ShowProgressModal(pages, app, ProgressOptions{Title: "Working", Throttle: time.Second},
			func(update func(ProgressUpdate)) error {
				for index := range 50 {
					update(ProgressUpdate{Text: "step", Current: index, Total: 50})
				}
				return sentinel
			}, func(err error) { done <- err })
	})
	select {
	case err := <-done:
		if !errors.Is(err, sentinel) {
			t.Fatalf("onDone error = %v", err)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("progress task did not complete")
	}
	var hasPage bool
	var shown ProgressUpdate
	app.QueueUpdate(func() {
		hasPage = pages.HasPage(progressModalPage)
		shown = progress.Update()
		if app.GetFocus() != base {
			t.Error("progress did not restore focus")
		}
	})
	if hasPage {
		t.Fatal("progress modal was not removed")
	}
	if shown.Current != 0 {
		t.Fatalf("updates were not throttled: %+v", shown)
	}
}

func TestProgressOverlayConsumesMouseOutsideBox(t *testing.T) {
	overlay := &progressOverlay{Primitive: Centered(60, 7, NewProgress("Working"))}
	overlay.SetRect(0, 0, 100, 30)
	consumed, capture := overlay.MouseHandler()(
		tview.MouseLeftClick, tcell.NewEventMouse(0, 0, tcell.Button1, tcell.ModNone),
		func(tview.Primitive) { t.Error("mouse changed focus") },
	)
	if !consumed || capture != nil {
		t.Fatal("progress overlay leaked mouse input")
	}
}
