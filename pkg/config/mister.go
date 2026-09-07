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

// TODO: should this be hardcoded? how common is usb0 setup?
const (
	SdFolder         = "/media/fat"
	CoreConfigFolder = SdFolder + "/config"
	FontFolder       = SdFolder + "/font"
	TempFolder       = "/tmp"
	LinuxFolder      = SdFolder + "/linux"
	ScriptsFolder    = SdFolder + "/Scripts"
	CifsFolder       = SdFolder + "/cifs"
)

const MenuConfigFile = CoreConfigFolder + "/MENU.CFG"

const (
	MisterIniFile     = SdFolder + "/MiSTer.ini"
	MisterIniFileAlt1 = SdFolder + "/MiSTer_alt_1.ini"
	MisterIniFileAlt2 = SdFolder + "/MiSTer_alt_2.ini"
	MisterIniFileAlt3 = SdFolder + "/MiSTer_alt_3.ini"
)

const (
	StartupFile     = LinuxFolder + "/user-startup.sh"
	UBootConfigFile = LinuxFolder + "/u-boot.txt"
)

const (
	CoreNameFile    = TempFolder + "/CORENAME"
	CurrentPathFile = TempFolder + "/CURRENTPATH"
	StartPathFile   = TempFolder + "/STARTPATH"
	FullPathFile    = TempFolder + "/FULLPATH"
)

const CoresRecentFile = CoreConfigFolder + "/cores_recent.cfg"

const MenuCore = "MENU"

// RACoresFolder is the isolated RetroAchievements core directory relative to SD root.
const RACoresFolder = "_RA_Cores/Cores"

const (
	CmdInterface      = "/dev/MiSTer_cmd"
	SSHConfigFolder   = "/root/.ssh"
	SSHKeysFile       = SSHConfigFolder + "/authorized_keys"
	UserSSHKeysFile   = LinuxFolder + "/authorized_keys"
	DownloaderLastRun = ScriptsFolder + "/.config/downloader/downloader.last_successful_run"
)

// TODO: this can't be hardcoded if we want dynamic arcade folders
const ArcadeCoresFolder = "/media/fat/_Arcade/cores"

// NamesFile is MiSTer's optional display-name mapping, one "Key: Name" per line.
const NamesFile = SdFolder + "/names.txt"

// GamesMenuFolder is the stock-menu folder GamesMenu mirrors game libraries into.
const GamesMenuFolder = SdFolder + "/_Games"

// StorageRoots are the removable and network mount points games can live
// under. Everything below one of these is only reachable while that storage is
// attached; an unmounted mount point is left behind as an empty directory.
// Paths on the SD card are not listed because it is always present.
var StorageRoots = []string{
	"/media/usb0",
	"/media/usb1",
	"/media/usb2",
	"/media/usb3",
	"/media/usb4",
	"/media/usb5",
	"/media/network",
	"/media/fat/cifs",
}

// TODO: not the order mister actually checks, it does games folders second, but this is simpler for checking prefix
var GamesFolders = []string{
	"/media/usb0/games",
	"/media/usb0",
	"/media/usb1/games",
	"/media/usb1",
	"/media/usb2/games",
	"/media/usb2",
	"/media/usb3/games",
	"/media/usb3",
	"/media/usb4/games",
	"/media/usb4",
	"/media/usb5/games",
	"/media/usb5",
	"/media/network/games",
	"/media/network",
	"/media/fat/cifs/games",
	"/media/fat/cifs",
	"/media/fat/games",
	"/media/fat",
}
