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
	"strings"

	"github.com/rivo/tview"
	"github.com/wizzomafizzo/mrext/pkg/config"
	"github.com/wizzomafizzo/mrext/pkg/gamesmenu"
	"github.com/wizzomafizzo/mrext/pkg/tui"
)

const pageMain = "gamesmenu_main"

const welcomeText = "This script will generate a new folder in your main menu called Games, " +
	"which will allow you to directly launch games without first opening a core. " +
	"This process will take up a lot of storage space if you have large game sets. " +
	"Press the Generate button to continue."

//nolint:govet // Groups runtime dependencies and navigation state.
type ui struct {
	manager                   *gamesmenu.Manager
	cfg                       *config.UserConfig
	app                       *tview.Application
	pages                     *tview.Pages
	options                   tui.ApplicationOptions
	entries                   []gamesmenu.Entry
	mainSelection, mainButton int
}

func newUI(cfg *config.UserConfig, manager *gamesmenu.Manager) *ui {
	return &ui{
		manager: manager, cfg: cfg,
		options: tui.ApplicationOptions{Theme: cfg.TUI.Theme, Mouse: cfg.TUI.Mouse, CRTMode: cfg.TUI.CRTMode},
	}
}

func (u *ui) Run() error {
	builder := func() (*tview.Application, error) {
		app, err := tui.NewApplication(u.options)
		if err != nil {
			return nil, fmt.Errorf("create TUI application: %w", err)
		}
		u.app, u.pages = app, tview.NewPages()
		if err := u.showMain(); err != nil {
			return nil, err
		}
		app.SetRoot(tui.WrapRoot(u.options, u.pages), true)
		u.showWelcome()
		return app, nil
	}
	if err := tui.BuildAndRetry(builder); err != nil {
		return fmt.Errorf("run GamesMenu TUI: %w", err)
	}
	return nil
}

func (u *ui) showWelcome() {
	if !u.manager.MenuExists() {
		tui.ShowInfoModal(u.pages, u.app, "Welcome", welcomeText, u.renderMain)
	}
}

func (u *ui) showMain() error {
	entries, err := u.manager.Discover()
	if err != nil {
		return fmt.Errorf("discover games folders: %w", err)
	}
	u.entries = entries
	u.renderMain()
	return nil
}

func (u *ui) mustShowMain() {
	if err := u.showMain(); err != nil {
		u.showError(err, u.app.Stop)
	}
}

func (u *ui) renderMain() {
	list := tui.NewMenuList()
	for _, entry := range u.entries {
		marker := "[ ] "
		if entry.Selected {
			marker = "[x] "
		}
		list.AddRow(marker+entry.Display, tui.MenuRowItem)
	}
	if len(u.entries) == 0 {
		list.AddHeader("No games folders found")
	} else {
		list.SetCurrentItem(min(u.mainSelection, len(u.entries)-1))
	}
	bar := tui.NewButtonBar(u.app)
	updateSelection := func(update func()) {
		if len(u.entries) == 0 {
			return
		}
		u.mainSelection, u.mainButton = list.GetCurrentItem(), bar.FocusedIndex()
		buttonFocus := bar.HasFocus()
		update()
		u.renderMain()
		if buttonFocus {
			_, primitive := u.pages.GetFrontPage()
			frame, ok := primitive.(*tui.PageFrame)
			if ok {
				frame.FocusButtonBar()
			}
		}
	}
	toggle := func() {
		updateSelection(func() { u.entries[u.mainSelection].Selected = !u.entries[u.mainSelection].Selected })
	}
	bulk := func() {
		updateSelection(func() {
			selectAll := false
			for _, entry := range u.entries {
				if !entry.Selected {
					selectAll = true
					break
				}
			}
			for i := range u.entries {
				u.entries[i].Selected = selectAll
			}
		})
	}
	remember := func(action func()) func() {
		return func() { u.mainSelection, u.mainButton = list.GetCurrentItem(), bar.FocusedIndex(); action() }
	}
	list.SetSelectedFunc(func(int, string, string, rune) { toggle() })
	bar.AddButtonWithHelp("Toggle", "Include or exclude this games folder", toggle).
		AddButtonWithHelp("All/None", "Select all folders, or clear selection when all are selected", bulk).
		AddButtonWithHelp("Generate", "Create shortcuts; confirm folder removal", remember(u.startGenerate)).
		AddButtonWithHelp("Clean Up", "Remove shortcuts whose game is gone", remember(u.startCleanUp)).
		AddButtonWithHelp("Settings", "Configure GamesMenu interface", remember(u.startSettings)).
		AddButtonWithHelp("Exit", "Exit without changing the Games menu", u.app.Stop).
		SetupNavigation(u.app.Stop)
	frame := tui.NewPageFrame(u.app).SetTitle("GamesMenu").SetContent(list).
		SetButtonBar(bar).SetOnEscape(u.app.Stop)
	help := func(index int) {
		if index >= 0 && index < len(u.entries) {
			frame.SetHelpText(strings.Join(u.entries[index].Paths, ", "))
		} else {
			frame.SetHelpText("Add games to a supported system folder, then restart GamesMenu.")
		}
	}
	list.SetRowChangedFunc(func(index int) { u.mainSelection = index; help(index) })
	bar.SetHelpCallback(func(text string) { frame.SetHelpText(text) })
	frame.SetupContentToButtonNavigation()
	bar.SetFocusedIndex(u.mainButton)
	help(list.GetCurrentItem())
	u.pages.AddAndSwitchToPage(pageMain, frame, true)
	u.app.SetFocus(list)
}

