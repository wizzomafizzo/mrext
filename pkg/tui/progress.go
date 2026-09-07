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
	"strings"
	"sync"
	"time"
	"unicode/utf8"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

const (
	progressModalPage = "tui_progress_modal"
	// The box fits itself to what it holds, between these bounds. The ceiling
	// keeps a long step name from running off a 75-column CRT screen, and the
	// line cap makes it wrap rather than grow past 15 rows.
	progressModalMinWidth   = 24
	progressModalMaxWidth   = 60
	progressModalMaxLines   = 3
	progressBarWidth        = 50
	defaultProgressThrottle = 150 * time.Millisecond
	defaultProgressDelay    = 500 * time.Millisecond
)

// ProgressOptions configures ShowProgressModal. Title is optional and belongs
// in the border only when it says something the text does not: naming the
// operation while the text names the item it is working on. A box reporting
// one thing should leave it empty rather than repeat the app name.
//
// Throttle bounds how often the task goroutine may redraw the modal; zero
// selects a 150ms minimum interval.
//
// Delay is how long the task may run before the box is drawn at all; zero
// selects 500ms and a negative value draws it immediately. Work that finishes
// inside the delay never shows a box, so a fast device does not flash one up
// and tear it down again. Input is blocked for the whole task either way.
type ProgressOptions struct {
	Title    string
	Initial  ProgressUpdate
	Throttle time.Duration
	Delay    time.Duration
}

// Progress is a bordered message box with an optional bar. It swallows all
// keys so a running task cannot be interrupted by the controller.
type Progress struct {
	*tview.Box
	update ProgressUpdate
	hidden bool
}

func NewProgress(title string) *Progress {
	theme := CurrentTheme()
	progress := &Progress{Box: tview.NewBox()}
	progress.SetBackgroundColor(tview.Styles.ContrastBackgroundColor).
		SetBorder(true).
		SetBorderColor(theme.BorderColor).
		SetTitleColor(theme.BorderColor).
		SetTitleAlign(tview.AlignCenter)
	// An empty title still clears two cells of the top border, leaving a gap
	// where a name should be. Draw the border unbroken instead.
	if title != "" {
		progress.SetTitle(" " + title + " ")
	}
	return progress
}

// progressBoxWidth fits the box to the longest line it has to show, plus room
// for the bar and the title when either is there. A fixed 60 cells made a box
// that says one word look like a dialog waiting for an answer.
func progressBoxWidth(title string, update ProgressUpdate) int {
	width := progressModalMinWidth
	if title != "" {
		width = max(width, utf8.RuneCountInString(title)+6)
	}
	for _, line := range strings.Split(update.Text, "\n") {
		width = max(width, utf8.RuneCountInString(line)+4)
	}
	if update.Total > 0 {
		width = max(width, progressBarWidth+4)
	}
	return min(width, progressModalMaxWidth)
}

// progressBoxHeight is a border, the text wrapped to the given width, and a
// blank line plus the bar when there is one. A fixed height left a one-word
// message sitting on top of four empty rows.
func progressBoxHeight(width int, update ProgressUpdate) int {
	lines := min(max(len(tview.WordWrap(update.Text, width-2)), 1), progressModalMaxLines)
	height := lines + 2
	if update.Total > 0 {
		height += 2
	}
	return height
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

// SetHidden stops the box drawing itself without taking it off the page, so it
// keeps swallowing keys and mouse clicks while showing nothing. Call it on the
// UI goroutine.
func (p *Progress) SetHidden(hidden bool) *Progress {
	p.hidden = hidden
	return p
}

// Hidden reports whether the box is currently drawing nothing.
func (p *Progress) Hidden() bool {
	return p.hidden
}

func (p *Progress) Draw(screen tcell.Screen) {
	if p.hidden {
		return
	}
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

// ShowProgressModal covers pages with an input-blocking overlay and runs task
// on a new goroutine. The update callback handed to task may be called from
// that goroutine as often as wanted; redraws are throttled and delivered with
// app.QueueUpdateDraw, so the application must be running. When task returns
// the overlay is removed and onDone runs on the UI goroutine with its error.
// update must not be called after task returns.
//
// The overlay goes up straight away but the box inside it stays hidden until
// options.Delay has passed, so quick work is silent rather than a box that
// appears and vanishes in the same breath. The overlay is transparent while
// the box is hidden, and it swallows keys and mouse clicks throughout, so the
// task always has the screen to itself.
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
	delay := options.Delay
	if delay == 0 {
		delay = defaultProgressDelay
	}
	previousFocus := app.GetFocus()
	progress := NewProgress(options.Title).SetUpdate(options.Initial).SetHidden(delay > 0)
	width := progressBoxWidth(options.Title, options.Initial)
	height := progressBoxHeight(width, options.Initial)
	show := func() {
		overlay := &progressOverlay{Primitive: Centered(width, height, progress)}
		pages.AddPage(progressModalPage, overlay, true, true)
		app.SetFocus(progress)
	}
	show()

	// done and the hidden flag are only touched from queued callbacks, so the
	// event loop orders them and no lock is needed.
	var done bool
	var reveal *time.Timer
	if delay > 0 {
		reveal = time.AfterFunc(delay, func() {
			app.QueueUpdateDraw(func() {
				if done {
					return
				}
				progress.SetHidden(false)
			})
		})
	}

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
		app.QueueUpdateDraw(func() {
			progress.SetUpdate(update)
			// A bar arriving, or a longer step name, needs a bigger box. Only
			// ever grow it: resizing to every step would leave the modal
			// twitching around the screen for the whole run.
			nextWidth := max(width, progressBoxWidth(options.Title, update))
			nextHeight := max(height, progressBoxHeight(nextWidth, update))
			if nextWidth != width || nextHeight != height {
				width, height = nextWidth, nextHeight
				show()
			}
		})
	}

	go func() {
		err := task(update)
		if reveal != nil {
			reveal.Stop()
		}
		app.QueueUpdateDraw(func() {
			done = true
			pages.RemovePage(progressModalPage)
			app.SetFocus(previousFocus)
			if onDone != nil {
				onDone(err)
			}
		})
	}()
	return progress
}
