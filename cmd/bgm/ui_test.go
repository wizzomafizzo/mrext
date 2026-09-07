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

//nolint:gosec // Tests operate only on paths rooted in temporary directories.
package main

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	"github.com/wizzomafizzo/mrext/pkg/bgm"
	"github.com/wizzomafizzo/mrext/pkg/tui"
)

const stubPlayer = "#!/bin/sh\nexec sleep 30\n"

func installStubPlayers(t *testing.T) {
	t.Helper()
	dir := t.TempDir()
	for _, name := range []string{"mpg123", "ogg123", "aplay", "aplaymidi", "vgmplay"} {
		if err := os.WriteFile(filepath.Join(dir, name), []byte(stubPlayer), 0o755); err != nil {
			t.Fatal(err)
		}
	}
	t.Setenv("PATH", dir+string(os.PathListSeparator)+os.Getenv("PATH"))
}

func sendUIKey(view *ui, key tcell.Key, value rune) {
	view.app.GetFocus().InputHandler()(tcell.NewEventKey(key, value, tcell.ModNone), func(p tview.Primitive) {
		view.app.SetFocus(p)
	})
}

// newWorkflowUI runs a fake service on a temporary root and opens the main
// screen headlessly.
func newWorkflowUI(t *testing.T) (*ui, *bgm.Player) {
	t.Helper()
	installStubPlayers(t)
	application, _ := newTestApp(t)
	writeTestFile(t, application.paths.IniFile, bgm.DefaultINI)
	writeTestFile(t, filepath.Join(application.paths.MusicFolder, "song.wav"), "")
	writeTestFile(t, filepath.Join(application.paths.MusicFolder, "chip", "one.wav"), "")
	player := startFakeService(t, application)
	cfg, err := bgm.LoadConfig(&application.paths)
	if err != nil {
		t.Fatal(err)
	}
	view := newUI(&application.paths, &cfg, application.logger)
	view.statusDelay = 0
	app, err := tui.NewApplication(tui.ApplicationOptions{Theme: "default"})
	if err != nil {
		t.Fatal(err)
	}
	view.app = app
	view.pages = tview.NewPages()
	if err := view.showMain(); err != nil {
		t.Fatal(err)
	}
	return view, player
}

// rowLabels returns the rendered row text; tview escapes the marker's
// closing bracket, so tests compare against escapedActive.
func rowLabels(view *ui) []string {
	labels := make([]string, 0, view.mainList.GetItemCount())
	for index := range view.mainList.GetItemCount() {
		text, _ := view.mainList.GetItemText(index)
		labels = append(labels, text)
	}
	return labels
}

var escapedActive = tview.Escape(activeMarker)

func findRow(t *testing.T, view *ui, fragment string) int {
	t.Helper()
	for index, label := range rowLabels(view) {
		if strings.Contains(label, fragment) {
			return index
		}
	}
	t.Fatalf("row %q not found in %v", fragment, rowLabels(view))
	return -1
}

func selectRow(t *testing.T, view *ui, fragment string) {
	t.Helper()
	view.mainList.SetCurrentItem(findRow(t, view, fragment))
	sendUIKey(view, tcell.KeyEnter, 0)
}

func readINI(t *testing.T, view *ui) string {
	t.Helper()
	data, err := os.ReadFile(view.paths.IniFile)
	if err != nil {
		t.Fatal(err)
	}
	return string(data)
}

