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

func TestProgressBoxFitsWhatItHolds(t *testing.T) {
	wide := ProgressUpdate{Text: strings.Repeat("x", 30)}
	overflow := ProgressUpdate{Text: strings.Repeat("scanning a very long folder path ", 10)}
	bar := ProgressUpdate{Text: "Scanning", Current: 1, Total: 10}
	for _, testCase := range []struct {
		name       string
		title      string
		update     ProgressUpdate
		wantWidth  int
		wantHeight int
	}{
		{"one short line", "", ProgressUpdate{Text: "Searching..."}, progressModalMinWidth, 3},
		{"nothing to say yet", "", ProgressUpdate{}, progressModalMinWidth, 3},
		{"a line past the minimum", "", wide, 34, 3},
		{"a title wider than the text", "Creating MGL files...", ProgressUpdate{Text: "Done"}, 27, 3},
		{"a bar needs room for itself", "", bar, progressBarWidth + 4, 5},
		{"text stops at the ceiling", "", overflow, progressModalMaxWidth, progressModalMaxLines + 2},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			width := progressBoxWidth(testCase.title, testCase.update)
			if width != testCase.wantWidth {
				t.Fatalf("width = %d, want %d", width, testCase.wantWidth)
			}
			if height := progressBoxHeight(width, testCase.update); height != testCase.wantHeight {
				t.Fatalf("height = %d, want %d", height, testCase.wantHeight)
			}
		})
	}
}

// A title of "" was still written to the border as two spaces, leaving a gap
// where a name should be.
func TestProgressBoxWithoutATitleDrawsAnUnbrokenBorder(t *testing.T) {
	if !SetCurrentTheme("default") {
		t.Fatal("failed to select default theme")
	}
	screen := tcell.NewSimulationScreen("UTF-8")
	if err := screen.Init(); err != nil {
		t.Fatal(err)
	}
	defer screen.Fini()
	update := ProgressUpdate{Text: "Searching..."}
	width := progressBoxWidth("", update)
	height := progressBoxHeight(width, update)
	screen.SetSize(width, height)
	progress := NewProgress("").SetUpdate(update)
	progress.SetRect(0, 0, width, height)
	progress.Draw(screen)

	var border strings.Builder
	for x := range width {
		value, _, _ := screen.Get(x, 0)
		_, _ = border.WriteString(value)
	}
	if strings.Contains(border.String(), " ") {
		t.Fatalf("untitled border has a gap in it: %q", border.String())
	}

	var text strings.Builder
	for x := range width {
		value, _, _ := screen.Get(x, 1)
		_, _ = text.WriteString(value)
	}
	if !strings.Contains(text.String(), "Searching...") {
		t.Fatalf("progress text missing: %q", text.String())
	}
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

// runningProgressApp starts an application on a simulation screen with one
// focusable page underneath, and stops it when the test ends.
func runningProgressApp(t *testing.T) (*tview.Application, *tview.Pages, tview.Primitive) {
	t.Helper()
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
	return app, pages, base
}

func TestProgressDrawsNothingWhileHidden(t *testing.T) {
	screen := tcell.NewSimulationScreen("UTF-8")
	if err := screen.Init(); err != nil {
		t.Fatal(err)
	}
	defer screen.Fini()
	screen.SetSize(60, 7)
	progress := NewProgress("Favorites").
		SetUpdate(ProgressUpdate{Text: "Checking favorites...", Current: 1, Total: 2}).
		SetHidden(true)
	progress.SetRect(0, 0, 60, 7)
	progress.Draw(screen)
	for y := range 7 {
		for x := range 60 {
			value, _, _ := screen.Get(x, y)
			if value != " " {
				t.Fatalf("hidden progress drew %q at %d,%d", value, x, y)
			}
		}
	}
}

// A modal that appears and disappears inside the same second reads as a glitch
// rather than as feedback, so the box waits. The overlay does not: whatever the
// task is doing, it does it alone.
func TestShowProgressModalBlocksInputWhileTheBoxIsStillHidden(t *testing.T) {
	app, pages, _ := runningProgressApp(t)
	release := make(chan struct{})
	done := make(chan error, 1)
	var progress *Progress
	app.QueueUpdate(func() {
		progress = ShowProgressModal(pages, app, ProgressOptions{Title: "Working", Delay: time.Hour},
			func(func(ProgressUpdate)) error {
				<-release
				return nil
			}, func(err error) { done <- err })
	})

	var hasPage, focused, hidden bool
	app.QueueUpdate(func() {
		hasPage = pages.HasPage(progressModalPage)
		focused = app.GetFocus() == tview.Primitive(progress)
		hidden = progress.Hidden()
	})
	if !hasPage {
		t.Fatal("progress overlay was not added")
	}
	if !focused {
		t.Fatal("progress overlay did not take focus, so keys reach the page underneath")
	}
	if !hidden {
		t.Fatal("progress box drew before its delay elapsed")
	}

	close(release)
	select {
	case <-done:
	case <-time.After(5 * time.Second):
		t.Fatal("progress task did not complete")
	}
}

func TestShowProgressModalShowsTheBoxWhenWorkOutlastsTheDelay(t *testing.T) {
	app, pages, _ := runningProgressApp(t)
	release := make(chan struct{})
	done := make(chan error, 1)
	var progress *Progress
	app.QueueUpdate(func() {
		progress = ShowProgressModal(pages, app, ProgressOptions{Title: "Working", Delay: time.Millisecond},
			func(func(ProgressUpdate)) error {
				<-release
				return nil
			}, func(err error) { done <- err })
	})

	var visible bool
	deadline := time.Now().Add(5 * time.Second)
	for !visible && time.Now().Before(deadline) {
		app.QueueUpdate(func() { visible = !progress.Hidden() })
		if visible {
			break
		}
		time.Sleep(time.Millisecond)
	}
	if !visible {
		t.Fatal("progress box never appeared for work that outlasted its delay")
	}

	close(release)
	select {
	case <-done:
	case <-time.After(5 * time.Second):
		t.Fatal("progress task did not complete")
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
