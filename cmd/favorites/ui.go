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
	"os"
	"path/filepath"
	"strings"

	"github.com/rivo/tview"
	"github.com/wizzomafizzo/mrext/pkg/config"
	"github.com/wizzomafizzo/mrext/pkg/favorites"
	"github.com/wizzomafizzo/mrext/pkg/tui"
)

const (
	pageMain = "favorites_main"
	// appTitle is the first segment of every page title.
	appTitle    = "Favorites Manager"
	pageBrowser = "favorites_browser"
	pageFolder  = "favorites_folder"
	pageModify  = "favorites_modify"
)

//nolint:govet // Field order groups runtime dependencies and navigation state.
type ui struct {
	manager          *favorites.Manager
	cfg              *config.UserConfig
	app              *tview.Application
	pages            *tview.Pages
	options          tui.ApplicationOptions
	mainSelection    int
	browserSelection map[string]int
}

type browserChoice struct {
	entry    *favorites.BrowseEntry
	path     string
	external bool
	parent   bool
}

func newUI(cfg *config.UserConfig, manager *favorites.Manager) *ui {
	return &ui{
		manager: manager,
		cfg:     cfg,
		options: tui.ApplicationOptions{
			Theme:            cfg.TUI.Theme,
			Mouse:            cfg.TUI.Mouse,
			CRTMode:          cfg.TUI.CRTMode,
			OnScreenKeyboard: cfg.TUI.OnScreenKeyboard,
		},
		browserSelection: make(map[string]int),
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
		return app, nil
	}
	if err := tui.BuildAndRetry(builder); err != nil {
		return fmt.Errorf("run Favorites TUI: %w", err)
	}
	return nil
}

func (u *ui) showMain() error {
	items, err := u.manager.EditableItems()
	if err != nil {
		return fmt.Errorf("list editable Favorites: %w", err)
	}
	list := tui.NewMenuList().
		AddRow("Add favorite", tui.MenuRowAdd).
		AddHeader("Existing favorites")
	for _, item := range items {
		label := strings.TrimSuffix(u.manager.RelativePath(item.Path), filepath.Ext(item.Path))
		if item.Kind == favorites.FavoriteFolder {
			label += "/"
		}
		kind := tui.MenuRowItem
		if item.Kind == favorites.FavoriteFolder {
			kind = tui.MenuRowFolder
		}
		list.AddRow(truncateLeading(label, 65), kind)
	}
	if list.GetItemCount() > 0 {
		list.SetCurrentItem(min(u.mainSelection, list.GetItemCount()-1))
	}
	activate := func() {
		selection := list.GetCurrentItem()
		u.mainSelection = selection
		if selection == 0 {
			u.showBrowser(u.managerRoot())
			return
		}
		if selection >= 2 {
			u.showModify(&items[selection-2])
		}
	}
	list.SetSelectedFunc(func(int, string, string, rune) { activate() })

	bar := tui.NewButtonBar(u.app).
		AddButtonWithHelp("Select", "Open selected favorite or add a new one", activate).
		AddButtonWithHelp("Create Folder", "Create a nested Favorites folder", u.startCreateFolder).
		AddButtonWithHelp("Settings", "Configure Favorites Manager", u.startSettings).
		AddButtonWithHelp("Exit", "Exit Favorites Manager", u.app.Stop).
		SetupNavigation(u.app.Stop)
	frame := tui.NewPageFrame(u.app).
		SetTitle(appTitle).
		SetContent(list).
		SetHelpText("Add a new favorite or select an existing one to modify.").
		SetButtonBar(bar).
		SetOnEscape(u.app.Stop)
	bar.SetHelpCallback(func(text string) { frame.SetHelpText(text) })
	frame.SetupContentToButtonNavigation()
	u.pages.AddAndSwitchToPage(pageMain, frame, true)
	u.app.SetFocus(list)
	return nil
}

func (u *ui) managerRoot() string {
	return u.manager.Root()
}

func (u *ui) startCreateFolder() {
	u.createFolder(u.mustShowMain)
}

