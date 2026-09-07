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
	"fmt"

	"github.com/rivo/tview"
	"github.com/wizzomafizzo/mrext/pkg/config"
	"github.com/wizzomafizzo/mrext/pkg/games"
	"github.com/wizzomafizzo/mrext/pkg/gamesdb"
	"github.com/wizzomafizzo/mrext/pkg/mister"
	"github.com/wizzomafizzo/mrext/pkg/tui"
	"github.com/wizzomafizzo/mrext/pkg/utils"
)

const (
	appTitle    = "Search"
	pageMain    = "search_main"
	pageOptions = "search_options"
)

// resultsPageSize matches the old ListPicker's paging, which stepped by the
// window height less its chrome.
const resultsPageSize = 13

type ui struct {
	app      *tview.Application
	pages    *tview.Pages
	cfg      *config.UserConfig
	list     *tui.MenuList
	query    string
	results  []gamesdb.SearchResult
	names    []string
	selected *gamesdb.SearchResult
	options  tui.ApplicationOptions
	launch   bool
}

func newUI(cfg *config.UserConfig, launch bool) *ui {
	return &ui{
		cfg:    cfg,
		launch: launch,
		options: tui.ApplicationOptions{
			Theme:            cfg.TUI.Theme,
			Mouse:            cfg.TUI.Mouse,
			CRTMode:          cfg.TUI.CRTMode,
			OnScreenKeyboard: cfg.TUI.OnScreenKeyboard,
		},
	}
}

func (u *ui) Run() error {
	builder := func() (*tview.Application, error) {
		app, err := tui.NewApplication(u.options)
		if err != nil {
			return nil, fmt.Errorf("create TUI application: %w", err)
		}
		u.app, u.pages = app, tview.NewPages()
		u.results, u.names, u.query = nil, nil, ""
		u.showMain()
		app.SetRoot(tui.WrapRoot(u.options, u.pages), true)
		// Typing is still the first thing that happens, as it was when the
		// keyboard was the whole first screen.
		u.startIndexThenSearch()
		return app, nil
	}
	if err := tui.BuildAndRetry(builder); err != nil {
		return fmt.Errorf("run Search TUI: %w", err)
	}
	return nil
}

// startIndexThenSearch builds the index first when there is none, exactly as
// the previous first-launch flow did, then opens the keyboard.
func (u *ui) startIndexThenSearch() {
	if gamesdb.DBExists() {
		u.startSearchInput()
		return
	}
	u.rebuildIndex(u.startSearchInput)
}

func (u *ui) showMain() {
	u.list = tui.NewMenuList()
	if len(u.names) == 0 {
		u.list.AddHeader("No results")
	}
	for _, name := range u.names {
		u.list.AddRow(name, tui.MenuRowItem)
	}

	launchLabel := "Launch"
	if !u.launch {
		launchLabel = "Select"
	}
	activate := func() { u.chooseResult() }
	bar := tui.NewButtonBar(u.app).
		AddButtonWithHelp(launchLabel, "Launch the selected game", activate).
		AddButtonWithHelp("Search", "Enter a new search", u.startSearchInput).
		AddButtonWithHelp("PgUp", "Jump up a page of results", func() { u.page(-resultsPageSize) }).
		AddButtonWithHelp("PgDn", "Jump down a page of results", func() { u.page(resultsPageSize) }).
		AddButtonWithHelp("Options", "Update the games database", u.showOptions).
		AddButtonWithHelp("Exit", "Exit Search", u.app.Stop).
		SetupNavigation(u.app.Stop)
	u.list.SetSelectedFunc(func(int, string, string, rune) { activate() })

	frame := tui.NewPageFrame(u.app).
		SetTitle(appTitle).
		SetContent(u.list).
		SetHelpText(u.status()).
		SetButtonBar(bar).
		SetOnEscape(u.app.Stop)
	u.list.SetRowChangedFunc(func(int) { frame.SetHelpText(u.status()) })
	bar.SetHelpCallback(func(text string) { frame.SetHelpText(text) })
	frame.SetupContentToButtonNavigation()
	u.pages.AddAndSwitchToPage(pageMain, frame, true)
	u.app.SetFocus(u.list)
}

// status is the footer line: the query and the position in its results, which
// is what the old picker's title and total counter carried.
func (u *ui) status() string {
	if u.query == "" {
		return "Press Search to look for a game."
	}
	if len(u.names) == 0 {
		return fmt.Sprintf("No results for %q.", u.query)
	}
	return fmt.Sprintf("%q: %d/%d", u.query, u.list.GetCurrentItem()+1, len(u.names))
}

func (u *ui) page(delta int) {
	if len(u.names) == 0 {
		return
	}
	target := min(max(u.list.GetCurrentItem()+delta, 0), len(u.names)-1)
	u.list.SetCurrentItem(target)
}

