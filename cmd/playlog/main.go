package main

import (
	"io"
	"log"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/fsnotify/fsnotify"

	"github.com/wizzomafizzo/mrext/pkg/config"
)

// TODO: confirm recent ini setting is on
// TODO: store game as hash
// TODO: read mgl launches from cores_recent
// TODO: handle failed mgl launch
// TODO: function to get active games folder
// TODO: function to read recent files properly

type Tracker struct {
	logger     *log.Logger
	mu         sync.Mutex
	activeCore string
	activeGame string
	events     []string
	coreTimes  map[string]int32
	gameTimes  map[string]int32
}

func (t *Tracker) loadCore() bool {
	t.mu.Lock()
	defer t.mu.Unlock()

	data, err := os.ReadFile(config.CoreNameFile)
	coreName := string(data)

	if err != nil {
		t.logger.Println("error reading core name:", err)
		return false
	}

	if coreName != t.activeCore {
		if coreName == "" || coreName == "MENU" {
			// clear loaded game
			t.activeGame = ""
		}
		t.activeCore = coreName
		t.logger.Println("core changed:", t.activeCore)
		return true
	} else {
		return false
	}
}

func (t *Tracker) loadGame(filename string) bool {
	t.mu.Lock()
	defer t.mu.Unlock()

	if strings.Contains(filename, "_recent") {
		if strings.HasSuffix(filename, "cores_recent.cfg") {
			return false
		}

		file, err := os.Open(filename)
		if err != nil {
			t.logger.Println("error opening game file:", err)
			return false
		}
		defer file.Close()

		folderBuf := make([]byte, 1024)
		n, err := file.Read(folderBuf)
		if err != nil && err != io.EOF {
			t.logger.Println("error reading game folder:", err)
			return false
		}
		gameFolder := string(folderBuf[:n])

		nameBuf := make([]byte, 256)
		n, err = file.Read(nameBuf)
		if err != nil && err != io.EOF {
			t.logger.Println("error reading game name:", err)
			return false
		}
		gameName := string(nameBuf[:n])

		gameFile := gameFolder + "/" + gameName

		if gameName != t.activeGame {
			t.activeGame = gameName
			t.logger.Println("game changed:", gameFile)
			return true
		}
	}

	return false
}

func (t *Tracker) tick() {
	t.mu.Lock()
	defer t.mu.Unlock()

	if t.activeCore != "" {
		t.coreTimes[t.activeCore]++
		t.logger.Println("core time:", t.activeCore, t.coreTimes[t.activeCore])
	}

	if t.activeGame != "" {
		t.gameTimes[t.activeGame]++
		t.logger.Println("game time:", t.activeGame, t.gameTimes[t.activeGame])
	}
}

func startFileWatch(logger *log.Logger, tracker *Tracker) (*fsnotify.Watcher, error) {
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
					} else if strings.HasPrefix(event.Name, config.CoreConfigFolder) {
						tracker.loadGame(event.Name)
					}
				}
			case err, ok := <-watcher.Errors:
				if !ok {
					return
				}
				logger.Println("error on watch event:", err)
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

	return watcher, nil
}

func startTicker(logger *log.Logger, tracker *Tracker) {
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

	watcher, err := startFileWatch(logger, tracker)
	if err != nil {
		logger.Println("error starting file watch:", err)
		os.Exit(1)
	}
	defer watcher.Close()

	startTicker(logger, tracker)

	<-make(chan struct{})
}