func (u *ui) createFolder(onReturn func()) {
	u.showFolderPicker(false, "", "Select parent folder", func(parent string) {
		u.showNameInput(tui.InputOptions{
			Title:        "Create Folder",
			Prompt:       "Enter a folder name. It must start with an underscore (_).",
			InitialValue: "_",
			Validate:     favorites.ValidateFolderName,
		}, func(name string) {
			if _, err := u.manager.CreateFolder(parent, name); err != nil {
				u.showError(err, func() { u.createFolder(onReturn) })
				return
			}
			onReturn()
		}, onReturn)
	}, onReturn)
}

func (u *ui) showBrowser(folder string) {
	entries, err := u.manager.ListDirectory(folder)
	if err != nil {
		u.showError(err, u.mustShowMain)
		return
	}
	choices := make([]browserChoice, 0, len(entries)+2)
	list := tui.NewMenuList()

	if samePath(folder, u.managerRoot()) {
		external := u.manager.ExternalFolder()
		if files, readErr := os.ReadDir(external); readErr == nil && len(files) > 0 {
			choices = append(choices, browserChoice{path: external, external: true})
			list.AddRow("Go to USB drive", tui.MenuRowAction)
		}
	}
	if parent, parentErr := u.manager.ParentFolder(folder); parentErr == nil && !samePath(parent, folder) {
		choices = append(choices, browserChoice{path: parent, parent: true})
		list.AddRow("..  Parent folder", tui.MenuRowAction)
	}
	for index := range entries {
		entry := entries[index]
		choices = append(choices, browserChoice{entry: &entry, path: entry.Path})
		kind := tui.MenuRowItem
		if entry.IsDirectory {
			kind = tui.MenuRowFolder
		}
		list.AddRow(entry.Label, kind)
	}
	if len(choices) == 0 {
		list.AddHeader("No selectable files found")
	}
	if selection := u.browserSelection[folder]; selection < list.GetItemCount() {
		list.SetCurrentItem(selection)
	}
	activate := func() {
		if len(choices) == 0 {
			return
		}
		selection := list.GetCurrentItem()
		u.browserSelection[folder] = selection
		choice := choices[selection]
		if choice.external || choice.parent || choice.entry.IsDirectory {
			next := choice.path
			if choice.entry != nil && choice.entry.IsArchive && !strings.Contains(
				strings.ToLower(filepath.ToSlash(next)), ".zip/",
			) {
				next += string(filepath.Separator)
			}
			u.showBrowser(next)
			return
		}
		u.startAddFavorite(*choice.entry)
	}
	list.SetSelectedFunc(func(int, string, string, rune) { activate() })
	back := func() {
		parent, parentErr := u.manager.ParentFolder(folder)
		if parentErr != nil || samePath(parent, folder) {
			u.mustShowMain()
			return
		}
		u.showBrowser(parent)
	}
	browseArchive := func() {
		if len(choices) == 0 {
			return
		}
		choice := choices[list.GetCurrentItem()]
		if choice.entry != nil && choice.entry.IsArchive {
			u.showBrowser(choice.path + string(filepath.Separator))
			return
		}
		activate()
	}
	hasSelectableArchive := false
	for _, choice := range choices {
		if choice.entry != nil && choice.entry.IsArchive && !choice.entry.IsDirectory {
			hasSelectableArchive = true
			break
		}
	}
	bar := tui.NewButtonBar(u.app).
		AddButtonWithHelp("Select", "Open folder or select favorite", activate)
	if hasSelectableArchive {
		bar.AddButtonWithHelp("Browse", "Browse inside selected ZIP", browseArchive)
	}
	bar.AddButtonWithHelp("Back", "Return to previous folder", back).
		AddButtonWithHelp("Cancel", "Return to Favorites Manager", u.mustShowMain).
		SetupNavigation(back)
	frame := tui.NewPageFrame(u.app).
		SetTitle(appTitle, "Select Favorite").
		SetContent(list).
		SetHelpText(folder).
		SetButtonBar(bar).
		SetOnEscape(back)
	bar.SetHelpCallback(func(text string) {
		if text == "" {
			frame.SetHelpText(folder)
		} else {
			frame.SetHelpText(text)
		}
	})
	frame.SetupContentToButtonNavigation()
	u.pages.AddAndSwitchToPage(pageBrowser, frame, true)
	u.app.SetFocus(list)
}

