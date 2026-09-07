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
	"time"

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

	watch := &fileWatch{
		watcher: watcher,
		logger:  tr.Logger,
		files: map[string]func(){
			config.CurrentPathFile: tr.trackMenu,
			config.CoreNameFile:    tr.LoadCore,
			config.ActiveGameFile:  tr.loadGame,
		},
		folder: config.CoreConfigFolder,
		onFolderEvent: func(name string) {
			if recentErr := loadRecent(name); recentErr != nil {
				tr.Logger.Error("error loading recent file: %s", recentErr)
			}
		},
	}
	go watch.run()

	return watcher, nil
}

const (
	rewatchAttempts = 5
	rewatchDelay    = 200 * time.Millisecond
)

// fileWatch dispatches fsnotify events. It is separate from StartFileWatch so
// tests can drive it against temporary paths.
//
//nolint:govet // Field order groups the watch targets with their handlers.
type fileWatch struct {
	watcher       *fsnotify.Watcher
	logger        trackerLogger
	files         map[string]func()
	folder        string
	onFolderEvent func(string)
	// attempts and delay bound the re-watch retry; zero selects the defaults.
	attempts int
	delay    time.Duration
}

func (f *fileWatch) run() {
	for {
		select {
		case event, ok := <-f.watcher.Events:
			if !ok {
				return
			}
			f.dispatch(event)
		case err, ok := <-f.watcher.Errors:
			if !ok {
				return
			}
			f.logger.Error("error in watcher: %s", err)
		}
	}
}

func (f *fileWatch) dispatch(event fsnotify.Event) {
	// Create matters as much as Write: MiSTer tooling and other scripts
	// replace these files rather than truncating them, and the first content
	// of a newly created file arrives as Create.
	if event.Op&(fsnotify.Write|fsnotify.Create) != 0 {
		if handle, ok := f.files[event.Name]; ok {
			handle()
			return
		}
		if f.folder != "" && strings.HasPrefix(event.Name, f.folder) {
			f.onFolderEvent(event.Name)
		}
		return
	}
	// inotify watches follow the inode, so a file that is removed and
	// recreated leaves the watch on a dead one and the tracker goes silent
	// until the service restarts. Re-add the path and re-read it.
	if event.Op&(fsnotify.Remove|fsnotify.Rename) != 0 {
		if _, ok := f.files[event.Name]; ok {
			go f.rewatch(event.Name)
		}
	}
}

func (f *fileWatch) rewatch(name string) {
	attempts, delay := f.attempts, f.delay
	if attempts <= 0 {
		attempts = rewatchAttempts
	}
	if delay <= 0 {
		delay = rewatchDelay
	}
	for attempt := range attempts {
		if err := f.watcher.Add(name); err == nil {
			// The replacement may already hold new content.
			if handle, ok := f.files[name]; ok {
				handle()
			}
			return
		}
		if attempt < attempts-1 {
			time.Sleep(delay)
		}
	}
	f.logger.Error("gave up re-watching %s after %d attempts", name, attempts)
}
