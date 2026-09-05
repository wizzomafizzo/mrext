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
	"fmt"
	"strings"
	"sync"
	"time"
	"unicode"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	"github.com/wizzomafizzo/mrext/pkg/bgm"
	"github.com/wizzomafizzo/mrext/pkg/tui"
)

const (
	pageMain      = "bgm_main"
	appTitle      = "Background Music"
	activeMarker  = " [ACTIVE]"
	statusDelay   = 200 * time.Millisecond
	refreshPeriod = 2 * time.Second
)

// status is the parsed reply to the "status" socket command.
type status struct {
	playback string
	playlist string
	track    string
	playing  bool
}

//nolint:govet // Field order groups runtime dependencies and navigation state.
type ui struct {
	paths   bgm.Paths
	cfg     bgm.Config
	logger  *bgm.Logger
	app     *tview.Application
	pages   *tview.Pages
	options tui.ApplicationOptions
	// send is the socket client; tests replace it.
	send        func(string) (string, bool, error)
	statusDelay time.Duration

	mainSelection int
	mainList      *tui.MenuList
	mainBar       *tui.ButtonBar

	mu         sync.Mutex
	lastStatus status
	runningApp *tview.Application
}

func newUI(paths *bgm.Paths, cfg *bgm.Config, logger *bgm.Logger) *ui {
	view := &ui{
		paths:       *paths,
		cfg:         *cfg,
		logger:      logger,
		statusDelay: statusDelay,
	}
	view.send = func(message string) (string, bool, error) { return bgm.Send(&view.paths, message) }
	view.options = view.applicationOptions()
	return view
}

func (u *ui) applicationOptions() tui.ApplicationOptions {
	return tui.ApplicationOptions{
		Theme:            u.cfg.TUI.Theme,
		Mouse:            u.cfg.TUI.Mouse,
		CRTMode:          u.cfg.TUI.CRTMode,
		OnScreenKeyboard: u.cfg.TUI.OnScreenKeyboard,
	}
}

func (u *ui) Run() error {
	builder := func() (*tview.Application, error) {
		app, err := tui.NewApplication(u.options)
		if err != nil {
			return nil, fmt.Errorf("create TUI application: %w", err)
		}
		u.app = app
		u.pages = tview.NewPages()
		if err := u.showMain(); err != nil {
			return nil, err
		}
		app.SetRoot(tui.WrapRoot(u.options, u.pages), true)
		app.SetAfterDrawFunc(func(tcell.Screen) { u.markRunning(app) })
		return app, nil
	}
	stop := make(chan struct{})
	defer close(stop)
	go u.refreshLoop(stop)
	if err := tui.BuildAndRetry(builder); err != nil {
		return fmt.Errorf("run BGM TUI: %w", err)
	}
	return nil
}

func (u *ui) markRunning(app *tview.Application) {
	u.mu.Lock()
	defer u.mu.Unlock()
	u.runningApp = app
}

func (u *ui) runningApplication() *tview.Application {
	u.mu.Lock()
	defer u.mu.Unlock()
	return u.runningApp
}

func (u *ui) setLastStatus(current status) {
	u.mu.Lock()
	defer u.mu.Unlock()
	u.lastStatus = current
}

func (u *ui) statusChanged(current status) bool {
	u.mu.Lock()
	defer u.mu.Unlock()
	return u.lastStatus != current
}

// refreshLoop keeps "Now playing" current while the main screen is idle.
func (u *ui) refreshLoop(stop <-chan struct{}) {
	ticker := time.NewTicker(refreshPeriod)
	defer ticker.Stop()
	for {
		select {
		case <-stop:
			return
		case <-ticker.C:
			app := u.runningApplication()
			if app == nil {
				continue
			}
			current, err := u.fetchStatus(false)
			if err != nil || !u.statusChanged(current) {
				continue
			}
			app.QueueUpdateDraw(func() { u.refreshMain(current) })
		}
	}
}