func TestMainScreenPlaybackSelectionPersists(t *testing.T) {
	view, player := newWorkflowUI(t)
	labels := rowLabels(view)
	if !strings.Contains(labels[findRow(t, view, "(random)")], escapedActive) {
		t.Fatalf("random should be active: %v", labels)
	}
	if !strings.Contains(labels[findRow(t, view, "None (just top level files)")], escapedActive) {
		t.Fatalf("none should be active: %v", labels)
	}
	selectRow(t, view, "(loop)")
	if player.Playback() != bgm.PlaybackLoop {
		t.Fatalf("service playback %q", player.Playback())
	}
	if !strings.Contains(readINI(t, view), "playback = loop\n") {
		t.Fatalf("ini %q", readINI(t, view))
	}
	if !strings.Contains(rowLabels(view)[findRow(t, view, "(loop)")], escapedActive) {
		t.Fatal("loop row should now be active")
	}
	if strings.Contains(rowLabels(view)[findRow(t, view, "(random)")], escapedActive) {
		t.Fatal("random row should no longer be active")
	}
	player.StopPlaylist()
}

func TestMainScreenPlaylistSelectionPersists(t *testing.T) {
	view, player := newWorkflowUI(t)
	selectRow(t, view, "chip")
	if player.Playlist().Name() != "chip" {
		t.Fatalf("service playlist %q", player.Playlist().Name())
	}
	if !strings.Contains(readINI(t, view), "playlist = chip\n") {
		t.Fatalf("ini %q", readINI(t, view))
	}
	selectRow(t, view, "None (just top level files)")
	if !player.Playlist().IsNone() || !strings.Contains(readINI(t, view), "playlist = none\n") {
		t.Fatalf("playlist %+v ini %q", player.Playlist(), readINI(t, view))
	}
	player.StopPlaylist()
}

func TestMainScreenStartStopToggle(t *testing.T) {
	view, player := newWorkflowUI(t)
	// Like the Python menu, give the service a moment to start the track.
	view.statusDelay = 100 * time.Millisecond
	selectRow(t, view, "Start playing")
	if !player.InPlaylist() {
		t.Fatal("play should start the playlist")
	}
	view.mainList.SetCurrentItem(findRow(t, view, "Stop playing"))
	sendUIKey(view, tcell.KeyEnter, 0)
	if player.InPlaylist() {
		t.Fatal("stop should end the playlist")
	}
	findRow(t, view, "Start playing")
}

func TestRefreshKeepsCursorAndOnlyRunsWhenIdle(t *testing.T) {
	view, _ := newWorkflowUI(t)
	row := findRow(t, view, "(loop)")
	view.mainList.SetCurrentItem(row)
	current, err := view.fetchStatus(false)
	if err != nil {
		t.Fatal(err)
	}
	current.playing, current.track = true, "one.wav"
	if !view.refreshMain(current) {
		t.Fatal("idle main screen should refresh")
	}
	if got := view.mainList.GetCurrentItem(); got != row {
		t.Fatalf("cursor moved from %d to %d", row, got)
	}
	if view.statusChanged(current) {
		t.Fatal("refresh should record the new status")
	}
	findRow(t, view, "Stop playing")

	sendUIKey(view, tcell.KeyRight, 0) // highlight Settings
	current.track = "two.wav"
	if !view.refreshMain(current) || view.mainBar.FocusedIndex() != 1 {
		t.Fatalf("refresh should keep the highlighted button, got %d", view.mainBar.FocusedIndex())
	}
	sendUIKey(view, tcell.KeyEnter, 0)
	if name, _ := view.pages.GetFrontPage(); name != pageSettings {
		t.Fatalf("Enter should open the highlighted Settings button, on %s", name)
	}
	if view.refreshMain(current) {
		t.Fatal("refresh must not run on the Settings page")
	}
	sendUIKey(view, tcell.KeyEscape, 0)

	view.showError(errors.New("busy"), func() {})
	if view.refreshMain(current) {
		t.Fatal("refresh must not run under a modal")
	}
}

