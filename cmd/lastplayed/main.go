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
	"github.com/wizzomafizzo/mrext/pkg/version"
)

const (
	appName                 = "lastplayed"
	defaultLastPlayedName   = "Last Played"
	defaultRecentFolderName = "Recently Played"
	maxRecentEntries        = 99
)

func launcherSystem(cfg *config.UserConfig, path string) (games.System, error) {
	if strings.TrimSpace(path) == "" {
		return games.System{}, errors.New("launcher target is empty")
	}
	if strings.EqualFold(filepath.Ext(path), ".mra") {
		info, err := os.Stat(path)
		if err != nil {
			return games.System{}, fmt.Errorf("inspect arcade target: %w", err)
		}
		if !info.Mode().IsRegular() {
			return games.System{}, errors.New("arcade target is not a regular file")
		}
		system, err := games.GetSystem(tracker.ArcadeSystem)
		if err != nil {
			return games.System{}, fmt.Errorf("get arcade system: %w", err)
		}
		return *system, nil
	}
	system, err := games.BestSystemMatch(cfg, path)
	if err != nil {
		return games.System{}, fmt.Errorf("find launcher system: %w", err)
	}
	return system, nil
}

func createLastPlayedMgl(cfg *config.UserConfig, path, sdRoot string) error {
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

	system, err := launcherSystem(cfg, path)
	if err != nil {
		return err
	}

	created, err := mister.CreateLauncher(cfg, &system, path, sdRoot, mglName)
	if err != nil {
		return fmt.Errorf("error creating mgl: %w", err)
	}

	return removeStaleLastPlayed(sdRoot, mglName, created)
}

// Arcade launchers are `.mra` links and everything else is a `.mgl` file, so a
// system change leaves the previous extension behind. Remove it: a stale
// shortcut keeps appearing in the menu and can boot the wrong game.
func removeStaleLastPlayed(sdRoot, name, created string) error {
	for _, extension := range []string{".mgl", ".mra"} {
		stale := filepath.Join(sdRoot, name+extension)
		if stale == created {
			continue
		}
		if err := os.Remove(stale); err != nil && !os.IsNotExist(err) {
			return fmt.Errorf("remove stale last played launcher: %w", err)
		}
	}

	return nil
}

type recentFile struct {
	Modified    time.Time
	Path        string
	Filename    string
	NewFilename string
}

func addToRecentFolder(cfg *config.UserConfig, path, sdRoot string) error {
	system, err := launcherSystem(cfg, path)
	if err != nil {
		return err
	}
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

	recentPath := filepath.Join(sdRoot, "_"+recentFolderName)

	if _, statErr := os.Stat(recentPath); os.IsNotExist(statErr) {
		// #nosec G301 -- MiSTer menu folder must remain world-readable.
		if mkdirErr := os.Mkdir(recentPath, 0o755); mkdirErr != nil {
			return fmt.Errorf("error creating recent folder: %w", mkdirErr)
		}
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
		extension := strings.ToLower(filepath.Ext(file.Name()))
		if file.IsDir() || (extension != ".mgl" && extension != ".mra") {
			continue
		}
		// Only number entries owned by LastPlayed, not unrelated menu files.
		if len(file.Name()) < 4 || file.Name()[2] != ' ' {
			continue
		}
		if _, parseErr := strconv.Atoi(file.Name()[:2]); parseErr != nil {
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
		if recentFiles[i].Modified.Equal(recentFiles[j].Modified) {
			return recentFiles[i].Filename < recentFiles[j].Filename
		}
		return recentFiles[i].Modified.After(recentFiles[j].Modified)
	})

	knownFiles := make(map[string]bool)

	i := 0
	retained := make([]recentFile, 0, len(recentFiles))
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

		file.NewFilename = fmt.Sprintf("%0*d %s", prefixLength, i+1, filename)
		retained = append(retained, file)
		i++
	}

	// Remove duplicates before renumbering: otherwise an old duplicate's path
	// can refer to the freshly renamed launcher when it is deleted.
	for _, file := range retained {
		newPath := filepath.Join(recentPath, file.NewFilename)
		if renameErr := os.Rename(file.Path, newPath); renameErr != nil {
			return fmt.Errorf("error renaming recent file: %w", renameErr)
		}
	}

	return nil
}

type fakeDb struct {
	config *config.UserConfig
	sdRoot string
}

func (*fakeDb) FixPowerLoss() (bool, error) {
	return false, nil
}

func (f *fakeDb) AddEvent(ev *tracker.EventAction) error {
	if ev.Action != tracker.EventActionGameStart || strings.TrimSpace(ev.TargetPath) == "" {
		return nil
	}
	root := f.sdRoot
	if root == "" {
		root = config.SdFolder
	}

	if !f.config.LastPlayed.DisableLastPlayed {
		err := createLastPlayedMgl(f.config, ev.TargetPath, root)
		if err != nil {
			return fmt.Errorf("error creating last played mgl: %w", err)
		}
	}

	if !f.config.LastPlayed.DisableRecentFolder {
		err := addToRecentFolder(f.config, ev.TargetPath, root)
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

	watcher, err := tracker.StartFileWatchWithRetry(tr)
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

// startupInstalled reports whether the boot entry is present.
func startupInstalled() (bool, error) {
	var startup mister.Startup
	if err := startup.Load(); err != nil {
		return false, fmt.Errorf("load startup configuration: %w", err)
	}
	return startup.Exists("mrext/" + appName), nil
}

// addToStartup installs the boot entry. The question that used to gate this
// was a raw-terminal arrow-key prompt; it is now the shared confirm dialog, so
// it looks and behaves like every other question the apps ask.
func addToStartup() error {
	var startup mister.Startup
	if err := startup.Load(); err != nil {
		return fmt.Errorf("load startup configuration: %w", err)
	}
	if startup.Exists("mrext/" + appName) {
		return nil
	}
	if err := startup.AddService("mrext/" + appName); err != nil {
		return fmt.Errorf("add LastPlayed startup service: %w", err)
	}
	if err := startup.Save(); err != nil {
		return fmt.Errorf("save startup configuration: %w", err)
	}
	return nil
}

func main() {
	svcOpt := flag.String("service", "", "manage lastplayed service (start, stop, restart, status)")
	showVersion := flag.Bool("version", false, "print the version and exit")
	flag.Parse()
	if *showVersion {
		_, _ = fmt.Printf("%s %s\n", "lastplayed", version.String())
		return
	}

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

	// Interactive launch from the Scripts menu shows the service screen rather
	// than printing one line and exiting.
	if screenErr := showServiceScreen(svc, cfg); screenErr != nil {
		logger.Error("error showing service screen: %s", screenErr)
		_, _ = fmt.Println(screenErr)
		os.Exit(1)
	}
	os.Exit(0)
}