func (u *ui) startSearchInput() {
	options := tui.InputOptions{
		Title:            appTitle,
		Prompt:           "Enter part of a game name.",
		InitialValue:     u.query,
		OnScreenKeyboard: u.options.OnScreenKeyboard,
	}
	tui.ShowInputModal(u.pages, u.app, options, u.runSearch, u.showMain)
}

func (u *ui) runSearch(query string) {
	if query == "" {
		u.showMain()
		return
	}
	u.query = query
	var results []gamesdb.SearchResult
	progress := tui.ProgressOptions{Initial: tui.ProgressUpdate{Text: "Searching..."}}
	tui.ShowProgressModal(u.pages, u.app, progress, func(func(tui.ProgressUpdate)) error {
		found, err := gamesdb.SearchNamesWords(games.AllSystems(), query)
		if err != nil {
			return fmt.Errorf("search game names: %w", err)
		}
		results = found
		return nil
	}, func(err error) {
		if err != nil {
			u.showError(err, u.showMain)
			return
		}
		u.setResults(results)
		u.showMain()
	})
}

// setResults keeps the old picker's de-duplication and "[System] Name" format.
func (u *ui) setResults(results []gamesdb.SearchResult) {
	u.names, u.results = nil, nil
	for _, result := range results {
		systemName := result.SystemID
		if system, err := games.GetSystem(result.SystemID); err == nil {
			systemName = system.Name
		}
		display := fmt.Sprintf("[%s] %s", systemName, result.Name)
		if utils.Contains(u.names, display) {
			continue
		}
		u.names = append(u.names, display)
		u.results = append(u.results, result)
	}
}

func (u *ui) chooseResult() {
	index := u.list.GetCurrentItem()
	if index < 0 || index >= len(u.results) {
		return
	}
	game := u.results[index]
	if !u.launch {
		u.selected = &game
		u.app.Stop()
		return
	}
	system, err := games.GetSystem(game.SystemID)
	if err != nil {
		u.showError(fmt.Errorf("get selected system: %w", err), u.showMain)
		return
	}
	if err := mister.LaunchGame(u.cfg, system, game.Path); err != nil {
		u.showError(fmt.Errorf("launch selected game: %w", err), u.showMain)
		return
	}
	u.app.Stop()
}

func (u *ui) showOptions() {
	list := tui.NewMenuList()
	list.AddRow("Update games database", tui.MenuRowAction)
	back := func() { u.showMain() }
	update := func() { u.confirmRebuild() }
	list.SetSelectedFunc(func(int, string, string, rune) { update() })
	bar := tui.NewButtonBar(u.app).
		AddButtonWithHelp("Select", "Run the selected action", update).
		AddButtonWithHelp("Back", "Return to search results", back).
		SetupNavigation(back)
	frame := tui.NewPageFrame(u.app).
		SetTitle(appTitle, "Options").
		SetContent(list).
		SetHelpText("Rescans every games folder. This can take several minutes.").
		SetButtonBar(bar).
		SetOnEscape(back)
	bar.SetHelpCallback(func(text string) { frame.SetHelpText(text) })
	frame.SetupContentToButtonNavigation()
	u.pages.AddAndSwitchToPage(pageOptions, frame, true)
	u.app.SetFocus(list)
}

func (u *ui) confirmRebuild() {
	tui.ShowConfirmModal(u.pages, u.app, "Update games database",
		"Rescan every games folder and rebuild the database?\n\nThis can take several minutes.",
		func() { u.rebuildIndex(u.showMain) },
		u.showOptions)
}

// rebuildIndex runs the scan behind a progress modal, reporting the same steps
// the standalone progress window did.
func (u *ui) rebuildIndex(onDone func()) {
	progress := tui.ProgressOptions{
		Initial: tui.ProgressUpdate{Text: "Finding games folders...", Current: 1, Total: 100},
	}
	tui.ShowProgressModal(u.pages, u.app, progress, func(update func(tui.ProgressUpdate)) error {
		_, err := gamesdb.RebuildNamesIndex(u.cfg, games.AllSystems(), func(status gamesdb.IndexStatus) {
			update(tui.ProgressUpdate{Text: indexStepText(status), Current: status.Step, Total: status.Total})
		})
		if err != nil {
			// ErrIndexBusy is already a complete sentence; do not bury it.
			return fmt.Errorf("%w", err)
		}
		return nil
	}, func(err error) {
		if err != nil {
			u.showError(err, onDone)
			return
		}
		onDone()
	})
}

func indexStepText(status gamesdb.IndexStatus) string {
	switch status.Step {
	case 1:
		return "Finding games folders..."
	case status.Total:
		return "Writing database to disk..."
	default:
		name := status.SystemID
		if system, err := games.GetSystem(status.SystemID); err == nil {
			name = system.Name
		}
		return "Indexing " + name + "..."
	}
}

func (u *ui) showError(err error, restore func()) {
	tui.ShowErrorModal(u.pages, u.app, err.Error(), restore)
}
