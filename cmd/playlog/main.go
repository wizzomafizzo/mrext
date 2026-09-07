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
	"github.com/wizzomafizzo/mrext/pkg/version"
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

	watcher, err := tracker.StartFileWatchWithRetry(tr)
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

func main() {
	svcOpt := flag.String("service", "", "manage playlog service (start, stop, restart, status)")
	showVersion := flag.Bool("version", false, "print the version and exit")
	flag.Parse()
	if *showVersion {
		_, _ = fmt.Printf("%s %s\n", "playlog", version.String())
		return
	}

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

	// -service is how user-startup.sh and the shell invoke this, and it must
	// stay headless. ServiceHandler exits for every command it recognises.
	svc.ServiceHandler(svcOpt)

	if !svc.Running() {
		if startErr := svc.Start(); startErr != nil {
			logger.Error("error starting service: %s", startErr)
			_, _ = fmt.Println("Error starting service:", startErr)
			os.Exit(1)
		}
	}

	// Interactive launch shows the service screen, with the play-time report
	// on its own page instead of printed and scrolled past.
	if screenErr := showServiceScreen(svc, cfg); screenErr != nil {
		logger.Error("error showing service screen: %s", screenErr)
		_, _ = fmt.Println(screenErr)
		os.Exit(1)
	}
}
