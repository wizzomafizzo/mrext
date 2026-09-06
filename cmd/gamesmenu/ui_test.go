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

//nolint:gosec // Tests only operate on temporary fixture paths.
package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	"github.com/wizzomafizzo/mrext/pkg/gamesmenu"
	"github.com/wizzomafizzo/mrext/pkg/tui"
)

func newWorkflowUI(t *testing.T, menuExists bool) *ui {
	t.Helper()
	cfg, manager, err := loadRuntime(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	for _, system := range []string{"NES", "SNES"} {
		folder := filepath.Join(manager.Paths().SDRoot, "games", system)
		if mkdirErr := os.MkdirAll(folder, 0o755); mkdirErr != nil {
			t.Fatal(mkdirErr)
		}
		if menuExists {
			if mkdirErr := os.MkdirAll(filepath.Join(manager.Paths().MenuFolder, "_"+system), 0o755); mkdirErr != nil {
				t.Fatal(mkdirErr)
			}
		}
	}
	if configErr := gamesmenu.EnsureConfigFile(cfg); configErr != nil {
		t.Fatal(configErr)
	}
	app, err := tui.NewApplication(tui.ApplicationOptions{Theme: "default"})
	if err != nil {
		t.Fatal(err)
	}
	view := newUI(cfg, manager)
	view.app, view.pages = app, tview.NewPages()
	if err := view.showMain(); err != nil {
		t.Fatal(err)
	}
	app.SetRoot(view.pages, true)
	t.Cleanup(func() { _ = tui.SetCurrentTheme("default") })
	return view
}

func sendUIKey(view *ui, key tcell.Key) {
	view.app.GetFocus().InputHandler()(tcell.NewEventKey(key, 0, tcell.ModNone), func(p tview.Primitive) {
		view.app.SetFocus(p)
	})
}

func drawText(t *testing.T, view *ui, width, height int) string {
	t.Helper()
	screen := tcell.NewSimulationScreen("UTF-8")
	if err := screen.Init(); err != nil {
		t.Fatal(err)
	}
	defer screen.Fini()
	screen.SetSize(width, height)
	view.pages.SetRect(0, 0, width, height)
	view.pages.Draw(screen)
	screen.Show()
	var text strings.Builder
	for y := range height {
		for x := range width {
			ch, _, _ := screen.Get(x, y)
			_, _ = text.WriteString(ch)
		}
		_ = text.WriteByte('\n')
	}
	return text.String()
}

func TestMainPageRendersAndPreservesToggleFocus(t *testing.T) {
	view := newWorkflowUI(t, false)
	for _, size := range [][2]int{{75, 15}, {100, 30}} {
		text := drawText(t, view, size[0], size[1])
		for _, label := range []string{"[x] NES", "[x] SNES", "Toggle", "Generate", "Clean Up", "Settings", "Exit"} {
			if !strings.Contains(text, label) {
				t.Errorf("missing %q:\n%s", label, text)
			}
		}
	}
	sendUIKey(view, tcell.KeyDown)
	sendUIKey(view, tcell.KeyEnter)
	if view.entries[1].Selected || view.mainSelection != 1 {
		t.Fatal("toggle lost selection")
	}
	if text := drawText(t, view, 75, 15); !strings.Contains(text, "[ ] SNES") {
		t.Fatal(text)
	}
	_, primitive := view.pages.GetFrontPage()
	frame, ok := primitive.(*tui.PageFrame)
	if !ok {
		t.Fatalf("page: %T", primitive)
	}
	frame.FocusButtonBar()
	bar, ok := view.app.GetFocus().(*tui.ButtonBar)
	if !ok {
		t.Fatalf("focus: %T", view.app.GetFocus())
	}
	bar.SetFocusedIndex(0)
	sendUIKey(view, tcell.KeyEnter)
	if !view.entries[1].Selected {
		t.Fatal("button did not toggle")
	}
	if bar, ok := view.app.GetFocus().(*tui.ButtonBar); !ok || bar.FocusedIndex() != 0 {
		t.Fatal("button focus lost")
	}
}

func TestWelcomeOnlyWhenMenuMissing(t *testing.T) {
	for _, exists := range []bool{false, true} {
		view := newWorkflowUI(t, exists)
		view.showWelcome()
		if view.pages.HasPage("tui_info_modal") == exists {
			t.Fatalf("welcome for exists=%v", exists)
		}
	}
}

func TestGenerateRemovalConfirmationAndCancel(t *testing.T) {
	view := newWorkflowUI(t, true)
	view.entries[1].Selected = false
	view.startGenerate()
	if !view.pages.HasPage("tui_confirm_modal") {
		t.Fatal("missing confirmation")
	}
	if text := drawText(t, view, 100, 30); !strings.Contains(text, "_SNES") {
		t.Fatal(text)
	}
	sendUIKey(view, tcell.KeyRight)
	sendUIKey(view, tcell.KeyEnter)
	if name, _ := view.pages.GetFrontPage(); name != pageMain {
		t.Fatal(name)
	}
	if view.entries[1].Selected {
		t.Fatal("cancel lost staged toggles")
	}
	view.entries[0].Selected = false
	view.startGenerate()
	if text := drawText(t, view, 100, 30); !strings.Contains(text, "Remove the Games menu") {
		t.Fatal(text)
	}
	sendUIKey(view, tcell.KeyEnter)
	if view.manager.MenuExists() {
		t.Fatal("menu not removed")
	}
	sendUIKey(view, tcell.KeyEnter)
	for _, entry := range view.entries {
		if !entry.Selected {
			t.Fatal("removal did not restore first-run state")
		}
	}
}

func TestCleanUpWithoutMenuShowsInfo(t *testing.T) {
	view := newWorkflowUI(t, false)
	view.startCleanUp()
	if !view.pages.HasPage("tui_info_modal") {
		t.Fatal("missing info")
	}
	if text := drawText(t, view, 75, 15); !strings.Contains(text, "No Games menu exists") {
		t.Fatal(text)
	}
}

func TestRemovalPreviewScrollsWithoutConfirming(t *testing.T) {
	view := newWorkflowUI(t, true)
	for index := 1; index <= 20; index++ {
		folder := filepath.Join(view.manager.Paths().MenuFolder, fmt.Sprintf("_User%02d", index))
		if err := os.Mkdir(folder, 0o755); err != nil {
			t.Fatal(err)
		}
	}
	view.startGenerate()
	text := drawText(t, view, 75, 15)
	if !strings.Contains(text, "20 folders") || !strings.Contains(text, "scroll") {
		t.Fatalf("missing removal count or scrolling help: %s", text)
	}
	for range 30 {
		sendUIKey(view, tcell.KeyDown)
		text = drawText(t, view, 75, 15)
	}
	if !strings.Contains(text, "_User20") {
		t.Fatalf("last removal is inaccessible: %s", text)
	}
	if name, _ := view.pages.GetFrontPage(); name != "tui_confirm_modal" {
		t.Fatalf("scrolling dismissed confirmation: %s", name)
	}
	if _, err := os.Stat(filepath.Join(view.manager.Paths().MenuFolder, "_User20")); err != nil {
		t.Fatalf("scrolling removed a folder: %v", err)
	}
	sendUIKey(view, tcell.KeyRight)
	sendUIKey(view, tcell.KeyEnter)
	if name, _ := view.pages.GetFrontPage(); name != pageMain {
		t.Fatalf("No did not return to main: %s", name)
	}
	if _, err := os.Stat(filepath.Join(view.manager.Paths().MenuFolder, "_User20")); err != nil {
		t.Fatalf("No removed a folder: %v", err)
	}
}

func TestSettingsStageSaveHelpAndDiscard(t *testing.T) {
	view := newWorkflowUI(t, false)
	settings := newSettingsPage(view)
	settings.show(1)
	if len(settings.actions) != 4 {
		t.Fatalf("rows: %d", len(settings.actions))
	}
	for index, help := range settings.help {
		if help == "" || len(help) > 73 {
			t.Errorf("row %d help: %q", index, help)
		}
	}
	settings.actions[2]()
	if !view.cfg.TUI.Mouse || settings.staged.Mouse {
		t.Fatal("setting applied before Save")
	}
	settings.back()
	if !view.pages.HasPage("tui_confirm_modal") {
		t.Fatal("discard confirmation missing")
	}
	sendUIKey(view, tcell.KeyRight)
	sendUIKey(view, tcell.KeyEnter)
	settings.staged.Theme = "dracula"
	settings.save()
	if view.cfg.TUI.Mouse || view.cfg.TUI.Theme != "dracula" {
		t.Fatal("settings not applied")
	}
	cfg, err := gamesmenu.LoadConfigAt(view.cfg.IniPath, view.cfg.AppPath)
	if err != nil {
		t.Fatal(err)
	}
	if cfg.TUI.Mouse || cfg.TUI.Theme != "dracula" {
		t.Fatal("settings not persisted")
	}
	settings = newSettingsPage(view)
	settings.show(2)
	settings.actions[2]()
	settings.back()
	sendUIKey(view, tcell.KeyEnter)
	if view.cfg.TUI.Mouse {
		t.Fatal("discard applied setting")
	}
}

func TestGenerateAndCleanupProgressWorkflow(t *testing.T) {
	view := newWorkflowUI(t, false)
	game := filepath.Join(view.manager.Paths().SDRoot, "games", "NES", "game.nes")
	if err := os.WriteFile(game, nil, 0o644); err != nil {
		t.Fatal(err)
	}
	screen := tcell.NewSimulationScreen("UTF-8")
	view.app.SetScreen(screen)
	screen.SetSize(75, 15)
	done := make(chan error, 1)
	go func() { done <- view.app.Run() }()
	t.Cleanup(func() {
		view.app.Stop()
		select {
		case err := <-done:
			if err != nil {
				t.Error(err)
			}
		case <-time.After(3 * time.Second):
			t.Error("UI did not stop")
		}
	})
	view.app.QueueUpdateDraw(view.startGenerate)
	waitForInfo := func() {
		t.Helper()
		deadline := time.Now().Add(3 * time.Second)
		for time.Now().Before(deadline) {
			ready := false
			view.app.QueueUpdate(func() { ready = view.pages.HasPage("tui_info_modal") })
			if ready {
				return
			}
			time.Sleep(time.Millisecond)
		}
		t.Fatal("progress did not complete")
	}
	waitForInfo()
	shortcut := filepath.Join(view.manager.Paths().MenuFolder, "_NES", "game.mgl")
	if _, err := os.Stat(shortcut); err != nil {
		t.Fatal(err)
	}
	view.app.QueueUpdateDraw(func() { sendUIKey(view, tcell.KeyEnter) })
	if err := os.Remove(game); err != nil {
		t.Fatal(err)
	}
	view.app.QueueUpdateDraw(view.cleanUp)
	waitForInfo()
	if _, err := os.Stat(shortcut); !os.IsNotExist(err) {
		t.Fatalf("shortcut remains: %v", err)
	}
}
