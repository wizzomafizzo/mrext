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

package config

const (
	UserConfigEnv  = "MREXT_CONFIG"
	UserAppPathEnv = "MREXT_APP_PATH"
)

const (
	ActiveGameFile = TempFolder + "/ACTIVEGAME"
	SearchDbFile   = SdFolder + "/search.db"
	PlayLogDbFile  = SdFolder + "/playlog.db"
)

const (
	PidFileTemplate = TempFolder + "/%s.pid"
	LogFileTemplate = TempFolder + "/%s.log"
)

const (
	ScriptsConfigFolder = ScriptsFolder + "/.config"
	MrextConfigFolder   = ScriptsConfigFolder + "/mrext"
)

// SharedTUIConfigFile holds [tui] settings every mrext app falls back to, so a
// theme can be chosen once instead of in each app's own INI. Optional: when it
// is absent nothing changes. An app's own [tui] section still wins.
const SharedTUIConfigFile = MrextConfigFolder + "/tui.ini"

// SharedTUIConfigEnv overrides SharedTUIConfigFile, so tests never read or
// write the real one.
const SharedTUIConfigEnv = "MREXT_TUI_CONFIG"

const (
	// The database used to be a folder of dated ArcadeDatabaseYYMMDD.csv
	// files. It is now a single file at the repository root, and the old
	// folder is gone, so the previous URL answers 404 and every MiSTer
	// silently loses arcade game names.
	ArcadeDBURL  = "https://api.github.com/repos/MiSTer-devel/ArcadeDatabase_MiSTer/contents/ArcadeDatabase.csv"
	ArcadeDBFile = MrextConfigFolder + "/ArcadeDatabase.csv"
)

const GamesDB = ScriptsConfigFolder + "/mrext/games.db"

// GamesDBLockSuffix names the persistent writer lock beside the shared index.
const GamesDBLockSuffix = ".lock"

const LastLaunchFile = SdFolder + "/.LASTLAUNCH.mgl"

const NeoGeoRomsetsFile = "romsets.xml"

// BGMDefaultBootFolder holds opt-in core boot sounds under BGM's boot directory.
const BGMDefaultBootFolder = "default"

const SAMActivityFile = TempFolder + "/.SAM_tmp/SAM_Joy_Activity"

const (
	ServiceProcFolder       = "/proc"
	RemoteWatchAttempts     = 6
	RemoteWatchRetrySeconds = 1
)