func TestRefreshDiscardsStatusFetchedBeforeACommand(t *testing.T) {
	view, player := newWorkflowUI(t)
	revision := view.statusRevision()
	stale, err := view.fetchStatus(false)
	if err != nil {
		t.Fatal(err)
	}
	selectRow(t, view, "(loop)") // rebuilds with newer status
	if view.applyRefresh(revision, stale) {
		t.Fatal("a status fetched before the command must be discarded")
	}
	if !strings.Contains(rowLabels(view)[findRow(t, view, "(loop)")], escapedActive) {
		t.Fatal("the command's status must remain on screen")
	}
	fresh, err := view.fetchStatus(false)
	if err != nil {
		t.Fatal(err)
	}
	fresh.track = "one.wav"
	if !view.applyRefresh(view.statusRevision(), fresh) {
		t.Fatal("a status fetched after the command must apply")
	}
	player.StopPlaylist()
}

func TestStartupDelaySettingsAreStagedAndSaved(t *testing.T) {
	view, _ := newWorkflowUI(t)
	page := newSettingsPage(view)
	page.show(0)
	page.staged.BootDelay = 1.25
	if view.cfg.BootDelay != 0 {
		t.Fatal("unsaved delay applied")
	}
	page.back()
	if !view.pages.HasPage("tui_confirm_modal") {
		t.Fatal("discard confirmation missing")
	}
	if strings.Contains(readINI(t, view), "bootdelay = 1.25") {
		t.Fatal("unsaved delay persisted")
	}
	page.show(0)
	page.save()
	if view.cfg.BootDelay != 1.25 || !strings.Contains(readINI(t, view), "bootdelay = 1.25\n") {
		t.Fatal("saved startup delay not applied and persisted")
	}
}

func TestBootRotationSettingsAreStagedAndApplied(t *testing.T) {
	view, player := newWorkflowUI(t)
	view.startSettings()
	sendUIKey(view, tcell.KeyDown, 0)
	sendUIKey(view, tcell.KeyDown, 0) // Boot sounds in rotation
	sendUIKey(view, tcell.KeyEnter, 0)
	if player.BootInPlaylist() || view.cfg.BootInPlaylist {
		t.Fatal("unsaved setting applied")
	}
	sendUIKey(view, tcell.KeyRight, 0) // Save
	sendUIKey(view, tcell.KeyEnter, 0)
	if !player.BootInPlaylist() || !view.cfg.BootInPlaylist {
		t.Fatal("saved setting not applied")
	}
	if !strings.Contains(readINI(t, view), "bootinplaylist = yes\n") {
		t.Fatal("setting not persisted")
	}
	view.startSettings()
	sendUIKey(view, tcell.KeyDown, 0)
	sendUIKey(view, tcell.KeyDown, 0)
	sendUIKey(view, tcell.KeyEnter, 0)
	sendUIKey(view, tcell.KeyRight, 0)
	sendUIKey(view, tcell.KeyEnter, 0)
	if player.BootInPlaylist() || view.cfg.BootInPlaylist {
		t.Fatal("disable not applied")
	}
}

func TestSettingsSaveWritesINIAndUpdatesService(t *testing.T) {
	view, player := newWorkflowUI(t)
	view.startSettings()
	view.pages.HasPage(pageSettings)
	sendUIKey(view, tcell.KeyDown, 0) // Play music in cores
	sendUIKey(view, tcell.KeyEnter, 0)
	sendUIKey(view, tcell.KeyRight, 0) // Save
	sendUIKey(view, tcell.KeyEnter, 0)
	if name, _ := view.pages.GetFrontPage(); name != pageMain {
		t.Fatalf("expected to return to the main page, on %s", name)
	}
	ini := readINI(t, view)
	if !strings.Contains(ini, "playincore = yes\n") || !strings.Contains(ini, "[tui]\ntheme = default\n") {
		t.Fatalf("ini %q", ini)
	}
	if !player.PlayInCore() {
		t.Fatal("service should receive set playincore yes")
	}
	if !view.cfg.PlayInCore {
		t.Fatal("ui config should be updated")
	}
}