// refreshMain rebuilds the main screen for a new status only while it is the
// front page and the list has focus, so modals and the Settings page are
// never disturbed. The cursor row and the highlighted button are kept.
func (u *ui) refreshMain(current status) bool {
	if name, _ := u.pages.GetFrontPage(); name != pageMain || u.mainList == nil || !u.mainList.HasFocus() {
		return false
	}
	u.mainSelection = u.mainList.GetCurrentItem()
	button := u.mainBar.FocusedIndex()
	if err := u.buildMain(current); err != nil {
		u.showError(err, u.app.Stop)
		return false
	}
	u.mainBar.SetFocusedIndex(button)
	return true
}

// fetchStatus mirrors get_status(), including the pause that gives the
// service time to apply the previous command.
func (u *ui) fetchStatus(delay bool) (status, error) {
	if delay {
		time.Sleep(u.statusDelay)
	}
	reply, replied, err := u.send("status")
	if err != nil {
		return status{}, err
	}
	if !replied {
		return status{}, errors.New("BGM service is not running")
	}
	parts := strings.Split(reply, "\t")
	if len(parts) < 4 {
		return status{}, fmt.Errorf("invalid status from BGM service: %q", reply)
	}
	return status{playing: parts[0] == "yes", playback: parts[1], playlist: parts[2], track: parts[3]}, nil
}

func (u *ui) showMain() error {
	current, err := u.fetchStatus(false)
	if err != nil {
		return err
	}
	return u.buildMain(current)
}

type mainRow struct {
	action func()
	label  string
	help   string
	kind   tui.MenuRowKind
}

func (u *ui) mainRows(current status) ([]mainRow, error) {
	playlists, err := bgm.Playlists(&u.paths)
	if err != nil {
		return nil, fmt.Errorf("list BGM playlists: %w", err)
	}
	rows := []mainRow{
		{
			label: "Skip current track", help: "Stop the current track and move on.", kind: tui.MenuRowAction,
			action: func() { u.command("skip", nil) },
		},
	}
	if current.playing {
		rows = append(rows, mainRow{
			label: "Stop playing", help: "Stop background music until started again.",
			kind: tui.MenuRowAction, action: func() { u.command("stop", nil) },
		})
	} else {
		rows = append(rows, mainRow{
			label: "Start playing", help: "Start the configured playlist.",
			kind: tui.MenuRowAction, action: func() { u.command("play", nil) },
		})
	}
	rows = append(rows, mainRow{label: "Playback", kind: tui.MenuRowHeader})
	playbackRows := []struct{ label, help, value string }{
		{"Play random tracks (random)", "Keep picking random tracks from the playlist.", bgm.PlaybackRandom},
		{"Play a single random track on repeat (loop)", "Pick one track and repeat it.", bgm.PlaybackLoop},
		{"Disable all playback (disabled)", "Play nothing except boot sounds.", bgm.PlaybackDisabled},
	}
	for _, row := range playbackRows {
		rows = append(rows, mainRow{
			label: row.label + active(current.playback == row.value), help: row.help, kind: tui.MenuRowItem,
			action: func() {
				u.command("set playback "+row.value, func(updated status) error {
					return bgm.SavePlayback(u.paths.IniFile, updated.playback)
				})
			},
		})
	}
	rows = append(rows, mainRow{label: "Playlist", kind: tui.MenuRowHeader})
	playlistRows := []struct{ label, help, value string }{
		{"None (just top level files)", "Play only files in the top level of the music folder.", "none"},
		{"All (all playlists combined)", "Play every file from every playlist folder.", "all"},
	}
	for _, name := range playlists {
		playlistRows = append(playlistRows, struct{ label, help, value string }{
			name, "Play files from the " + name + " folder.", name,
		})
	}
	for _, row := range playlistRows {
		kind := tui.MenuRowItem
		if row.value != "none" && row.value != "all" {
			kind = tui.MenuRowFolder
		}
		rows = append(rows, mainRow{
			label: row.label + active(current.playlist == row.value), help: row.help, kind: kind,
			action: func() {
				u.command("set playlist "+row.value, func(updated status) error {
					return bgm.SavePlaylist(u.paths.IniFile, bgm.ParsePlaylist(updated.playlist))
				})
			},
		})
	}
	return rows, nil
}

