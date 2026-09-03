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
	"errors"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/wizzomafizzo/mrext/pkg/config"
	"github.com/wizzomafizzo/mrext/pkg/games"
	"github.com/wizzomafizzo/mrext/pkg/mister"
	"github.com/wizzomafizzo/mrext/pkg/service"
	"github.com/wizzomafizzo/mrext/pkg/tracker"
	"github.com/wizzomafizzo/mrext/pkg/utils"
)

const (
	appName                 = "lastplayed"
	defaultLastPlayedName   = "Last Played"
	defaultRecentFolderName = "Recently Played"
	maxRecentEntries        = 99
)

func createLastPlayedMgl(cfg *config.UserConfig, path string) error {
	var mglName string

	switch {
	case cfg.LastPlayed.LastPlayedName != "":
		mglName = cfg.LastPlayed.LastPlayedName
	case cfg.LastPlayed.Name != "":
		mglName = cfg.LastPlayed.Name
	default:
		mglName = defaultLastPlayedName
	}

	mglName = utils.StripBadFileChars(mglName)

	if mglName == "" {
		return errors.New("name cannot be empty")
	}

	system, err := games.BestSystemMatch(cfg, path)
	if err != nil {
		return fmt.Errorf("no system match found: %s", path)
	}

	_, err = mister.CreateLauncher(cfg, &system, path, config.SdFolder, mglName)
	if err != nil {
		return fmt.Errorf("error creating mgl: %w", err)
	}

	return nil
}

type recentFile struct {
	Modified    time.Time
	Path        string
	Filename    string
	NewFilename string
}

func addToRecentFolder(cfg *config.UserConfig, path string) error {
	var recentFolderName string

	if cfg.LastPlayed.RecentFolderName == "" {
		recentFolderName = defaultRecentFolderName
	} else {
		recentFolderName = cfg.LastPlayed.RecentFolderName
	}

	recentFolderName = utils.StripBadFileChars(recentFolderName)

	if recentFolderName == "" {
		return errors.New("name cannot be empty")
	}

	recentPath := filepath.Join(config.SdFolder, "_"+recentFolderName)

	if _, err := os.Stat(recentPath); os.IsNotExist(err) {
		// #nosec G301 -- MiSTer menu folder must remain world-readable.
		err = os.Mkdir(recentPath, 0o755)
		if err != nil {
			return fmt.Errorf("error creating recent folder: %w", err)
		}
	}

	system, err := games.BestSystemMatch(cfg, path)
	if err != nil {
		return fmt.Errorf("no system match found: %s", path)
	}

	mglName := filepath.Base(path)
	mglName = strings.TrimSuffix(mglName, filepath.Ext(mglName))
	mglName = utils.StripBadFileChars(mglName)
	mglName = fmt.Sprintf("00 %s [%s]", mglName, system.Name)

	_, err = mister.CreateLauncher(cfg, &system, path, recentPath, mglName)
	if err != nil {
		return fmt.Errorf("error creating mgl: %w", err)
	}

	recentFolder, err := os.ReadDir(recentPath)
	if err != nil {
		return fmt.Errorf("error reading recent folder: %w", err)
	}

	var recentFiles []recentFile
	for _, file := range recentFolder {
		if file.IsDir() || filepath.Ext(strings.ToLower(file.Name())) != ".mgl" {
			continue
		}

		info, err := file.Info()
		if err != nil {
			continue
		}

		recentFiles = append(recentFiles, recentFile{
			Path:     filepath.Join(recentPath, file.Name()),
			Filename: file.Name(),
			Modified: info.ModTime(),
		})
	}

	prefixLength := len(strconv.Itoa(maxRecentEntries))

	sort.Slice(recentFiles, func(i, j int) bool {
		return recentFiles[i].Modified.After(recentFiles[j].Modified)
	})

	knownFiles := make(map[string]bool)

	i := 0
	for _, file := range recentFiles {
		if i >= maxRecentEntries {
			err := os.Remove(file.Path)
			if err != nil {
				return fmt.Errorf("error removing recent file: %w", err)
			}
			continue
		}

		filename := file.Filename[prefixLength+1:]

		if knownFiles[filename] {
			err := os.Remove(file.Path)
			if err != nil {
				return fmt.Errorf("error removing recent file: %w", err)
			}
			continue
		}
		knownFiles[filename] = true

		newFilename := fmt.Sprintf("%0*d %s", prefixLength, i+1, filename)
		newPath := filepath.Join(recentPath, newFilename)
		err := os.Rename(file.Path, newPath)
		if err != nil {
			return fmt.Errorf("error renaming recent file: %w", err)
		}

		i++
	}

	return nil
}

type fakeDb struct {
	config *config.UserConfig
}

func (*fakeDb) FixPowerLoss() (bool, error) {
	return false, nil
}

func (f *fakeDb) AddEvent(ev *tracker.EventAction) error {
	if ev.Action != tracker.EventActionGameStart {
		return nil
	}

	if !f.config.LastPlayed.DisableLastPlayed {
		err := createLastPlayedMgl(f.config, ev.TargetPath)
		if err != nil {
			return fmt.Errorf("error creating last played mgl: %w", err)
		}
	}

	if !f.config.LastPlayed.DisableRecentFolder {
		err := addToRecentFolder(f.config, ev.TargetPath)
		if err != nil {
			return fmt.Errorf("error adding to recent folder: %w", err)
		}
	}

	return nil
}

func (*fakeDb) UpdateCore(_ tracker.CoreTime) error {
	return nil
}

func (*fakeDb) GetCore(_ string) (tracker.CoreTime, error) {
	return tracker.CoreTime{}, nil
}

func (*fakeDb) UpdateGame(_ tracker.GameTime) error {
	return nil
}

func (*fakeDb) GetGame(_ string) (tracker.GameTime, error) {
	return tracker.GameTime{}, nil
}

func (*fakeDb) NoResults(_ error) bool {
	return true
}

func startService(logger *service.Logger, cfg *config.UserConfig) (func() error, error) {
	tr, err := tracker.NewTracker(logger, cfg, &fakeDb{
		config: cfg,
	})
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
		if utils.YesOrNoPrompt("LastPlayed must be set to run on MiSTer startup. Add it now?") {
			err = startup.AddService("mrext/" + appName)
			if err != nil {
				return fmt.Errorf("add LastPlayed startup service: %w", err)
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
		LastPlayed: config.LastPlayedConfig{
			LastPlayedName:      defaultLastPlayedName,
			DisableLastPlayed:   false,
			RecentFolderName:    defaultRecentFolderName,
			DisableRecentFolder: false,
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
				"Make sure the \"recents\" option is enabled if lastplayed doesn't work.",
		)
	} else if !recents {
		logger.Error("recents option not enabled, exiting...")
		_, _ = fmt.Println("The \"recents\" option must be enabled for lastplayed to work.")
		_, _ = fmt.Println("Configure it in the MiSTer.ini file and run lastplayed again.")
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
		_, _ = fmt.Println("Service started successfully.")
		os.Exit(0)
	}

	_, _ = fmt.Println("Service is running.")
	os.Exit(0)
}
