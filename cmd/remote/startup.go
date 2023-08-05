package main

import (
	"github.com/wizzomafizzo/mrext/pkg/config"
	"github.com/wizzomafizzo/mrext/pkg/metadata"
	"github.com/wizzomafizzo/mrext/pkg/mister"
	"github.com/wizzomafizzo/mrext/pkg/service"
	"github.com/wizzomafizzo/mrext/pkg/tracker"
	"github.com/wizzomafizzo/mrext/pkg/utils"
	"os"
)

func setupSSHKeys(logger *service.Logger, cfg *config.UserConfig) {
	if !cfg.Remote.SyncSSHKeys {
		return
	}

	userFile, err := os.Stat(config.UserSSHKeysFile)
	userFileExists := false
	if err == nil {
		userFileExists = true
	}

	authFile, err := os.Stat(config.SSHKeysFile)
	authFileExists := false
	if err == nil {
		authFileExists = true
	}

	if !authFileExists && !userFileExists {
		sshFolder, err := os.Stat(config.SSHConfigFolder)
		if err != nil {
			return
		}

		if sshFolder.Mode().Perm() != 0700 {
			err := mister.FixRootSSHPerms()
			if err != nil {
				logger.Error("failed to fix root ssh perms: %s", err)
			} else {
				logger.Info("fixed root ssh perms")
			}
		}
	} else if authFileExists && !userFileExists {
		err := mister.CopyAndFixSSHKeys(true)
		if err != nil {
			logger.Error("failed to copy system ssh keys to user: %s", err)
		} else {
			logger.Info("backed up system ssh keys to linux folder")
		}
	} else if !authFileExists && userFileExists {
		err := mister.CopyAndFixSSHKeys(false)
		if err != nil {
			logger.Error("failed to copy user ssh keys to system: %s", err)
		} else {
			logger.Info("installed user ssh keys to system folder")
		}
	} else if userFile.ModTime().After(authFile.ModTime()) {
		err := mister.CopyAndFixSSHKeys(false)
		if err != nil {
			logger.Error("failed to copy user ssh keys to system: %s", err)
		} else {
			logger.Info("installed updated user ssh keys to system folder")
		}
	} else if authFile.ModTime().After(userFile.ModTime()) {
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

		arcadeDbUpdated, err := metadata.UpdateArcadeDb()
		if err != nil {
			logger.Error("failed to download arcade database: %s", err)
		}

		if arcadeDbUpdated {
			logger.Info("arcade database updated")
			trk.ReloadNameMap()
		} else {
			logger.Info("arcade database is up to date")
		}

		m, err := metadata.ReadArcadeDb()
		if err != nil {
			logger.Error("failed to read arcade database: %s", err)
		} else {
			logger.Info("arcade database has %d entries", len(m))
		}
	}()
}
