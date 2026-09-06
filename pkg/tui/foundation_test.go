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
	"strconv"
	"strings"
	"testing"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

func TestKeyboardModalMatchesCompactZaparooLayout(t *testing.T) {
	app := tview.NewApplication()
	pages := tview.NewPages()
	ShowInputModal(pages, app, InputOptions{
		Title: "Name", Prompt: "Floating prompt must not appear", OnScreenKeyboard: true,
	}, func(string) {}, func() {})
	screen := tcell.NewSimulationScreen("UTF-8")
	if err := screen.Init(); err != nil {
		t.Fatal(err)
	}
	defer screen.Fini()
	screen.SetSize(75, 15)
	pages.SetRect(0, 0, 75, 15)
	pages.Draw(screen)
	keyboard, ok := app.GetFocus().(*VirtualKeyboard)
	if !ok {
		t.Fatal("keyboard not focused")
	}
	_, _, width, height := keyboard.GetRect()
	if width != 41 || height != 8 {
		t.Fatalf("keyboard size = %dx%d", width, height)
	}
	var text strings.Builder
	for y := range 15 {
		for x := range 75 {
			value, _, _ := screen.Get(x, y)
			_, _ = text.WriteString(value)
		}
	}
	if strings.Contains(text.String(), "Floating prompt") || !strings.Contains(text.String(), "SPC") {
		t.Fatalf("unexpected keyboard rendering: %s", text.String())
	}
}

func TestChoiceModalPreservesSelectionOnCancel(t *testing.T) {
	app := tview.NewApplication()
	pages := tview.NewPages()
	selected, cancelled := -1, false
	ShowChoiceModal(pages, app, "Core", []Choice{{Label: "Standard"}, {Label: "LLAPI"}, {Label: "YC"}},
		1, func(index int) { selected = index }, func() { cancelled = true })
	list, ok := app.GetFocus().(*MenuList)
	if !ok || list.GetCurrentItem() != 1 {
		t.Fatal("picker did not preselect current value")
	}
	list.InputHandler()(tcell.NewEventKey(tcell.KeyDown, 0, tcell.ModNone), nil)
	list.InputHandler()(tcell.NewEventKey(tcell.KeyEscape, 0, tcell.ModNone), nil)
	if selected != -1 || !cancelled || pages.HasPage("tui_choice_modal") {
		t.Fatal("cancel committed choice or left modal open")
	}
}

func TestSelectedRowsRenderPlainBlackAcrossThemes(t *testing.T) {
	t.Cleanup(func() { SetCurrentTheme("default") })
	screen := tcell.NewSimulationScreen("UTF-8")
	if err := screen.Init(); err != nil {
		t.Fatal(err)
	}
	defer screen.Fini()
	screen.SetSize(75, 15)
	for _, name := range ThemeNames {
		SetCurrentTheme(name)
		menu := NewMenuList().AddRow("Selected", MenuRowItem)
		menu.SetRect(0, 0, 75, 1)
		menu.Draw(screen)
		_, style, _ := screen.Get(2, 0)
		foreground, _, attributes := style.Decompose()
		if foreground != tcell.ColorBlack || attributes&tcell.AttrBold != 0 {
			t.Fatalf("theme %s selected style = %v", name, style)
		}
	}
}

func TestMenuListPagingSkipsHeaders(t *testing.T) {
	for _, height := range []int{2, 11, 30} {
		menu := NewMenuList().AddHeader("Start")
		for index := range 24 {
			if index%4 == 0 {
				menu.AddHeader("Section")
			} else {
				menu.AddRow(strconv.Itoa(index), MenuRowItem)
			}
		}
		menu.AddHeader("End")
		menu.SetRect(0, 0, 75, height)
		handle := menu.InputHandler()
		for _, key := range []tcell.Key{tcell.KeyPgDn, tcell.KeyPgUp} {
			for range 30 {
				handle(tcell.NewEventKey(key, 0, tcell.ModNone), nil)
				index := menu.GetCurrentItem()
				if !menu.selectable(index) || menu.selected != index {
					t.Fatalf("height=%d key=%v row=%d highlight=%d", height, key, index, menu.selected)
				}
			}
		}
	}
}

