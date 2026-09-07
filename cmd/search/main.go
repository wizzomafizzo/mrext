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
	"flag"
	"fmt"
	"os"

	"github.com/wizzomafizzo/mrext/pkg/config"
)

const appName = "search"

func main() {
	printPath := flag.Bool("print", false, "Print game path to stderr instead of launching the game")
	flag.Parse()

	cfg, err := config.LoadUserConfig(appName, &config.UserConfig{
		TUI: config.TUIConfig{Theme: "default", Mouse: true, CRTMode: true, OnScreenKeyboard: true},
	})
	if err != nil {
		_, _ = fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	view := newUI(cfg, !*printPath)
	if runErr := view.Run(); runErr != nil {
		_, _ = fmt.Fprintln(os.Stderr, runErr)
		os.Exit(1)
	}

	// -print keeps writing the chosen path to stderr, as it always has.
	if *printPath && view.selected != nil {
		_, _ = fmt.Fprintln(os.Stderr, view.selected.Path)
	}
}