func (u *ui) startAddFavorite(entry favorites.BrowseEntry) {
	back := func() {
		parent, err := u.manager.ParentFolder(entry.Path)
		if err != nil {
			u.showError(err, u.mustShowMain)
			return
		}
		u.showBrowser(parent)
	}
	u.showFolderPicker(true, "", "Select destination folder", func(destination string) {
		defaultName := u.manager.DefaultName(entry.System, entry.Path)
		u.showNameInput(tui.InputOptions{
			Title:        "Favorite Name",
			Prompt:       "Enter a display name for the favorite.",
			InitialValue: defaultName,
			Validate:     favorites.ValidateDisplayName,
		}, func(name string) {
			var err error
			if entry.System == nil {
				_, err = u.manager.CreateCoreFavorite(entry.Path, destination, name)
			} else {
				_, err = u.manager.CreateGameFavorite(entry.System, entry.Path, destination, name)
			}
			if err != nil {
				u.showError(err, func() { u.startAddFavorite(entry) })
				return
			}
			u.mustShowMain()
		}, back)
	}, back)
}

func (u *ui) showFolderPicker(
	includeRoot bool,
	ignorePath, title string,
	onSelect func(string),
	onCancel func(),
) {
	folders, err := u.manager.DestinationFolders(includeRoot, ignorePath)
	if err != nil {
		u.showError(err, onCancel)
		return
	}
	list := tui.NewMenuList()
	for _, folder := range folders {
		label := u.manager.RelativePath(folder)
		kind := tui.MenuRowFolder
		if label == "." {
			label = "Top level"
			kind = tui.MenuRowAction
		} else {
			label += "/"
		}
		list.AddRow(label, kind)
	}
	if len(folders) == 0 {
		list.AddHeader("No Favorites folders found")
	}
	activate := func() {
		if len(folders) > 0 {
			onSelect(folders[list.GetCurrentItem()])
		}
	}
	list.SetSelectedFunc(func(int, string, string, rune) { activate() })
	bar := tui.NewButtonBar(u.app).
		AddButtonWithHelp("Select", "Use selected destination", activate)
	if includeRoot {
		bar.AddButtonWithHelp("Create Folder", "Create a destination folder", func() {
			u.createFolder(func() {
				u.showFolderPicker(includeRoot, ignorePath, title, onSelect, onCancel)
			})
		})
	}
	bar.AddButtonWithHelp("Cancel", "Cancel folder selection", onCancel).
		SetupNavigation(onCancel)
	frame := tui.NewPageFrame(u.app).
		SetTitle(appTitle, title).
		SetContent(list).
		SetHelpText(title + ".").
		SetButtonBar(bar).
		SetOnEscape(onCancel)
	bar.SetHelpCallback(func(text string) { frame.SetHelpText(text) })
	frame.SetupContentToButtonNavigation()
	u.pages.AddAndSwitchToPage(pageFolder, frame, true)
	u.app.SetFocus(list)
}