func TestMenuListBuildAllocationGrowthIsLinear(t *testing.T) {
	allocations := func(count int) float64 {
		return testing.AllocsPerRun(2, func() {
			menu := NewMenuList()
			for range count {
				menu.AddRow("Game (USA)", MenuRowItem)
			}
		})
	}
	small, large := allocations(200), allocations(800)
	if large > small*5 {
		t.Fatalf("superlinear menu allocation growth: 200=%f 800=%f", small, large)
	}
}

func BenchmarkMenuListBuild(b *testing.B) {
	for _, count := range []int{1000, 3000, 6000} {
		b.Run(strconv.Itoa(count), func(b *testing.B) {
			b.ReportAllocs()
			for b.Loop() {
				menu := NewMenuList()
				for range count {
					menu.AddRow("Super Mario World (USA) [Rev 1].sfc", MenuRowItem)
				}
			}
		})
	}
}

func TestNewApplicationAppliesTheme(t *testing.T) {
	app, err := NewApplication(ApplicationOptions{Theme: "nord", Mouse: true})
	if err != nil {
		t.Fatal(err)
	}
	if app == nil || CurrentTheme().Name != "nord" {
		t.Fatalf("application=%v theme=%q", app, CurrentTheme().Name)
	}
	if _, err := NewApplication(ApplicationOptions{Theme: "missing"}); err == nil {
		t.Fatal("unknown theme accepted")
	}
	if !SetCurrentTheme("default") {
		t.Fatal("failed to restore default theme")
	}
}

func TestVirtualKeyboardSupportsModesAndInput(t *testing.T) {
	submitted := ""
	cancelled := false
	keyboard := NewVirtualKeyboard("fav", func(text string) {
		submitted = text
	}, func() {
		cancelled = true
	})
	handle := keyboard.InputHandler()
	handle(tcell.NewEventKey(tcell.KeyRune, 's', tcell.ModNone), nil)
	if keyboard.layout()[keyboard.cursorRow][keyboard.cursorCol] != keyboardSubmit {
		t.Fatal("keyboard must start on OK")
	}
	handle(tcell.NewEventKey(tcell.KeyEnter, 0, tcell.ModNone), nil)
	if submitted != "favs" {
		t.Fatalf("submitted text = %q", submitted)
	}
	handle(tcell.NewEventKey(tcell.KeyEscape, 0, tcell.ModNone), nil)
	if !cancelled {
		t.Fatal("escape did not cancel")
	}
}

func TestMenuListStylesAndSkipsHeaders(t *testing.T) {
	if !SetCurrentTheme("default") {
		t.Fatal("failed to select default theme")
	}
	menu := NewMenuList().
		AddRow("Add favorite", MenuRowAdd).
		AddHeader("Existing favorites").
		AddRow("_@Favorites/SNES/Zelda", MenuRowItem)
	if menu.GetCurrentItem() != 0 {
		t.Fatalf("initial row = %d", menu.GetCurrentItem())
	}
	handle := menu.InputHandler()
	handle(tcell.NewEventKey(tcell.KeyDown, 0, tcell.ModNone), func(tview.Primitive) {})
	if menu.GetCurrentItem() != 2 {
		t.Fatalf("down selected header: row=%d", menu.GetCurrentItem())
	}
	handle(tcell.NewEventKey(tcell.KeyUp, 0, tcell.ModNone), func(tview.Primitive) {})
	if menu.GetCurrentItem() != 0 {
		t.Fatalf("up selected header: row=%d", menu.GetCurrentItem())
	}
	selected, _ := menu.GetItemText(0)
	header, _ := menu.GetItemText(1)
	if !strings.Contains(selected, "[black:yellow:-]") || !strings.Contains(header, "[gray:darkblue:b]") {
		t.Fatalf("safe palette formatting missing: selected=%q header=%q", selected, header)
	}
}

