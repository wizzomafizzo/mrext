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

// Package version reports which build of an mrext app is running.
package version

import (
	"runtime/debug"
	"strings"
)

// Version is stamped at build time with -ldflags "-X .../pkg/version.Version=...".
// Left empty for a plain "go build", which falls back to the VCS stamps Go
// records automatically.
var Version = ""

// Commit is stamped the same way. Empty falls back to vcs.revision.
var Commit = ""

const unknown = "unknown"

// String renders "<version> (<commit>)" for a binary, or as much of it as is
// known. Apps print this for -version and show it in their title bar, so a user
// can say which build they are on without guessing from a file date.
func String() string {
	version, commit := Version, Commit
	if version == "" || commit == "" {
		buildVersion, buildCommit := fromBuildInfo()
		if version == "" {
			version = buildVersion
		}
		if commit == "" {
			commit = buildCommit
		}
	}
	if version == "" {
		version = unknown
	}
	if commit == "" {
		return version
	}
	return version + " (" + commit + ")"
}

// Short renders just the version, for a title bar where the commit does not fit.
func Short() string {
	if Version != "" {
		return Version
	}
	if buildVersion, _ := fromBuildInfo(); buildVersion != "" {
		return buildVersion
	}
	return unknown
}

func fromBuildInfo() (version, commit string) {
	info, ok := debug.ReadBuildInfo()
	if !ok {
		return "", ""
	}
	if info.Main.Version != "" && info.Main.Version != "(devel)" {
		version = info.Main.Version
	}
	modified := false
	for _, setting := range info.Settings {
		switch setting.Key {
		case "vcs.revision":
			commit = setting.Value
			if len(commit) > 7 {
				commit = commit[:7]
			}
		case "vcs.modified":
			modified = setting.Value == "true"
		}
	}
	if modified && commit != "" {
		commit += "-dirty"
	}
	if version == "" && commit != "" {
		version = "dev"
	}
	return version, strings.TrimSpace(commit)
}