func (u *ui) showError(err error, onDismiss func()) {
	tui.ShowErrorModal(u.pages, u.app, err.Error(), onDismiss)
}

func (u *ui) startGenerate() {
	var selected []gamesmenu.Entry
	for _, entry := range u.entries {
		if entry.Selected {
			selected = append(selected, entry)
		}
	}
	if len(selected) == 0 {
		tui.ShowConfirmModal(u.pages, u.app, "Remove Games menu", "Remove the Games menu from your system?", func() {
			if err := u.manager.RemoveAll(); err != nil {
				u.showError(err, u.renderMain)
				return
			}
			tui.ShowInfoModal(u.pages, u.app, "GamesMenu", "Games menu removed.", u.mustShowMain)
		}, u.renderMain)
		return
	}
	removed, err := u.manager.DeselectedFolders(selected)
	if err != nil {
		u.showError(err, u.renderMain)
		return
	}
	generate := func() { u.generate(selected) }
	if len(removed) > 0 {
		message := fmt.Sprintf("Generate will remove %d folders and everything inside them:\n", len(removed)) +
			strings.Join(removed, "\n")
		tui.ShowConfirmModal(u.pages, u.app, "Remove deselected folders", message, generate, u.renderMain)
	} else {
		generate()
	}
}

func (u *ui) generate(selected []gamesmenu.Entry) {
	var result gamesmenu.GenerateResult
	options := tui.ProgressOptions{Title: "Creating MGL files..."}
	tui.ShowProgressModal(u.pages, u.app, options, func(update func(tui.ProgressUpdate)) error {
		var err error
		result, err = u.manager.Generate(selected, func(p gamesmenu.Progress) {
			text := "Finished scanning"
			if p.Entry != nil {
				text = fmt.Sprintf("Scanning %s (%s)", p.Entry.Display, p.Folder)
			}
			percent := 100
			if p.Total > 0 {
				percent = (p.Index*100 + p.Total - 1) / p.Total
			}
			update(tui.ProgressUpdate{
				Text: fmt.Sprintf("%s\n%d shortcuts created", text, p.Created), Current: percent, Total: 100,
			})
		})
		if err != nil {
			return fmt.Errorf("generate menu: %w", err)
		}
		return nil
	}, func(err error) {
		if err != nil {
			u.showError(err, u.mustShowMain)
			return
		}
		text := fmt.Sprintf("Created: %d\nSkipped: %d\nFailed: %d\nFolders removed: %d",
			result.Scan.Created, result.Scan.Skipped, result.Scan.Failed, len(result.Removed))
		if len(result.Removed) > 0 {
			text += "\n" + strings.Join(result.Removed, ", ")
		}
		text += errorSummary(result.Scan.Errors, result.Scan.ErrorsDropped)
		tui.ShowInfoModal(u.pages, u.app, "Finished scanning", text, u.mustShowMain)
	})
}

func errorSummary(errs []error, dropped int) string {
	var text strings.Builder
	for _, err := range errs {
		_, _ = text.WriteString("\n" + err.Error())
	}
	if dropped > 0 {
		_, _ = fmt.Fprintf(&text, "\n%d more errors", dropped)
	}
	return text.String()
}

func (u *ui) startCleanUp() {
	if !u.manager.MenuExists() {
		tui.ShowInfoModal(u.pages, u.app, "Clean Up", "No Games menu exists yet.", u.renderMain)
		return
	}
	message := "Remove shortcuts whose game or ZIP member is gone? Empty subfolders will also be removed."
	tui.ShowConfirmModal(u.pages, u.app, "Clean Up", message, u.cleanUp, u.renderMain)
}

func (u *ui) cleanUp() {
	var result gamesmenu.CleanupResult
	options := tui.ProgressOptions{Title: "Cleaning up..."}
	tui.ShowProgressModal(u.pages, u.app, options, func(update func(tui.ProgressUpdate)) error {
		var err error
		result, err = u.manager.CleanUp(func(p gamesmenu.CleanupProgress) {
			update(tui.ProgressUpdate{Text: fmt.Sprintf("Checked: %d\nRemoved: %d", p.Checked, p.Removed)})
		})
		if err != nil {
			return fmt.Errorf("clean up menu: %w", err)
		}
		return nil
	}, func(err error) {
		if err != nil {
			u.showError(err, u.mustShowMain)
			return
		}
		text := fmt.Sprintf("Checked: %d\nRemoved: %d\nEmpty folders pruned: %d\nUnreadable: %d",
			result.Checked, result.Removed, result.PrunedFolders, result.Unreadable)
		if result.Unavailable > 0 {
			text += fmt.Sprintf("\nSkipped (storage not attached): %d", result.Unavailable)
		}
		text += errorSummary(result.Errors, result.ErrorsDropped)
		tui.ShowInfoModal(u.pages, u.app, "Clean Up complete", text, u.mustShowMain)
	})
}
