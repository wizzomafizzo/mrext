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

//nolint:gosec // Tests operate only on paths rooted in t.TempDir.
package main

import (
	"archive/zip"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	"github.com/wizzomafizzo/mrext/pkg/favorites"
	"github.com/wizzomafizzo/mrext/pkg/tui"
)

func sendUIKey(view *ui, key tcell.Key, value rune) {
	view.app.GetFocus().InputHandler()(tcell.NewEventKey(key, value, tcell.ModNone), func(p tview.Primitive) {
		view.app.SetFocus(p)
	})
}

func newWorkflowUI(t *testing.T) *ui {
	t.Helper()
	cfg, manager, err := loadRuntime(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	if _, createErr := manager.CreateDefaultFolder(); createErr != nil {
		t.Fatal(createErr)
	}
	app, err := tui.NewApplication(tui.ApplicationOptions{Theme: "default"})
	if err != nil {
		t.Fatal(err)
	}
	view := newUI(cfg, manager)
	view.app = app
	view.pages = tview.NewPages()
	return view
}

func submitKeyboard(view *ui) {
	sendUIKey(view, tcell.KeyEnter, 0)
}

func TestCreateDestinationDuringAddAndMove(t *testing.T) {
	for _, move := range []bool{false, true} {
		view := newWorkflowUI(t)
		source := filepath.Join(view.manager.Root(), "Core.rbf")
		if err := os.WriteFile(source, nil, 0o600); err != nil {
			t.Fatal(err)
		}
		if move {
			favorite, err := view.manager.CreateCoreFavorite(source, view.manager.Root(), "Favorite")
			if err != nil {
				t.Fatal(err)
			}
			view.moveItem(&favorites.Favorite{Path: favorite, Kind: favorites.FavoriteCore})
		} else {
			view.startAddFavorite(favorites.BrowseEntry{Path: source})
		}
		// Create Folder, then cancel its parent picker and resume pending operation.
		sendUIKey(view, tcell.KeyRight, 0)
		sendUIKey(view, tcell.KeyEnter, 0)
		sendUIKey(view, tcell.KeyEscape, 0)
		sendUIKey(view, tcell.KeyRight, 0)
		sendUIKey(view, tcell.KeyEnter, 0)
		sendUIKey(view, tcell.KeyEnter, 0)
		for _, char := range "Nested" {
			sendUIKey(view, tcell.KeyRune, char)
		}
		submitKeyboard(view)
		sendUIKey(view, tcell.KeyEnd, 0)
		sendUIKey(view, tcell.KeyEnter, 0)
		name := "Favorite.rbf"
		if !move {
			submitKeyboard(view)
			name = "Core.rbf"
		}
		path := filepath.Join(view.manager.Root(), "_@Favorites", "_Nested", name)
		if _, err := os.Lstat(path); err != nil {
			t.Fatalf("move=%v pending operation not retained: %v", move, err)
		}
	}
}

func TestCancelAddInsideArchiveReturnsToBrowser(t *testing.T) {
	view := newWorkflowUI(t)
	archive := filepath.Join(view.manager.Root(), "games.zip")
	file, err := os.Create(archive)
	if err != nil {
		t.Fatal(err)
	}
	writer := zip.NewWriter(file)
	if _, createErr := writer.Create("game.sfc"); createErr != nil {
		t.Fatal(createErr)
	}
	if closeErr := writer.Close(); closeErr != nil {
		t.Fatal(closeErr)
	}
	if closeErr := file.Close(); closeErr != nil {
		t.Fatal(closeErr)
	}
	for _, cancelName := range []bool{false, true} {
		view.startAddFavorite(favorites.BrowseEntry{Path: archive + "/game.sfc"})
		if cancelName {
			sendUIKey(view, tcell.KeyEnter, 0)
		}
		sendUIKey(view, tcell.KeyEscape, 0)
		page, _ := view.pages.GetFrontPage()
		if page != pageBrowser || view.pages.HasPage("tui_error_modal") {
			t.Fatalf("cancelName=%v returned to %s", cancelName, page)
		}
	}
}

func TestEverySettingHasContextualHelp(t *testing.T) {
	view := newWorkflowUI(t)
	page := &settingsPage{
		staged: favorites.SettingsFromConfig(view.cfg), original: favorites.SettingsFromConfig(view.cfg), ui: view,
	}
	page.show(0)
	for index := range page.actions {
		help := page.help[index]
		if help == "" || len(help) > 73 {
			t.Fatalf("missing or overflowing help for row %d: %#v", index, help)
		}
	}
	_, primitive := view.pages.GetFrontPage()
	frame, ok := primitive.(*tui.PageFrame)
	if !ok {
		t.Fatal("settings frame missing")
	}
	screen := tcell.NewSimulationScreen("UTF-8")
	if err := screen.Init(); err != nil {
		t.Fatal(err)
	}
	defer screen.Fini()
	screen.SetSize(75, 15)
	frame.SetRect(0, 0, 75, 15)
	sendUIKey(view, tcell.KeyDown, 0)
	frame.Draw(screen)
	var helpText strings.Builder
	for x := range 75 {
		value, _, _ := screen.Get(x, 12)
		_, _ = helpText.WriteString(value)
	}
	if !strings.Contains(helpText.String(), page.help[2]) {
		t.Fatalf("help did not follow selection: %q", helpText.String())
	}
	var buttons strings.Builder
	for x := range 75 {
		value, _, _ := screen.Get(x, 13)
		_, _ = buttons.WriteString(value)
	}
	if strings.Contains(buttons.String(), "Help") {
		t.Fatal("settings should have footer help, not a Help button")
	}
}

func TestListSettingsEditOneValueAtATime(t *testing.T) {
	for _, cancel := range []bool{false, true} {
		view := newWorkflowUI(t)
		page := &settingsPage{
			staged: favorites.SettingsFromConfig(view.cfg), original: favorites.SettingsFromConfig(view.cfg), ui: view,
		}
		page.show(0)
		page.actions[2]()
		sendUIKey(view, tcell.KeyRight, 0)
		sendUIKey(view, tcell.KeyEnter, 0)
		for _, char := range "collection" {
			sendUIKey(view, tcell.KeyRune, char)
		}
		submitKeyboard(view)
		if cancel {
			sendUIKey(view, tcell.KeyEscape, 0)
		} else {
			for range 3 {
				sendUIKey(view, tcell.KeyRight, 0)
			}
			sendUIKey(view, tcell.KeyEnter, 0)
		}
		expected := []string{"fav"}
		if !cancel {
			expected = append(expected, "collection")
		}
		if !slices.Equal(page.staged.FolderNameContains, expected) {
			t.Fatalf("cancel=%v staged list=%v", cancel, page.staged.FolderNameContains)
		}
		if !slices.Equal(view.cfg.Favorites.FolderNameContains, []string{"fav"}) {
			t.Fatal("list editor bypassed staged settings")
		}
	}
}

func TestLocalRootRuntime(t *testing.T) {
	root := filepath.Join(t.TempDir(), "fixture")
	options, err := parseCLI([]string{"--root", root, "refresh"})
	if err != nil {
		t.Fatal(err)
	}
	if options.root != root || options.command != "refresh" {
		t.Fatalf("options = %#v", options)
	}
	cfg, manager, err := loadRuntime(root)
	if err != nil {
		t.Fatal(err)
	}
	absoluteRoot, err := filepath.Abs(root)
	if err != nil {
		t.Fatal(err)
	}
	if manager.Root() != absoluteRoot {
		t.Fatalf("manager root = %q", manager.Root())
	}
	if cfg.IniPath != filepath.Join(absoluteRoot, "Scripts", "favorites.ini") {
		t.Fatalf("config path = %q", cfg.IniPath)
	}
	if cfg.AppPath != filepath.Join(absoluteRoot, "Scripts", "favorites.sh") {
		t.Fatalf("app path = %q", cfg.AppPath)
	}
	if err := run([]string{"--root", root, "refresh"}); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(cfg.IniPath); !os.IsNotExist(err) {
		t.Fatalf("refresh unexpectedly created configuration: %v", err)
	}
}

func TestMainLabelTruncationPreservesFilename(t *testing.T) {
	value := "_@Favorites/_Role Playing Games/_Translated/Very Long Favorite Name"
	got := truncateLeading(value, 30)
	if got != "...ted/Very Long Favorite Name" {
		t.Fatalf("truncated label = %q", got)
	}
	if got := truncateLeading("short", 30); got != "short" {
		t.Fatalf("short label = %q", got)
	}
}

func TestInvalidCLICommand(t *testing.T) {
	if _, err := parseCLI([]string{"invalid"}); err == nil {
		t.Fatal("invalid command accepted")
	}
}

func TestSettingsPageStagesAndSavesChanges(t *testing.T) {
	root := filepath.Join(t.TempDir(), "fat")
	for _, folder := range []string{
		filepath.Join(root, "Scripts"),
		filepath.Join(root, "_Arcade", "cores"),
	} {
		if err := os.MkdirAll(folder, 0o750); err != nil {
			t.Fatal(err)
		}
	}
	cfg := favorites.DefaultUserConfig()
	cfg.AppPath = filepath.Join(root, "Scripts", "favorites.sh")
	cfg.IniPath = filepath.Join(root, "Scripts", "favorites.ini")
	cfg.Systems.GamesFolder = []string{root}
	if err := os.WriteFile(cfg.IniPath, []byte(favorites.DefaultConfigFile), 0o600); err != nil {
		t.Fatal(err)
	}
	manager := favorites.NewManagerWithPaths(cfg, favorites.RuntimePaths{
		SDRoot:            root,
		StartupScript:     filepath.Join(root, "linux", "user-startup.sh"),
		ArcadeCoresFolder: filepath.Join(root, "_Arcade", "cores"),
	})
	application, err := tui.NewApplication(tui.ApplicationOptions{Theme: "default"})
	if err != nil {
		t.Fatal(err)
	}
	view := newUI(cfg, manager)
	view.app = application
	view.pages = tview.NewPages()
	page := &settingsPage{
		staged:   favorites.SettingsFromConfig(cfg),
		original: favorites.SettingsFromConfig(cfg),
		ui:       view,
	}
	page.show(0)
	if !view.pages.HasPage(pageSettings) || page.list.GetItemCount() != 16 {
		t.Fatalf("settings page rows = %d", page.list.GetItemCount())
	}
	screen := tcell.NewSimulationScreen("UTF-8")
	if initErr := screen.Init(); initErr != nil {
		t.Fatal(initErr)
	}
	screen.SetSize(75, 15)
	view.pages.SetRect(0, 0, 75, 15)
	view.pages.Draw(screen)
	screen.Fini()
	page.actions[5]()
	page.actions[10]()
	sendUIKey(view, tcell.KeyEnd, 0)
	sendUIKey(view, tcell.KeyEnter, 0)
	page.actions[12]()
	sendUIKey(view, tcell.KeyDown, 0)
	sendUIKey(view, tcell.KeyEnter, 0)
	if page.staged.HideRootFiles || page.staged.AlternateCore != "ra" || page.staged.Theme != "high_contrast" {
		t.Fatalf("staged settings = %#v", page.staged)
	}
	page.save()
	if cfg.Favorites.HideRootFiles || cfg.FavoritesCores.All != "ra" || cfg.TUI.Theme != "high_contrast" {
		t.Fatalf("applied settings = %#v", favorites.SettingsFromConfig(cfg))
	}
	loaded, err := favorites.LoadConfigAt(cfg.IniPath, cfg.AppPath)
	if err != nil {
		t.Fatal(err)
	}
	if loaded.Favorites.HideRootFiles || loaded.FavoritesCores.All != "ra" || loaded.TUI.Theme != "high_contrast" {
		t.Fatalf("saved settings = %#v", favorites.SettingsFromConfig(loaded))
	}
}

func TestSettingsBackConfirmsDiscard(t *testing.T) {
	cfg := favorites.DefaultUserConfig()
	application, err := tui.NewApplication(tui.ApplicationOptions{Theme: "default"})
	if err != nil {
		t.Fatal(err)
	}
	view := newUI(cfg, favorites.NewManager(cfg))
	view.app = application
	view.pages = tview.NewPages()
	page := &settingsPage{
		staged:   favorites.SettingsFromConfig(cfg),
		original: favorites.SettingsFromConfig(cfg),
		ui:       view,
	}
	page.show(0)
	page.staged.Mouse = !page.staged.Mouse
	page.back()
	if !view.pages.HasPage("tui_confirm_modal") {
		t.Fatal("discard confirmation was not shown")
	}
}

func TestMainAndBrowserPagesBuildWithoutDialogProcess(t *testing.T) {
	root := filepath.Join(t.TempDir(), "fat")
	for _, folder := range []string{
		root,
		filepath.Join(root, "games", "SNES"),
		filepath.Join(root, "_@Favorites"),
		filepath.Join(root, "_Arcade", "cores"),
	} {
		if err := os.MkdirAll(folder, 0o750); err != nil {
			t.Fatal(err)
		}
	}
	if err := os.WriteFile(filepath.Join(root, "games", "SNES", "EarthBound.sfc"), nil, 0o600); err != nil {
		t.Fatal(err)
	}
	cfg := favorites.DefaultUserConfig()
	cfg.Systems.GamesFolder = []string{root}
	manager := favorites.NewManagerWithPaths(cfg, favorites.RuntimePaths{
		SDRoot:            root,
		StartupScript:     filepath.Join(root, "linux", "user-startup.sh"),
		ArcadeCoresFolder: filepath.Join(root, "_Arcade", "cores"),
	})
	application, err := tui.NewApplication(tui.ApplicationOptions{Theme: "default"})
	if err != nil {
		t.Fatal(err)
	}
	view := newUI(cfg, manager)
	view.app = application
	view.pages = tview.NewPages()
	if err := view.showMain(); err != nil {
		t.Fatal(err)
	}
	if !view.pages.HasPage(pageMain) {
		t.Fatal("main page was not created")
	}
	view.showBrowser(root)
	if !view.pages.HasPage(pageBrowser) {
		t.Fatal("browser page was not created")
	}
	screen := tcell.NewSimulationScreen("UTF-8")
	if err := screen.Init(); err != nil {
		t.Fatal(err)
	}
	defer screen.Fini()
	screen.SetSize(100, 30)
	view.pages.SetRect(0, 0, 100, 30)
	view.pages.Draw(screen)
}