func TestSettingsSaveKeepsConfigWhenServiceUnreachable(t *testing.T) {
	view, _ := newWorkflowUI(t)
	realSend := view.send
	view.send = func(message string) (string, bool, error) {
		if strings.HasPrefix(message, "set playincore") {
			return "", false, errors.New("service gone")
		}
		return realSend(message)
	}
	view.startSettings()
	sendUIKey(view, tcell.KeyDown, 0) // Play music in cores
	sendUIKey(view, tcell.KeyEnter, 0)
	sendUIKey(view, tcell.KeyRight, 0) // Save
	sendUIKey(view, tcell.KeyEnter, 0)
	if !view.pages.HasPage("tui_error_modal") {
		t.Fatal("expected the socket error to be shown")
	}
	if !view.cfg.PlayInCore || !strings.Contains(readINI(t, view), "playincore = yes\n") {
		t.Fatalf("saved settings must be applied: cfg=%v ini=%q", view.cfg.PlayInCore, readINI(t, view))
	}
}

func TestSettingsCancelAsksBeforeDiscarding(t *testing.T) {
	view, _ := newWorkflowUI(t)
	view.startSettings()
	sendUIKey(view, tcell.KeyEnter, 0) // toggle Start on boot
	sendUIKey(view, tcell.KeyEscape, 0)
	if !view.pages.HasPage("tui_confirm_modal") {
		t.Fatal("expected a discard confirmation")
	}
	if strings.Contains(readINI(t, view), "startup = no") {
		t.Fatal("cancelled settings must not be written")
	}
}

func TestSettingsMenuVolumeAppliesLive(t *testing.T) {
	view, _ := newWorkflowUI(t)
	page := newSettingsPage(view)
	page.show(0)
	var index int
	for candidate, help := range page.help {
		if help == settingsHelp["Menu volume"] {
			index = candidate
		}
	}
	page.list.SetCurrentItem(index)
	sendUIKey(view, tcell.KeyEnter, 0)
	if !view.pages.HasPage("tui_choice_modal") {
		t.Fatal("expected the volume picker")
	}
	for range 4 { // Disabled -> Mute -> 1 -> 2 -> 3
		sendUIKey(view, tcell.KeyDown, 0)
	}
	sendUIKey(view, tcell.KeyEnter, 0)
	if page.staged.MenuVolume != 3 {
		t.Fatalf("staged menu volume %d", page.staged.MenuVolume)
	}
	data, err := os.ReadFile(view.paths.CmdInterface)
	if err != nil || string(data) != "volume 3\n" {
		t.Fatalf("live volume command %q %v", data, err)
	}
}

func TestEverySettingHasHelpAndRenders(t *testing.T) {
	view, _ := newWorkflowUI(t)
	page := newSettingsPage(view)
	page.show(0)
	for index := range page.actions {
		if help := page.help[index]; help == "" || len(help) > 73 {
			t.Fatalf("missing or overflowing help for row %d: %q", index, help)
		}
	}
	for _, size := range [][2]int{{75, 15}, {100, 30}} {
		for _, name := range []string{pageSettings, pageMain} {
			view.pages.SwitchToPage(name)
			_, primitive := view.pages.GetFrontPage()
			screen := tcell.NewSimulationScreen("UTF-8")
			if err := screen.Init(); err != nil {
				t.Fatal(err)
			}
			screen.SetSize(size[0], size[1])
			primitive.SetRect(0, 0, size[0], size[1])
			primitive.Draw(screen)
			screen.Fini()
		}
	}
}

func TestTitleCase(t *testing.T) {
	cases := map[string]string{"random": "Random", "foo_bar": "Foo_Bar", "abc3def": "Abc3Def", "LOOP": "Loop", "": ""}
	for input, want := range cases {
		if got := titleCase(input); got != want {
			t.Fatalf("titleCase(%q) = %q want %q", input, got, want)
		}
	}
}

func newSettingsPage(view *ui) *settingsPage {
	return &settingsPage{
		staged:   bgm.SettingsFromConfig(&view.cfg),
		original: bgm.SettingsFromConfig(&view.cfg),
		ui:       view,
	}
}
