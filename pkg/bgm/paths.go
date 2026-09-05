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

package bgm

import (
	"os"
	"path/filepath"

	"github.com/wizzomafizzo/mrext/pkg/config"
)

const (
	// AppFilename is the installed binary name; the .sh suffix is kept for
	// MiSTer's Scripts menu and existing startup hooks.
	AppFilename    = "bgm.sh"
	MusicFolder    = "music"
	BootFolder     = "boot"
	IniFilename    = "bgm.ini"
	SocketFilename = "bgm.sock"
	LogFilename    = "bgm.log"
	CoreNameName   = "CORENAME"
	MenuCore       = config.MenuCore
)

// Paths holds every filesystem location BGM touches so tests and the --root
// development mode can relocate the whole application.
type Paths struct {
	MusicFolder   string
	BootFolder    string
	IniFile       string
	SocketFile    string
	CoreNameFile  string
	LogFile       string
	StartupScript string
	CmdInterface  string
	TempFolder    string
	AppPath       string
}

// DefaultPaths returns the locations used on a real MiSTer.
func DefaultPaths() Paths {
	paths := pathsUnder(config.SdFolder, config.TempFolder)
	paths.CoreNameFile = config.CoreNameFile
	paths.StartupScript = config.StartupFile
	paths.CmdInterface = config.CmdInterface
	paths.AppPath = executablePath(filepath.Join(config.ScriptsFolder, AppFilename))
	return paths
}

// RootedPaths relocates every location under root, using a regular file in
// place of the MiSTer command device.
func RootedPaths(root string) Paths {
	paths := pathsUnder(root, filepath.Join(root, "tmp"))
	paths.CoreNameFile = filepath.Join(paths.TempFolder, CoreNameName)
	paths.StartupScript = filepath.Join(root, "linux", "user-startup.sh")
	paths.CmdInterface = filepath.Join(root, "dev", "MiSTer_cmd")
	paths.AppPath = executablePath(filepath.Join(root, "Scripts", AppFilename))
	return paths
}

func pathsUnder(sdRoot, tempFolder string) Paths {
	music := filepath.Join(sdRoot, MusicFolder)
	return Paths{
		MusicFolder: music,
		BootFolder:  filepath.Join(music, BootFolder),
		IniFile:     filepath.Join(music, IniFilename),
		SocketFile:  filepath.Join(tempFolder, SocketFilename),
		LogFile:     filepath.Join(tempFolder, LogFilename),
		TempFolder:  tempFolder,
	}
}

// executablePath mirrors Python's os.path.realpath(__file__).
func executablePath(fallback string) string {
	exe, err := os.Executable()
	if err != nil {
		return fallback
	}
	if resolved, resolveErr := filepath.EvalSymlinks(exe); resolveErr == nil {
		return resolved
	}
	return exe
}