func TestListAndButtonBarUseIndependentAxes(t *testing.T) {
	app, err := NewApplication(ApplicationOptions{Theme: "default"})
	if err != nil {
		t.Fatal(err)
	}
	selected := ""
	list := tview.NewList().ShowSecondaryText(false).
		AddItem("First row", "", 0, nil).
		AddItem("Second row", "", 0, nil)
	bar := NewButtonBar(app).
		AddButton("Select", func() { selected = "select" }).
		AddButton("Cancel", func() { selected = "cancel" })
	frame := NewPageFrame(app).
		SetContent(list).
		SetButtonBar(bar)
	frame.SetupContentToButtonNavigation()
	app.SetFocus(list)
	handle := frame.InputHandler()
	setFocus := func(primitive tview.Primitive) { app.SetFocus(primitive) }

	handle(tcell.NewEventKey(tcell.KeyDown, 0, tcell.ModNone), setFocus)
	if list.GetCurrentItem() != 1 || !list.HasFocus() {
		t.Fatalf("down changed focus or failed to select row: row=%d focus=%v", list.GetCurrentItem(), list.HasFocus())
	}
	handle(tcell.NewEventKey(tcell.KeyRight, 0, tcell.ModNone), setFocus)
	if list.GetCurrentItem() != 1 || !list.HasFocus() || bar.focusedIndex != 1 {
		t.Fatalf(
			"right changed row or focus: row=%d focus=%v button=%d",
			list.GetCurrentItem(),
			list.HasFocus(),
			bar.focusedIndex,
		)
	}
	handle(tcell.NewEventKey(tcell.KeyEnter, 0, tcell.ModNone), setFocus)
	if selected != "cancel" {
		t.Fatalf("selected button = %q", selected)
	}
	if !bar.persistentFocus {
		t.Fatal("button bar is not persistently highlighted")
	}
}

func TestButtonBarNavigationAndActivation(t *testing.T) {
	app, err := NewApplication(ApplicationOptions{Theme: "default"})
	if err != nil {
		t.Fatal(err)
	}
	selected := ""
	bar := NewButtonBar(app).
		AddButton("First", func() { selected = "first" }).
		AddButton("Second", func() { selected = "second" })
	bar.Focus(func(tview.Primitive) {})
	handle := bar.InputHandler()
	setFocus := func(primitive tview.Primitive) { app.SetFocus(primitive) }
	handle(tcell.NewEventKey(tcell.KeyRight, 0, tcell.ModNone), setFocus)
	handle(tcell.NewEventKey(tcell.KeyEnter, 0, tcell.ModNone), setFocus)
	if selected != "second" {
		t.Fatalf("selected = %q", selected)
	}
	if bar.FocusedIndex() != 1 {
		t.Fatalf("focused index = %d", bar.FocusedIndex())
	}

	// A rebuilt page restores the highlight; out-of-range values are ignored.
	help := ""
	rebuilt := NewButtonBar(app).
		AddButtonWithHelp("First", "first help", func() { selected = "first" }).
		AddButtonWithHelp("Second", "second help", func() { selected = "second" }).
		SetHelpCallback(func(text string) { help = text })
	rebuilt.SetFocusedIndex(bar.FocusedIndex()).SetFocusedIndex(5).SetFocusedIndex(-1)
	if rebuilt.FocusedIndex() != 1 || help != "second help" {
		t.Fatalf("focused index = %d help = %q", rebuilt.FocusedIndex(), help)
	}
	rebuilt.InputHandler()(tcell.NewEventKey(tcell.KeyEnter, 0, tcell.ModNone), setFocus)
	if selected != "second" {
		t.Fatalf("selected = %q", selected)
	}
}

func TestLongDialogScrollsWhileKeepingButtonsIndependent(t *testing.T) {
	for _, size := range [][2]int{{75, 15}, {100, 30}} {
		screen := tcell.NewSimulationScreen("UTF-8")
		if err := screen.Init(); err != nil {
			t.Fatal(err)
		}
		t.Cleanup(screen.Fini)
		screen.SetSize(size[0], size[1])
		app := tview.NewApplication()
		selected := -1
		dialog := NewDialog().SetText("First line\n" + strings.Repeat("Summary detail\n", 50) + "Last line").
			AddButtons([]string{"Yes", "No"}).SetDoneFunc(func(index int) { selected = index })
		app.SetRoot(dialog, true)
		dialog.SetRect(0, 0, size[0], size[1])
		draw := func() string {
			screen.Clear()
			dialog.Draw(screen)
			var text strings.Builder
			for y := range size[1] {
				for x := range size[0] {
					value, _, _ := screen.Get(x, y)
					_, _ = text.WriteString(value)
				}
			}
			return text.String()
		}
		press := func(key tcell.Key) {
			dialog.InputHandler()(tcell.NewEventKey(key, 0, tcell.ModNone), func(p tview.Primitive) { app.SetFocus(p) })
		}
		if text := draw(); !strings.Contains(text, "First line") || !strings.Contains(text, "scroll") {
			t.Fatalf("missing initial message/help at %v: %s", size, text)
		}
		press(tcell.KeyEnd)
		if text := draw(); !strings.Contains(text, "Last line") || selected != -1 {
			t.Fatalf("could not scroll summary safely at %v: %s", size, text)
		}
		press(tcell.KeyHome)
		if text := draw(); !strings.Contains(text, "First line") {
			t.Fatalf("could not scroll back at %v: %s", size, text)
		}
		press(tcell.KeyRight)
		press(tcell.KeyEnter)
		if selected != 1 {
			t.Fatalf("button selection changed while scrolling: %d", selected)
		}
		// Short dialogs retain the original Up/Down button navigation.
		dialog.SetText("Short message")
		if text := draw(); strings.Contains(text, "scroll") {
			t.Fatalf("short dialog still shows scrolling help: %s", text)
		}
		press(tcell.KeyUp)
		press(tcell.KeyEnter)
		if selected != 0 {
			t.Fatalf("short dialog navigation changed: %d", selected)
		}
	}
}