func (u *ui) showModify(item *favorites.Favorite) {
	name := strings.TrimSuffix(filepath.Base(item.Path), filepath.Ext(item.Path))
	info := []string{
		formatDetail("Name", name),
		formatDetail("Folder", u.manager.RelativePath(filepath.Dir(item.Path))),
		formatDetail("Type", string(item.Kind)),
	}
	if item.System != "" {
		info = append(info, formatDetail("System", item.System))
	}
	if item.SetName != "" {
		info = append(info, formatDetail("Set name", item.SetName))
	}
	if item.Target != "" {
		info = append(info, formatDetail("File", item.Target))
	}
	details := tview.NewTextView().
		SetDynamicColors(true).
		SetText(strings.Join(info, "\n")).
		SetWordWrap(true)
	list := tui.NewMenuList().
		AddRow("Rename", tui.MenuRowAction).
		AddRow("Move", tui.MenuRowAction).
		AddRow("Delete", tui.MenuRowAction)
	activate := func() {
		switch list.GetCurrentItem() {
		case 0:
			u.renameItem(item, name)
		case 1:
			u.moveItem(item)
		case 2:
			u.deleteItem(item)
		}
	}
	list.SetSelectedFunc(func(int, string, string, rune) { activate() })
	bar := tui.NewButtonBar(u.app).
		AddButtonWithHelp("Select", "Apply selected action", activate).
		AddButtonWithHelp("Back", "Return to Favorites Manager", u.mustShowMain).
		SetupNavigation(u.mustShowMain)
	content := tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(details, 7, 0, false).
		AddItem(list, 0, 1, true)
	frame := tui.NewPageFrame(u.app).
		SetTitle(appTitle, "Modify").
		SetContent(content).
		SetFocusTarget(list).
		SetHelpText("Select an action.").
		SetButtonBar(bar).
		SetOnEscape(u.mustShowMain)
	bar.SetHelpCallback(func(text string) { frame.SetHelpText(text) })
	frame.SetupContentToButtonNavigation()
	u.pages.AddAndSwitchToPage(pageModify, frame, true)
	u.app.SetFocus(list)
}

func (u *ui) renameItem(item *favorites.Favorite, initial string) {
	validator := favorites.ValidateDisplayName
	if item.Kind == favorites.FavoriteFolder {
		validator = favorites.ValidateFolderName
		initial = filepath.Base(item.Path)
	}
	u.showNameInput(tui.InputOptions{
		Title:        "Rename Favorite",
		Prompt:       "Enter a new display name.",
		InitialValue: initial,
		Validate:     validator,
	}, func(name string) {
		if _, err := u.manager.Rename(item.Path, name); err != nil {
			u.showError(err, func() { u.showModify(item) })
			return
		}
		u.mustShowMain()
	}, func() { u.showModify(item) })
}

func (u *ui) moveItem(item *favorites.Favorite) {
	includeRoot := item.Kind != favorites.FavoriteFolder
	u.showFolderPicker(includeRoot, item.Path, "Move Favorite", func(destination string) {
		if _, err := u.manager.Move(item.Path, destination); err != nil {
			u.showError(err, func() { u.showModify(item) })
			return
		}
		u.mustShowMain()
	}, func() { u.showModify(item) })
}

func (u *ui) deleteItem(item *favorites.Favorite) {
	message := "Delete favorite " + u.manager.RelativePath(item.Path) + "?"
	if item.Kind == favorites.FavoriteFolder {
		message = "Delete folder " + u.manager.RelativePath(item.Path) + "?"
	}
	tui.ShowConfirmModal(u.pages, u.app, "Favorites Manager", message, func() {
		if err := u.manager.Delete(item.Path); err != nil {
			u.showError(err, func() { u.showModify(item) })
			return
		}
		u.mustShowMain()
	}, func() { u.showModify(item) })
}

func (u *ui) showNameInput(options tui.InputOptions, onSubmit func(string), onCancel func()) {
	options.OnScreenKeyboard = u.options.OnScreenKeyboard
	tui.ShowInputModal(u.pages, u.app, options, onSubmit, onCancel)
}

func (u *ui) showError(err error, restore func()) {
	tui.ShowErrorModal(u.pages, u.app, err.Error(), restore)
}

func (u *ui) mustShowMain() {
	if err := u.showMain(); err != nil {
		u.showError(err, u.app.Stop)
	}
}

func formatDetail(label, value string) string {
	theme := tui.CurrentTheme()
	return fmt.Sprintf("[%s::b]%s:[-::-] %s", theme.LabelName, label, tview.Escape(value))
}

func truncateLeading(value string, maximum int) string {
	characters := []rune(value)
	if len(characters) < maximum {
		return value
	}
	return "..." + string(characters[len(characters)-(maximum-3):])
}

func samePath(first, second string) bool {
	return filepath.Clean(first) == filepath.Clean(second)
}
