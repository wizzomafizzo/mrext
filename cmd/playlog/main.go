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
	"github.com/wizzomafizzo/mrext/pkg/games"
	"github.com/wizzomafizzo/mrext/pkg/mister"
)

// TODO: confirm recent ini setting is on
// TODO: store game as hash
// TODO: handle failed mgl launch
// TODO: ticker interval and save interval should be configurable

// Read a core's recent file and attempt to write the newest entry's
// launchable path to ACTIVEGAME.
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

type Event struct {
	timestamp time.Time
	action    string
	target    string
	totalTime int32 // for recovery from power loss
}

type CoreTime struct {
	name string
	time int32
}

type GameTime struct {
	id     string
	path   string
	name   string
	folder string
	time   int32
}

type Tracker struct {
	logger     *log.Logger
	mu         sync.Mutex
	activeCore string
	activeGame string
	events     []Event
	coreTimes  map[string]CoreTime
	gameTimes  map[string]GameTime
}

// Load the current running core and set it as active.
func (t *Tracker) loadCore() {
	t.mu.Lock()
	defer t.mu.Unlock()

	data, err := os.ReadFile(config.CoreNameFile)
	coreName := string(data)

	if err != nil {
		t.activeCore = ""
		mister.SetActiveGame("")
		t.logger.Println("error reading core name:", err)
		return
	}

	if coreName == "MENU" {
		coreName = ""
	}

	if coreName != t.activeCore {
		// TODO: log events
		if coreName == "" {
			t.logger.Println("core exited:", t.activeCore)
			t.activeCore = ""
			mister.SetActiveGame("")
		} else {
			t.activeCore = coreName
			if _, ok := t.coreTimes[coreName]; !ok {
				t.coreTimes[coreName] = CoreTime{
					name: coreName,
					time: 0,
				}
			}
			t.logger.Println("core changed:", t.coreTimes[coreName])
		}
	}
}

// Load the current running game and set it as active.
func (t *Tracker) loadGame() {
	t.mu.Lock()
	defer t.mu.Unlock()

	exitGame := func() {
		if t.activeGame != "" {
			// TODO: log event
			t.logger.Println("game exited:", t.activeGame)
			t.activeGame = ""
		}
	}

	activeGame, err := mister.GetActiveGame()
	if err != nil {
		exitGame()
		t.logger.Println("error getting active game:", err)
		return
	} else if activeGame == "" {
		exitGame()
		return
	}

	path := mister.ResolvePath(activeGame)
	filename := filepath.Base(path)
	name := strings.TrimSuffix(filename, filepath.Ext(filename))

	systems := games.FolderToSystems(path)
	var folder string
	if len(systems) == 0 {
		folder = "UNKNOWN"
	} else {
		folder = systems[0].Folder
	}

	id := fmt.Sprintf("%s/%s", folder, filename)

	if id != t.activeGame {
		exitGame()
		t.activeGame = id
		if _, ok := t.gameTimes[id]; !ok {
			t.gameTimes[id] = GameTime{
				id:     id,
				path:   path,
				name:   name,
				folder: folder,
				time:   0,
			}
		}
		t.logger.Println("game started:", t.gameTimes[id])
		// TODO: log event
	}
}

// Increment time of active core and game.
func (t *Tracker) tick() {
	t.mu.Lock()
	defer t.mu.Unlock()

	if t.activeCore != "" {
		if coreTime, ok := t.coreTimes[t.activeCore]; ok {
			coreTime.time++
			t.coreTimes[t.activeCore] = coreTime
		}
	}

	if t.activeGame != "" {
		if gameTime, ok := t.gameTimes[t.activeGame]; ok {
			gameTime.time++
			t.gameTimes[t.activeGame] = gameTime
		}
	}
}

// Start thread for monitoring changes to all files relating to core/game launches.
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

// Start thread for updating core/game play times.
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
		events:     []Event{},
		coreTimes:  map[string]CoreTime{},
		gameTimes:  map[string]GameTime{},
	}

	tracker.loadCore()
	if !mister.ActiveGameEnabled() {
		mister.SetActiveGame("")
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