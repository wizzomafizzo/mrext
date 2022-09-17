package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/fsnotify/fsnotify"

	"github.com/wizzomafizzo/mrext/pkg/config"
	"github.com/wizzomafizzo/mrext/pkg/mister"
)

// TODO: confirm recent ini setting is on
// TODO: store game as hash
// TODO: read mgl launches from cores_recent
// TODO: handle failed mgl launch
// TODO: ticker interval and save interval should be configurable
// TODO: ignore menu core by default

func setActiveGame(path string) error {
	file, err := os.Create(config.ActiveGameFile)
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = file.WriteString(path)
	if err != nil {
		return err
	}

	return nil
}

func getActiveGame() (string, error) {
	data, err := os.ReadFile(config.ActiveGameFile)
	if err != nil {
		return "", err
	}

	return string(data), nil
}

func loadRecent(filename string) error {
	if !strings.Contains(filename, "_recent") {
		return nil
	}

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
		if strings.HasSuffix(strings.ToLower(newest.Name), ".mgl") {
			mglPath := mister.ResolvePath(filepath.Join(newest.Directory, newest.Name))
			mgl, err := mister.ReadMgl(mglPath)
			if err != nil {
				return fmt.Errorf("error reading mgl file: %w", err)
			}

			err = setActiveGame(mgl.File.Path)
			if err != nil {
				return fmt.Errorf("error setting active game: %w", err)
			}
		}
	} else {
		err = setActiveGame(filepath.Join(newest.Directory, newest.Name))
		if err != nil {
			return fmt.Errorf("error setting active game: %w", err)
		}
	}

	return nil
}

type Tracker struct {
	logger     *log.Logger
	mu         sync.Mutex
	activeCore string
	activeGame string
	events     []string
	coreTimes  map[string]int32
	gameTimes  map[string]int32
}

func (t *Tracker) loadCore() {
	t.mu.Lock()
	defer t.mu.Unlock()

	data, err := os.ReadFile(config.CoreNameFile)
	coreName := string(data)

	if err != nil {
		t.logger.Println("error reading core name:", err)
		// TODO: clear actives?
		return
	}

	if coreName != t.activeCore {
		// TODO: log events
		if coreName == "" || coreName == "MENU" {
			t.activeCore = ""
			t.activeGame = ""
			// TODO: set active
		} else {
			t.activeCore = coreName
			t.logger.Println("core changed:", t.activeCore)
		}
	}
}

func (t *Tracker) loadGame() {
	t.mu.Lock()
	defer t.mu.Unlock()

	activeGame, err := getActiveGame()
	if err != nil {
		t.logger.Println("error getting active game:", err)
		return
	}

	// TODO: always convert to an absolute path

	if activeGame != t.activeGame {
		t.activeGame = activeGame
		t.logger.Println("game changed:", t.activeGame)
		// TODO: log event
	}
}

func (t *Tracker) tick() {
	t.mu.Lock()
	defer t.mu.Unlock()

	if t.activeCore != "" {
		t.coreTimes[t.activeCore]++
	}

	if t.activeGame != "" {
		t.gameTimes[t.activeGame]++
	}
}

func startFileWatch(tracker *Tracker) (*fsnotify.Watcher, error) {
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
						tracker.loadCore()
					} else if event.Name == config.ActiveGameFile {
						tracker.loadGame()
					} else if strings.HasPrefix(event.Name, config.CoreConfigFolder) {
						err = loadRecent(event.Name)
						if err != nil {
							tracker.logger.Println("error loading recent file:", err)
						}
					}
				}
			case err, ok := <-watcher.Errors:
				if !ok {
					return
				}
				tracker.logger.Println("error on watcher:", err)
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

func startTicker(tracker *Tracker) {
	ticker := time.NewTicker(time.Second)
	go func() {
		count := 0
		for range ticker.C {
			tracker.tick()
			count++
		}
	}()
}

func main() {
	logger := log.New(os.Stdout, "", log.LstdFlags)

	tracker := &Tracker{
		logger:     logger,
		activeCore: "",
		activeGame: "",
		events:     []string{},
		coreTimes:  map[string]int32{},
		gameTimes:  map[string]int32{},
	}

	tracker.loadCore()
	if _, err := os.Stat(config.ActiveGameFile); err != nil {
		setActiveGame("")
	}

	watcher, err := startFileWatch(tracker)
	if err != nil {
		logger.Println("error starting file watch:", err)
		os.Exit(1)
	}
	defer watcher.Close()

	startTicker(tracker)

	<-make(chan struct{})
}
