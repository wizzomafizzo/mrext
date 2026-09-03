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
	"os"

	"github.com/wizzomafizzo/mrext/pkg/config"
	"github.com/wizzomafizzo/mrext/pkg/metadata"
	"github.com/wizzomafizzo/mrext/pkg/mister"
	"github.com/wizzomafizzo/mrext/pkg/service"
	"github.com/wizzomafizzo/mrext/pkg/tracker"
	"github.com/wizzomafizzo/mrext/pkg/utils"
)

func setupSSHKeys(logger *service.Logger, cfg *config.UserConfig) {
	if !cfg.Remote.SyncSSHKeys {
		return
	}

	userFile, err := os.Stat(config.UserSSHKeysFile)
	userFileExists := err == nil

	authFile, err := os.Stat(config.SSHKeysFile)
	authFileExists := err == nil

	switch {
	case !authFileExists && !userFileExists:
		sshFolder, err := os.Stat(config.SSHConfigFolder)
		if err != nil {
			return
		}

		if sshFolder.Mode().Perm() != 0o700 {
			err := mister.FixRootSSHPerms()
			if err != nil {
				logger.Error("failed to fix root ssh perms: %s", err)
			} else {
				logger.Info("fixed root ssh perms")
			}
		}
	case authFileExists && !userFileExists:
		err := mister.CopyAndFixSSHKeys(true)
		if err != nil {
			logger.Error("failed to copy system ssh keys to user: %s", err)
		} else {
			logger.Info("backed up system ssh keys to linux folder")
		}
	case !authFileExists && userFileExists:
		err := mister.CopyAndFixSSHKeys(false)
		if err != nil {
			logger.Error("failed to copy user ssh keys to system: %s", err)
		} else {
			logger.Info("installed user ssh keys to system folder")
		}
	case userFile.ModTime().After(authFile.ModTime()):
		err := mister.CopyAndFixSSHKeys(false)
		if err != nil {
			logger.Error("failed to copy user ssh keys to system: %s", err)
		} else {
			logger.Info("installed updated user ssh keys to system folder")
		}
	case authFile.ModTime().After(userFile.ModTime()):
		err := mister.CopyAndFixSSHKeys(true)
		if err != nil {
			logger.Error("failed to copy system ssh keys to user: %s", err)
		} else {
			logger.Info("backed up updated system ssh keys to linux folder")
		}
	}
}

func runStartupTasks(logger *service.Logger, cfg *config.UserConfig, trk *tracker.Tracker) {
	setupSSHKeys(logger, cfg)

	go func() {
		haveInternet := utils.WaitForInternet(30)
		if !haveInternet {
			logger.Info("no internet connection, skipping network tasks")
			return
		}

		arcadeDbUpdated, err := metadata.UpdateArcadeDB()
		if err != nil {
			logger.Error("failed to download arcade database: %s", err)
		}

		if arcadeDbUpdated {
			logger.Info("arcade database updated")
			trk.ReloadNameMap()
		} else {
			logger.Info("arcade database is up to date")
		}

		m, err := metadata.ReadArcadeDB()
		if err != nil {
			logger.Error("failed to read arcade database: %s", err)
		} else {
			logger.Info("arcade database has %d entries", len(m))
		}
	}()
}
