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

const (
	ArcadeDBURL  = "https://api.github.com/repositories/521644036/contents/ArcadeDatabase_CSV"
	ArcadeDBFile = MrextConfigFolder + "/ArcadeDatabase.csv"
)

const GamesDB = ScriptsConfigFolder + "/mrext/games.db"

const LastLaunchFile = SdFolder + "/.LASTLAUNCH.mgl"
