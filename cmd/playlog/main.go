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
	"github.com/wizzomafizzo/mrext/pkg/mister"
	"github.com/wizzomafizzo/mrext/pkg/service"
	"github.com/wizzomafizzo/mrext/pkg/tracker"
	"github.com/wizzomafizzo/mrext/pkg/utils"
)

// TODO: offer to enable recents option and reboot
// TODO: compatibility with GameEventHub
//       https://github.com/christopher-roelofs/GameEventHub/blob/main/mister.py
// TODO: hashing functions (including inside zips)
// TODO: create example ini file

const appName = "playlog"

func startService(logger *service.Logger, cfg *config.UserConfig) (func() error, error) {
	db, err := openPlayLogDb()
	if err != nil {
		return nil, err
	}

	tr, err := tracker.NewTracker(logger, cfg, db)
	if err != nil {
		logger.Error("error starting tracker: %s", err)
		os.Exit(1)
	}

	tr.LoadCore()
	if !mister.ActiveGameEnabled() {
		if activeErr := mister.SetActiveGame(""); activeErr != nil {
			tr.Logger.Error("error setting active game: %s", activeErr)
		}
	}

	watcher, err := tracker.StartFileWatch(tr)
	if err != nil {
		tr.Logger.Error("error starting file watch: %s", err)
		os.Exit(1)
	}

	interval := 0
	if cfg.PlayLog.SaveEvery > 0 {
		interval = cfg.PlayLog.SaveEvery
	}
	tr.StartTicker(interval)

	return func() error {
		err := watcher.Close()
		if err != nil {
			tr.Logger.Error("error closing file watcher: %s", err)
		}
		tr.StopAll()
		return nil
	}, nil
}

func tryAddStartup() error {
	var startup mister.Startup

	err := startup.Load()
	if err != nil {
		return fmt.Errorf("load startup configuration: %w", err)
	}

	if !startup.Exists("mrext/" + appName) {
		if utils.YesOrNoPrompt("PlayLog must be set to run on MiSTer startup. Add it now?") {
			err = startup.AddService("mrext/" + appName)
			if err != nil {
				return fmt.Errorf("add PlayLog startup service: %w", err)
			}

			err = startup.Save()
			if err != nil {
				return fmt.Errorf("save startup configuration: %w", err)
			}
		}
	}

	return nil
}

func main() {
	svcOpt := flag.String("service", "", "manage playlog service (start, stop, restart, status)")
	flag.Parse()

	logger := service.NewLogger(appName)

	cfg, err := config.LoadUserConfig(appName, &config.UserConfig{
		PlayLog: config.PlayLogConfig{
			SaveEvery: 5, // minutes
		},
	})
	if err != nil {
		logger.Error("error loading user config: %s", err)
		_, _ = fmt.Println("Error loading config:", err)
		os.Exit(1)
	}

	svc, err := service.NewService(service.ServiceArgs{
		Name:   appName,
		Logger: logger,
		Entry: func() (func() error, error) {
			return startService(logger, cfg)
		},
	})
	if err != nil {
		logger.Error("error creating service: %s", err)
		_, _ = fmt.Println("Error creating service:", err)
		os.Exit(1)
	}

	recents, err := mister.RecentsOptionEnabled()
	if err != nil {
		logger.Error("error checking recents option: %s", err)
		_, _ = fmt.Println(
			"Could not read the MiSTer.ini file. " +
				"Make sure the \"recents\" option is enabled if playlog doesn't work.",
		)
	} else if !recents {
		logger.Error("recents option not enabled, exiting...")
		_, _ = fmt.Println(
			"The \"recents\" option must be enabled for playlog to work. " +
				"Configure it in the MiSTer.ini file and reboot.",
		)
		os.Exit(1)
	}

	svc.ServiceHandler(svcOpt)

	err = tryAddStartup()
	if err != nil {
		logger.Error("error adding startup: %s", err)
		_, _ = fmt.Println("Error adding to startup:", err)
	}

	if !svc.Running() {
		if startErr := svc.Start(); startErr != nil {
			logger.Error("error starting service: %s", startErr)
			_, _ = fmt.Println("Error starting service:", startErr)
			os.Exit(1)
		}
	}

	db, err := openPlayLogDb()
	if err != nil {
		logger.Error("error opening db: %s", err)
		_, _ = fmt.Println("Error opening database:", err)
		os.Exit(1)
	}

	cores, err := db.topCores(10)
	if err != nil {
		logger.Error("error getting top cores: %s", err)
		_, _ = fmt.Println("Error getting top cores:", err)
		os.Exit(1)
	}
	maxCoreLen := 0
	for _, core := range cores {
		if len(core.Name) > maxCoreLen {
			maxCoreLen = len(core.Name)
		}
	}

	games, err := db.topGames(10)
	if err != nil {
		logger.Error("error getting top games: %s", err)
		_, _ = fmt.Println("Error getting top games:", err)
		os.Exit(1)
	}
	maxGameLen := 0
	for _, game := range games {
		if len(game.Name) > maxGameLen {
			maxGameLen = len(game.Name)
		}
	}

	_, _ = fmt.Println("Top played cores:")
	// TODO: convert names using names.txt
	for _, core := range cores {
		hours := core.Time / 3600
		minutes := (core.Time % 3600) / 60
		_, _ = fmt.Printf("%-*s  %dh %dm\n", maxCoreLen, core.Name, hours, minutes)
	}
	_, _ = fmt.Println()
	_, _ = fmt.Println("Top played games:")
	for _, game := range games {
		hours := game.Time / 3600
		minutes := (game.Time % 3600) / 60
		_, _ = fmt.Printf("%-*s  %dh %dm\n", maxGameLen, game.Name, hours, minutes)
	}
}
