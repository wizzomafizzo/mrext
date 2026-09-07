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

//go:build linux

package service

import (
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/wizzomafizzo/mrext/pkg/config"
)

func (*Service) matchesDaemon(pid int) bool {
	path := os.Getenv(config.UserAppPathEnv)
	if path == "" {
		var err error
		path, err = os.Executable()
		if err != nil {
			return false
		}
	}
	return daemonIdentity(config.ServiceProcFolder, pid, filepath.Join(config.TempFolder, filepath.Base(path)))
}

// A live PID alone is not proof of ownership: it may have been recycled after
// a crash. Require the copied executable and the daemon subcommand. Keep the
// numeric PID-file format compatible with installed scripts and other apps.
func daemonIdentity(procRoot string, pid int, executable string) bool {
	if pid <= 1 {
		return false
	}
	root := filepath.Join(procRoot, strconv.Itoa(pid))
	actual, err := os.Readlink(filepath.Join(root, "exe"))
	if err != nil || strings.TrimSuffix(actual, " (deleted)") != executable {
		return false
	}
	// #nosec G304 -- fixed proc path in production; injected temporary root in tests.
	data, err := os.ReadFile(filepath.Join(root, "cmdline"))
	if err != nil {
		return false
	}
	args := strings.Split(string(data), "\x00")
	return len(args) >= 3 && args[1] == "-service" && args[2] == "exec"
}
