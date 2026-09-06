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

// Package gamesmenu mirrors MiSTer game libraries into stock-menu MGL shortcut
// folders. It is the Go port of scripts/gamesmenu.sh and keeps that script's
// on-disk contract: a _Games folder at the SD root, one _<name> folder per
// games folder, names.txt renaming of every folder component, and shortcuts
// that are created once and never rewritten.
package gamesmenu

import (
	"path/filepath"

	"github.com/wizzomafizzo/mrext/pkg/config"
)

const (
	// AppFilename is the installed binary name; the .sh suffix is load-bearing
	// for MiSTer's Scripts menu and Downloader entries.
	AppFilename = "gamesmenu.sh"
	// IniFilename sits beside the binary in the Scripts folder.
	IniFilename = "gamesmenu.ini"
	// MenuFolderName is the stock-menu folder that receives the shortcuts.
	MenuFolderName = "_Games"
	namesFilename  = "names.txt"
)

// Paths holds every filesystem location GamesMenu reads or writes outside the
// games folders, so tests and the --root development mode can relocate it.
type Paths struct {
	SDRoot     string
	MenuFolder string
	NamesFile  string
}

// DefaultPaths returns the locations used on a real MiSTer.
func DefaultPaths() Paths {
	return Paths{
		SDRoot:     config.SdFolder,
		MenuFolder: config.GamesMenuFolder,
		NamesFile:  config.NamesFile,
	}
}

// RootedPaths relocates every location under root, mirroring the SD layout.
func RootedPaths(root string) Paths {
	return Paths{
		SDRoot:     root,
		MenuFolder: filepath.Join(root, MenuFolderName),
		NamesFile:  filepath.Join(root, namesFilename),
	}
}
