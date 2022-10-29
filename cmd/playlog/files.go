package main

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
// launchable path to ACTIVEGAME.
func loadRecent(tr *tracker, filename string) error {
	if !strings.Contains(filename, "_recent") {
		return nil
	}

	// tr.logger.Info("loading recent file: %s", filename)

	file, err := os.Open(filename)
	if err != nil {
		return fmt.Errorf("error opening game file: %w", err)
	}
	defer file.Close()

	recents, err := mister.ReadRecent(filename)
	if err != nil {
		return fmt.Errorf("error reading recent file: %w", err)
	} else if len(recents) == 0 {
		return nil
	}

	newest := recents[0]

	if strings.HasSuffix(filename, "cores_recent.cfg") {
		// main menu's recent file, written when launching mgls
		if strings.HasSuffix(strings.ToLower(newest.Name), ".mgl") {
			mglPath := mister.ResolvePath(filepath.Join(newest.Directory, newest.Name))
			mgl, err := mister.ReadMgl(mglPath)
			if err != nil {
				return fmt.Errorf("error reading mgl file: %w", err)
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

// Start thread for monitoring changes to all files relating to core/game launches.
func startFileWatch(tr *tracker) (*fsnotify.Watcher, error) {
	tr.logger.Info("starting file watcher")

	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, err
	}

	go func() {
		for {
			select {
			case event, ok := <-watcher.Events:
				if !ok {
					return
				}
				if event.Op&fsnotify.Write == fsnotify.Write {
					if event.Name == config.CoreNameFile {
						tr.loadCore()
					} else if event.Name == config.ActiveGameFile {
						tr.loadGame()
					} else if strings.HasPrefix(event.Name, config.CoreConfigFolder) {
						err = loadRecent(tr, event.Name)
						if err != nil {
							tr.logger.Error("error loading recent file: %s", err)
						}
					}
				}
			case err, ok := <-watcher.Errors:
				if !ok {
					return
				}
				tr.logger.Error("error in watcher: %s", err)
			}
		}
	}()

	err = watcher.Add(config.CoreNameFile)
	if err != nil {
		return nil, err
	}

	err = watcher.Add(config.CoreConfigFolder)
	if err != nil {
		return nil, err
	}

	err = watcher.Add(config.ActiveGameFile)
	if err != nil {
		return nil, err
	}

	return watcher, nil
}
