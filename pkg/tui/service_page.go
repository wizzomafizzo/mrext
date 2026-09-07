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
	"github.com/rivo/tview"
)

// ServiceActions are the operations a service page offers. They run off the UI
// goroutine, so they may block.
//
// This takes callbacks rather than a *service.Service on purpose: pkg/tui
// imports nothing from the rest of mrext, and keeping it that way is what lets
// it be reasoned about and tested on its own.
type ServiceActions struct {
	Running   func() bool
	Start     func() error
	Stop      func() error
	Restart   func() error
	Uninstall func()
	Exit      func()
	// Status renders the body for the current state. Called on every redraw.
	Status func(running bool) string
}

// ServicePage is the Scripts-menu screen shared by the daemon apps: a status
// block and a Start/Stop, Restart, Uninstall, Exit bar, with Exit selected by
// default because most visits are to read the state and leave.
type ServicePage struct {
	app     *tview.Application
	pages   *tview.Pages
	actions ServiceActions
	status  *tview.TextView
	bar     *ButtonBar
	frame   *PageFrame
	name    string
}

// NewServicePage builds the page. Call Show to put it on the pages.
func NewServicePage(
	app *tview.Application, pages *tview.Pages, name string, actions ServiceActions,
) *ServicePage {
	page := &ServicePage{
		app:     app,
		pages:   pages,
		actions: actions,
		name:    name,
		status:  tview.NewTextView().SetTextAlign(tview.AlignCenter).SetWrap(true).SetWordWrap(true),
	}
	page.build()
	return page
}

func (s *ServicePage) build() {
	s.bar = NewButtonBar(s.app).
		AddButtonWithHelp("Start", "Start or stop the "+s.name+" service", s.toggle).
		AddButtonWithHelp("Restart", "Stop and start the service again", func() {
			s.run("Restarting service...", s.actions.Restart)
		}).
		AddButtonWithHelp("Uninstall", "Remove "+s.name+" from startup and delete what it installed",
			func() {
				if s.actions.Uninstall != nil {
					s.actions.Uninstall()
				}
			}).
		AddButtonWithHelp("Exit", "Leave this screen; the service keeps running", s.exit).
		SetupNavigation(s.exit)
	s.bar.SetFocusedIndex(3)

	s.frame = NewPageFrame(s.app).
		SetTitle(s.name).
		SetContent(s.status).
		SetFocusTarget(s.bar).
		SetHelpText("Leave this screen; the service keeps running").
		SetButtonBar(s.bar).
		SetOnEscape(s.exit)
	s.bar.SetHelpCallback(func(text string) { s.frame.SetHelpText(text) })
	s.Redraw()
}

func (s *ServicePage) exit() {
	if s.actions.Exit != nil {
		s.actions.Exit()
		return
	}
	s.app.Stop()
}

func (s *ServicePage) toggle() {
	if s.actions.Running() {
		s.run("Stopping service...", s.actions.Stop)
		return
	}
	s.run("Starting service...", s.actions.Start)
}

// run performs a service command on its own goroutine. Doing it inline froze
// the screen for as long as the command took, with nothing to say why.
func (s *ServicePage) run(message string, action func() error) {
	if action == nil {
		return
	}
	s.status.SetText(message)
	go func() {
		err := action()
		s.app.QueueUpdateDraw(func() {
			s.Redraw()
			if err != nil {
				ShowErrorModal(s.pages, s.app, err.Error(), s.Redraw)
			}
		})
	}()
}

// Redraw refreshes the status block and the Start/Stop label.
func (s *ServicePage) Redraw() {
	running := s.actions.Running()
	s.status.SetText(s.actions.Status(running))
	toggle := "Start"
	if running {
		toggle = "Stop"
	}
	s.bar.UpdateButtonLabel(0, toggle)
}

// Frame exposes the page for callers that add it themselves.
func (s *ServicePage) Frame() *PageFrame { return s.frame }

// ButtonBar exposes the bar so an app can add its own action. Extra buttons
// land after Exit; the Start/Stop toggle has to stay at index 0.
func (s *ServicePage) ButtonBar() *ButtonBar { return s.bar }

// Show adds the page and focuses its buttons.
func (s *ServicePage) Show(name string) {
	s.pages.AddAndSwitchToPage(name, s.frame, true)
	s.app.SetFocus(s.bar)
}
