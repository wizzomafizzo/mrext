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

// Portions adapted from Zaparoo Core, Copyright (c) 2026 The Zaparoo Project Contributors.

package tui

import (
	"sync"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

type Theme struct {
	Name                     string
	DisplayName              string
	BackgroundName           string
	AccentName               string
	TextName                 string
	HighlightBackgroundName  string
	HighlightForegroundName  string
	SecondaryName            string
	ErrorName                string
	WarningName              string
	SuccessName              string
	LabelName                string
	PrimitiveBackgroundColor tcell.Color
	ContrastBackgroundColor  tcell.Color
	BorderColor              tcell.Color
	PrimaryTextColor         tcell.Color
	SecondaryTextColor       tcell.Color
	InverseTextColor         tcell.Color
	FieldFocusedBackground   tcell.Color
	FieldUnfocusedBackground tcell.Color
	ProgressFillColor        tcell.Color
	ProgressEmptyColor       tcell.Color
	ErrorColor               tcell.Color
	WarningColor             tcell.Color
	SuccessColor             tcell.Color
	LabelColor               tcell.Color
}

var ThemeDefault = Theme{
	Name:                     "default",
	DisplayName:              "Default (Dark Blue)",
	BackgroundName:           "darkblue",
	AccentName:               "yellow",
	TextName:                 "white",
	HighlightBackgroundName:  "yellow",
	HighlightForegroundName:  "black",
	SecondaryName:            "gray",
	ErrorName:                "red",
	WarningName:              "yellow",
	SuccessName:              "green",
	LabelName:                "gray",
	PrimitiveBackgroundColor: tcell.ColorDarkBlue,
	ContrastBackgroundColor:  tcell.ColorBlue,
	BorderColor:              tcell.ColorLightYellow,
	PrimaryTextColor:         tcell.ColorWhite,
	SecondaryTextColor:       tcell.ColorGray,
	InverseTextColor:         tcell.ColorDarkBlue,
	FieldFocusedBackground:   tcell.ColorBlue,
	FieldUnfocusedBackground: tcell.ColorDarkBlue,
	ProgressFillColor:        tcell.ColorGreen,
	ProgressEmptyColor:       tcell.ColorGray,
	ErrorColor:               tcell.ColorRed,
	WarningColor:             tcell.ColorYellow,
	SuccessColor:             tcell.ColorGreen,
	LabelColor:               tcell.ColorGray,
}

var ThemeHighContrast = Theme{
	Name:                     "high_contrast",
	DisplayName:              "High Contrast",
	BackgroundName:           "#000000",
	AccentName:               "yellow",
	TextName:                 "white",
	HighlightBackgroundName:  "yellow",
	HighlightForegroundName:  "black",
	SecondaryName:            "white",
	ErrorName:                "red",
	WarningName:              "yellow",
	SuccessName:              "lime",
	LabelName:                "white",
	PrimitiveBackgroundColor: tcell.NewHexColor(0x000000),
	ContrastBackgroundColor:  tcell.NewHexColor(0x000000),
	BorderColor:              tcell.ColorYellow,
	PrimaryTextColor:         tcell.ColorWhite,
	SecondaryTextColor:       tcell.ColorWhite,
	InverseTextColor:         tcell.NewHexColor(0x000000),
	FieldFocusedBackground:   tcell.ColorYellow,
	FieldUnfocusedBackground: tcell.NewHexColor(0x000000),
	ProgressFillColor:        tcell.ColorYellow,
	ProgressEmptyColor:       tcell.ColorWhite,
	ErrorColor:               tcell.ColorRed,
	WarningColor:             tcell.ColorYellow,
	SuccessColor:             tcell.ColorLime,
	LabelColor:               tcell.ColorWhite,
}

var ThemeDracula = Theme{
	Name:                     "dracula",
	DisplayName:              "Dracula",
	BackgroundName:           "#282a36",
	AccentName:               "#bd93f9",
	TextName:                 "#f8f8f2",
	HighlightBackgroundName:  "#bd93f9",
	HighlightForegroundName:  "black",
	SecondaryName:            "#6272a4",
	ErrorName:                "#ff5555",
	WarningName:              "#f1fa8c",
	SuccessName:              "#50fa7b",
	LabelName:                "#6272a4",
	PrimitiveBackgroundColor: tcell.NewHexColor(0x282A36),
	ContrastBackgroundColor:  tcell.NewHexColor(0x44475A),
	BorderColor:              tcell.NewHexColor(0xBD93F9),
	PrimaryTextColor:         tcell.NewHexColor(0xF8F8F2),
	SecondaryTextColor:       tcell.NewHexColor(0x6272A4),
	InverseTextColor:         tcell.NewHexColor(0x282A36),
	FieldFocusedBackground:   tcell.NewHexColor(0x44475A),
	FieldUnfocusedBackground: tcell.NewHexColor(0x282A36),
	ProgressFillColor:        tcell.NewHexColor(0x50FA7B),
	ProgressEmptyColor:       tcell.NewHexColor(0x44475A),
	ErrorColor:               tcell.NewHexColor(0xFF5555),
	WarningColor:             tcell.NewHexColor(0xF1FA8C),
	SuccessColor:             tcell.NewHexColor(0x50FA7B),
	LabelColor:               tcell.NewHexColor(0x6272A4),
}

var ThemeNord = Theme{
	Name:                     "nord",
	DisplayName:              "Nord",
	BackgroundName:           "#2e3440",
	AccentName:               "#88c0d0",
	TextName:                 "#eceff4",
	HighlightBackgroundName:  "#88c0d0",
	HighlightForegroundName:  "black",
	SecondaryName:            "#d8dee9",
	ErrorName:                "#bf616a",
	WarningName:              "#ebcb8b",
	SuccessName:              "#a3be8c",
	LabelName:                "#4c566a",
	PrimitiveBackgroundColor: tcell.NewHexColor(0x2E3440),
	ContrastBackgroundColor:  tcell.NewHexColor(0x3B4252),
	BorderColor:              tcell.NewHexColor(0x88C0D0),
	PrimaryTextColor:         tcell.NewHexColor(0xECEFF4),
	SecondaryTextColor:       tcell.NewHexColor(0xD8DEE9),
	InverseTextColor:         tcell.NewHexColor(0x2E3440),
	FieldFocusedBackground:   tcell.NewHexColor(0x3B4252),
	FieldUnfocusedBackground: tcell.NewHexColor(0x2E3440),
	ProgressFillColor:        tcell.NewHexColor(0xA3BE8C),
	ProgressEmptyColor:       tcell.NewHexColor(0x4C566A),
	ErrorColor:               tcell.NewHexColor(0xBF616A),
	WarningColor:             tcell.NewHexColor(0xEBCB8B),
	SuccessColor:             tcell.NewHexColor(0xA3BE8C),
	LabelColor:               tcell.NewHexColor(0x4C566A),
}

var ThemeGruvbox = Theme{
	Name:                     "gruvbox",
	DisplayName:              "Gruvbox",
	BackgroundName:           "#282828",
	AccentName:               "#fabd2f",
	TextName:                 "#ebdbb2",
	HighlightBackgroundName:  "#fabd2f",
	HighlightForegroundName:  "black",
	SecondaryName:            "#a89984",
	ErrorName:                "#fb4934",
	WarningName:              "#fabd2f",
	SuccessName:              "#b8bb26",
	LabelName:                "#a89984",
	PrimitiveBackgroundColor: tcell.NewHexColor(0x282828),
	ContrastBackgroundColor:  tcell.NewHexColor(0x3C3836),
	BorderColor:              tcell.NewHexColor(0xFABD2F),
	PrimaryTextColor:         tcell.NewHexColor(0xEBDBB2),
	SecondaryTextColor:       tcell.NewHexColor(0xA89984),
	InverseTextColor:         tcell.NewHexColor(0x282828),
	FieldFocusedBackground:   tcell.NewHexColor(0x504945),
	FieldUnfocusedBackground: tcell.NewHexColor(0x282828),
	ProgressFillColor:        tcell.NewHexColor(0xB8BB26),
	ProgressEmptyColor:       tcell.NewHexColor(0x504945),
	ErrorColor:               tcell.NewHexColor(0xFB4934),
	WarningColor:             tcell.NewHexColor(0xFABD2F),
	SuccessColor:             tcell.NewHexColor(0xB8BB26),
	LabelColor:               tcell.NewHexColor(0xA89984),
}

var ThemeMonoGreen = Theme{
	Name:                     "monogreen",
	DisplayName:              "Mono Green (Retro)",
	BackgroundName:           "black",
	AccentName:               "green",
	TextName:                 "green",
	HighlightBackgroundName:  "green",
	HighlightForegroundName:  "black",
	SecondaryName:            "darkgreen",
	ErrorName:                "red",
	WarningName:              "yellow",
	SuccessName:              "lime",
	LabelName:                "darkgreen",
	PrimitiveBackgroundColor: tcell.ColorBlack,
	ContrastBackgroundColor:  tcell.NewHexColor(0x0A1A0A),
	BorderColor:              tcell.ColorGreen,
	PrimaryTextColor:         tcell.ColorGreen,
	SecondaryTextColor:       tcell.ColorDarkGreen,
	InverseTextColor:         tcell.ColorBlack,
	FieldFocusedBackground:   tcell.ColorDarkGreen,
	FieldUnfocusedBackground: tcell.ColorBlack,
	ProgressFillColor:        tcell.ColorLime,
	ProgressEmptyColor:       tcell.ColorDarkGreen,
	ErrorColor:               tcell.ColorRed,
	WarningColor:             tcell.ColorYellow,
	SuccessColor:             tcell.ColorLime,
	LabelColor:               tcell.ColorDarkGreen,
}

var AvailableThemes = map[string]*Theme{
	"default":       &ThemeDefault,
	"high_contrast": &ThemeHighContrast,
	"dracula":       &ThemeDracula,
	"nord":          &ThemeNord,
	"gruvbox":       &ThemeGruvbox,
	"monogreen":     &ThemeMonoGreen,
}

var ThemeNames = []string{"default", "high_contrast", "dracula", "nord", "gruvbox", "monogreen"}

var (
	currentTheme = &ThemeDefault
	themeMu      sync.RWMutex
)

func CurrentTheme() *Theme {
	themeMu.RLock()
	defer themeMu.RUnlock()
	return currentTheme
}

func SetCurrentTheme(name string) bool {
	theme, ok := AvailableThemes[name]
	if !ok {
		return false
	}
	themeMu.Lock()
	currentTheme = theme
	themeMu.Unlock()
	ApplyTheme(theme)
	return true
}

func ApplyTheme(theme *Theme) {
	tview.Styles.PrimitiveBackgroundColor = theme.PrimitiveBackgroundColor
	tview.Styles.ContrastBackgroundColor = theme.ContrastBackgroundColor
	tview.Styles.BorderColor = theme.BorderColor
	tview.Styles.PrimaryTextColor = theme.PrimaryTextColor
	tview.Styles.SecondaryTextColor = theme.SecondaryTextColor
	tview.Styles.InverseTextColor = theme.InverseTextColor
}

// ThemeLabel returns a theme's display name, falling back to its key. Every
// settings page needs this; each app used to carry its own copy.
func ThemeLabel(name string) string {
	if theme := AvailableThemes[name]; theme != nil {
		return theme.DisplayName
	}
	return name
}