func TestModalDismissalRestoresPageFocus(t *testing.T) {
	type harness struct {
		app   *tview.Application
		pages *tview.Pages
		root  tview.Primitive
		list  *MenuList
	}
	newHarness := func(t *testing.T) harness {
		t.Helper()
		app := tview.NewApplication()
		pages := tview.NewPages()
		list := NewMenuList().AddRow("Item", MenuRowItem)
		frame := NewPageFrame(app).SetTitle("Main").SetContent(list).SetFocusTarget(list)
		pages.AddAndSwitchToPage("main", frame, true)
		root := WrapRoot(ApplicationOptions{}, pages)
		app.SetRoot(root, true)
		if app.GetFocus() != list {
			t.Fatalf("initial focus = %T", app.GetFocus())
		}
		return harness{app: app, pages: pages, root: root, list: list}
	}
	press := func(h harness, keys ...tcell.Key) {
		setFocus := func(p tview.Primitive) { h.app.SetFocus(p) }
		for _, key := range keys {
			h.root.InputHandler()(tcell.NewEventKey(key, 0, tcell.ModNone), setFocus)
		}
	}
	cases := []struct {
		show func(h harness, done *bool)
		name string
		page string
		keys []tcell.Key
	}{
		{func(h harness, done *bool) {
			ShowInfoModal(h.pages, h.app, "Title", "Message", func() { *done = true })
		}, "info", infoModalPage, []tcell.Key{tcell.KeyEnter}},
		{func(h harness, done *bool) {
			ShowErrorModal(h.pages, h.app, "Message", func() { *done = true })
		}, "error", errorModalPage, []tcell.Key{tcell.KeyEnter}},
		{func(h harness, done *bool) {
			ShowConfirmModal(h.pages, h.app, "Title", "Message", func() { *done = true }, nil)
		}, "confirm yes", confirmModalPage, []tcell.Key{tcell.KeyEnter}},
		{func(h harness, done *bool) {
			ShowConfirmModal(h.pages, h.app, "Title", "Message", nil, func() { *done = true })
		}, "confirm escape", confirmModalPage, []tcell.Key{tcell.KeyEscape}},
		{func(h harness, done *bool) {
			ShowInputModal(h.pages, h.app, InputOptions{Title: "Name"}, func(string) { *done = true }, nil)
		}, "input submit", inputModalPage, []tcell.Key{tcell.KeyEnter, tcell.KeyEnter}},
		{func(h harness, done *bool) {
			ShowInputModal(h.pages, h.app, InputOptions{Title: "Name"}, nil, func() { *done = true })
		}, "input escape", inputModalPage, []tcell.Key{tcell.KeyEscape}},
		{func(h harness, done *bool) {
			options := InputOptions{Title: "Name", OnScreenKeyboard: true}
			ShowInputModal(h.pages, h.app, options, nil, func() { *done = true })
		}, "keyboard escape", inputModalPage, []tcell.Key{tcell.KeyEscape}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			h := newHarness(t)
			done := false
			tc.show(h, &done)
			if !h.pages.HasPage(tc.page) || h.app.GetFocus() == h.list {
				t.Fatal("modal did not open or take focus")
			}
			press(h, tc.keys...)
			if !done || h.pages.HasPage(tc.page) {
				t.Fatal("modal did not dismiss")
			}
			if !h.root.HasFocus() || h.app.GetFocus() != h.list {
				t.Fatalf("focus after dismissal = %T", h.app.GetFocus())
			}
		})
	}
}
