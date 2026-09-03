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

package tracker

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/fsnotify/fsnotify"
	"github.com/wizzomafizzo/mrext/pkg/config"
	"github.com/wizzomafizzo/mrext/pkg/mister"
)

// Read a core's recent file and attempt to write the newest entry's
// launch-able path to ACTIVEGAME.
func loadRecent(filename string) error {
	if !strings.Contains(filename, "_recent") {
		return nil
	}

	recents, err := mister.ReadRecent(filename)
	if err != nil {
		return fmt.Errorf("error reading recent file: %w", err)
	}
	if len(recents) == 0 {
		return nil
	}

	newest := recents[0]

	if strings.HasSuffix(filename, "cores_recent.cfg") {
		// main menu's recent file, written when launching mgls
		if strings.HasSuffix(strings.ToLower(newest.Name), ".mgl") {
			mglPath := mister.ResolvePath(filepath.Join(newest.Directory, newest.Name))
			mgl, mglErr := mister.ReadMGL(mglPath)
			if mglErr != nil {
				return fmt.Errorf("error reading mgl file: %w", mglErr)
			}

			err = mister.SetActiveGame(mgl.File.Path)
			if err != nil {
				return fmt.Errorf("error setting active game: %w", err)
			}
		}
	} else {
		// individual core's recent file
		err = mister.SetActiveGame(filepath.Join(newest.Directory, newest.Name))
		if err != nil {
			return fmt.Errorf("error setting active game: %w", err)
		}
	}

	return nil
}

// StartFileWatch Start thread for monitoring changes to all files relating to core/game launches.
func StartFileWatch(tr *Tracker) (*fsnotify.Watcher, error) {
	tr.Logger.Info("starting file watcher")

	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, fmt.Errorf("create file watcher: %w", err)
	}

	closeOnError := func(watchErr error) (*fsnotify.Watcher, error) {
		if closeErr := watcher.Close(); closeErr != nil {
			tr.Logger.Error("error closing failed file watcher: %s", closeErr)
		}
		return nil, watchErr
	}

	if err = watcher.Add(config.CoreNameFile); err != nil {
		return closeOnError(fmt.Errorf("watch core name: %w", err))
	}
	if err = watcher.Add(config.CoreConfigFolder); err != nil {
		return closeOnError(fmt.Errorf("watch core configuration: %w", err))
	}
	if err = watcher.Add(config.ActiveGameFile); err != nil {
		return closeOnError(fmt.Errorf("watch active game: %w", err))
	}
	if _, statErr := os.Stat(config.CurrentPathFile); statErr == nil {
		if err = watcher.Add(config.CurrentPathFile); err != nil {
			return closeOnError(fmt.Errorf("watch current menu path: %w", err))
		}
	}

	go func() {
		for {
			select {
			case event, ok := <-watcher.Events:
				if !ok {
					return
				}
				if event.Op&fsnotify.Write == fsnotify.Write {
					switch {
					case event.Name == config.CurrentPathFile:
						tr.trackMenu()
					case event.Name == config.CoreNameFile:
						tr.LoadCore()
					case event.Name == config.ActiveGameFile:
						tr.loadGame()
					case strings.HasPrefix(event.Name, config.CoreConfigFolder):
						if recentErr := loadRecent(event.Name); recentErr != nil {
							tr.Logger.Error("error loading recent file: %s", recentErr)
						}
					}
				}
			case err, ok := <-watcher.Errors:
				if !ok {
					return
				}
				tr.Logger.Error("error in watcher: %s", err)
			}
		}
	}()

	return watcher, nil
}
