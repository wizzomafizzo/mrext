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
	"time"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

const (
	progressModalPage       = "tui_progress_modal"
	progressModalWidth      = 60
	progressModalHeight     = 7
	progressBarWidth        = 50
	defaultProgressThrottle = 150 * time.Millisecond
)

// ProgressOptions configures ShowProgressModal. Throttle bounds how often the
// task goroutine may redraw the modal; zero selects a 150ms minimum interval.
type ProgressOptions struct {
	Title    string
	Initial  ProgressUpdate
	Throttle time.Duration
}

// Progress is a bordered message box with an optional bar. It swallows all
// keys so a running task cannot be interrupted by the controller.
type Progress struct {
	*tview.Box
	update ProgressUpdate
}

func NewProgress(title string) *Progress {
	theme := CurrentTheme()
	progress := &Progress{Box: tview.NewBox()}
	progress.SetBackgroundColor(tview.Styles.ContrastBackgroundColor).
		SetBorder(true).
		SetBorderColor(theme.BorderColor).
		SetTitleColor(theme.BorderColor).
		SetTitleAlign(tview.AlignCenter).
		SetTitle(" " + title + " ")
	return progress
}

// SetUpdate replaces the displayed text and bar. Call it on the UI goroutine.
func (p *Progress) SetUpdate(update ProgressUpdate) *Progress {
	p.update = update
	return p
}

// Update returns the most recently displayed state.
func (p *Progress) Update() ProgressUpdate {
	return p.update
}

func (p *Progress) Draw(screen tcell.Screen) {
	p.DrawForSubclass(screen, p)
	x, y, width, height := p.GetInnerRect()
	if width <= 0 || height <= 0 {
		return
	}
	theme := CurrentTheme()
	textRows := height
	if p.update.Total > 0 {
		textRows = max(height-2, 1)
	}
	lines := tview.WordWrap(p.update.Text, width)
	if len(lines) > textRows {
		lines = lines[:textRows]
	}
	for index, line := range lines {
		tview.Print(screen, tview.Escape(line), x, y+index, width, tview.AlignCenter, theme.PrimaryTextColor)
	}
	if p.update.Total <= 0 || height < 2 {
		return
	}
	barWidth := min(width, progressBarWidth)
	barX := x + (width-barWidth)/2
	barY := y + height - 1
	current := min(max(p.update.Current, 0), p.update.Total)
	filled := current * barWidth / p.update.Total
	fill := tcell.StyleDefault.Foreground(theme.ProgressFillColor).Background(tview.Styles.ContrastBackgroundColor)
	empty := tcell.StyleDefault.Foreground(theme.ProgressEmptyColor).Background(tview.Styles.ContrastBackgroundColor)
	for index := range barWidth {
		if index < filled {
			screen.SetContent(barX+index, barY, '#', nil, fill)
		} else {
			screen.SetContent(barX+index, barY, '-', nil, empty)
		}
	}
}

func (p *Progress) InputHandler() func(*tcell.EventKey, func(tview.Primitive)) {
	return p.WrapInputHandler(func(*tcell.EventKey, func(tview.Primitive)) {})
}

// Cover the entire page, not just the centered box, so mouse clicks cannot
// activate underlying actions while a filesystem operation is running.
type progressOverlay struct{ tview.Primitive }

func (*progressOverlay) MouseHandler() func(
	tview.MouseAction, *tcell.EventMouse, func(tview.Primitive),
) (bool, tview.Primitive) {
	return func(tview.MouseAction, *tcell.EventMouse, func(tview.Primitive)) (bool, tview.Primitive) {
		return true, nil
	}
}

// ShowProgressModal overlays a progress box on pages and runs task on a new
// goroutine. The update callback handed to task may be called from that
// goroutine as often as wanted; redraws are throttled and delivered with
// app.QueueUpdateDraw, so the application must be running. When task returns
// the modal is removed and onDone runs on the UI goroutine with its error.
// update must not be called after task returns.
func ShowProgressModal(
	pages *tview.Pages,
	app *tview.Application,
	options ProgressOptions,
	task func(update func(ProgressUpdate)) error,
	onDone func(error),
) *Progress {
	throttle := options.Throttle
	if throttle <= 0 {
		throttle = defaultProgressThrottle
	}
	previousFocus := app.GetFocus()
	progress := NewProgress(options.Title).SetUpdate(options.Initial)
	overlay := &progressOverlay{Primitive: Centered(progressModalWidth, progressModalHeight, progress)}
	pages.AddPage(progressModalPage, overlay, true, true)
	app.SetFocus(progress)

	var mu sync.Mutex
	var last time.Time
	update := func(update ProgressUpdate) {
		mu.Lock()
		now := time.Now()
		if !last.IsZero() && now.Sub(last) < throttle {
			mu.Unlock()
			return
		}
		last = now
		mu.Unlock()
		app.QueueUpdateDraw(func() { progress.SetUpdate(update) })
	}

	go func() {
		err := task(update)
		app.QueueUpdateDraw(func() {
			pages.RemovePage(progressModalPage)
			app.SetFocus(previousFocus)
			if onDone != nil {
				onDone(err)
			}
		})
	}()
	return progress
}