func (u *ui) buildMain(current status) error {
	u.setLastStatus(current)
	rows, err := u.mainRows(current)
	if err != nil {
		return err
	}
	nowPlaying := current.track
	if nowPlaying == "" {
		nowPlaying = "---"
	}
	details := tview.NewTextView().
		SetDynamicColors(true).
		SetText(strings.Join([]string{
			formatDetail("Now playing", nowPlaying),
			formatDetail("Playback", titleCase(current.playback)),
			formatDetail("Playlist", current.playlist),
		}, "\n"))
	list := tui.NewMenuList()
	for _, row := range rows {
		list.AddRow(row.label, row.kind)
	}
	list.SetCurrentItem(min(u.mainSelection, list.GetItemCount()-1))
	activate := func() {
		index := list.GetCurrentItem()
		u.mainSelection = index
		if index >= 0 && index < len(rows) && rows[index].action != nil {
			rows[index].action()
		}
	}
	list.SetSelectedFunc(func(int, string, string, rune) { activate() })
	bar := tui.NewButtonBar(u.app).
		AddButtonWithHelp("Select", "Run the selected action", activate).
		AddButtonWithHelp("Settings", "Configure Background Music", u.startSettings).
		AddButtonWithHelp("Exit", "Exit Background Music", u.app.Stop).
		SetupNavigation(u.app.Stop)
	content := tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(details, 3, 0, false).
		AddItem(list, 0, 1, true)
	frame := tui.NewPageFrame(u.app).
		SetTitle(appTitle).
		SetContent(content).
		SetFocusTarget(list).
		SetHelpText("Select a playback action, playback type or playlist.").
		SetButtonBar(bar).
		SetOnEscape(u.app.Stop)
	showHelp := func(index int) {
		if index >= 0 && index < len(rows) && rows[index].help != "" {
			frame.SetHelpText(rows[index].help)
		}
	}
	list.SetRowChangedFunc(showHelp)
	showHelp(list.GetCurrentItem())
	bar.SetHelpCallback(func(text string) {
		if text == "" {
			showHelp(list.GetCurrentItem())
		} else {
			frame.SetHelpText(text)
		}
	})
	frame.SetupContentToButtonNavigation()
	u.mainList = list
	u.mainBar = bar
	u.pages.AddAndSwitchToPage(pageMain, frame, true)
	u.app.SetFocus(list)
	return nil
}

// command sends one socket command, refreshes the status the way the Python
// menu did, optionally persists the result and rebuilds the main screen.
func (u *ui) command(message string, persist func(status) error) {
	if _, _, err := u.send(message); err != nil {
		u.showError(err, u.app.Stop)
		return
	}
	current, err := u.fetchStatus(true)
	if err != nil {
		u.showError(err, u.app.Stop)
		return
	}
	if persist != nil {
		if err := persist(current); err != nil {
			u.showError(err, u.mustShowMain)
			return
		}
	}
	if err := u.buildMain(current); err != nil {
		u.showError(err, u.app.Stop)
	}
}

func (u *ui) showError(err error, restore func()) {
	tui.ShowErrorModal(u.pages, u.app, err.Error(), restore)
}

func (u *ui) mustShowMain() {
	if err := u.showMain(); err != nil {
		u.showError(err, u.app.Stop)
	}
}

func active(condition bool) string {
	if condition {
		return activeMarker
	}
	return ""
}

func formatDetail(label, value string) string {
	theme := tui.CurrentTheme()
	return fmt.Sprintf("[%s::b]%s:[-::-] %s", theme.LabelName, label, tview.Escape(value))
}

// titleCase mirrors Python's str.title() used for the playback type.
func titleCase(value string) string {
	var out strings.Builder
	previousLetter := false
	for _, character := range value {
		switch {
		case !unicode.IsLetter(character):
			_, _ = out.WriteRune(character)
			previousLetter = false
		case previousLetter:
			_, _ = out.WriteRune(unicode.ToLower(character))
		default:
			_, _ = out.WriteRune(unicode.ToUpper(character))
			previousLetter = true
		}
	}
	return out.String()
}
