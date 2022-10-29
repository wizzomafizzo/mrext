package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/wizzomafizzo/mrext/pkg/config"
	"github.com/wizzomafizzo/mrext/pkg/games"
	"github.com/wizzomafizzo/mrext/pkg/mister"
	"github.com/wizzomafizzo/mrext/pkg/service"
)

const (
	eventActionCoreStart = iota
	eventActionCoreStop
	eventActionGameStart
	eventActionGameStop
)

type eventAction struct {
	timestamp time.Time
	action    int
	target    string
	totalTime int // for recovery from power loss
}

type coreTime struct {
	name string
	time int
}

type gameTime struct {
	id     string
	path   string
	name   string
	folder string
	time   int
}

type tracker struct {
	logger     *service.Logger
	config     *config.UserConfig
	db         *playLogDb
	mu         sync.Mutex
	activeCore string
	activeGame string
	events     []eventAction
	coreTimes  map[string]coreTime
	gameTimes  map[string]gameTime
}

func newTracker(logger *service.Logger, cfg *config.UserConfig) (*tracker, error) {
	logger.Info("starting tracker")
	db, err := openPlayLogDb()
	if err != nil {
		return nil, err
	}

	fixed, err := db.fixPowerLoss()
	if err != nil {
		return nil, err
	} else if fixed {
		logger.Warn("fixed missing events from power loss")
	}

	return &tracker{
		logger:     logger,
		config:     cfg,
		db:         db,
		activeCore: "",
		activeGame: "",
		events:     []eventAction{},
		coreTimes:  map[string]coreTime{},
		gameTimes:  map[string]gameTime{},
	}, nil
}

func (tr *tracker) execHook(bin string, arg string) {
	if bin == "" {
		return
	}

	go func() {
		cmd := exec.Command(bin, arg)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		tr.logger.Info("executing hook: %s %s", bin, arg)
		err := cmd.Run()
		if err != nil {
			tr.logger.Error("error running hook: %s", err)
		}
	}()
}

func (tr *tracker) addEvent(action int, target string) {
	totalTime := 0

	if action == eventActionCoreStart || action == eventActionCoreStop {
		if ct, ok := tr.coreTimes[target]; ok {
			totalTime = ct.time
		}
	} else if action == eventActionGameStart || action == eventActionGameStop {
		if gt, ok := tr.gameTimes[target]; ok {
			totalTime = gt.time
		}
	}

	ev := eventAction{
		timestamp: time.Now(),
		action:    action,
		target:    target,
		totalTime: totalTime,
	}

	tr.events = append(tr.events, ev)
	tr.db.addEvent(ev)

	actionLabel := ""
	switch action {
	case eventActionCoreStart:
		tr.execHook(tr.config.PlayLog.OnCoreStart, target)
		actionLabel = "core started"
	case eventActionCoreStop:
		tr.execHook(tr.config.PlayLog.OnCoreStop, target)
		actionLabel = "core stopped"
	case eventActionGameStart:
		tr.execHook(tr.config.PlayLog.OnGameStart, target)
		actionLabel = "game started"
	case eventActionGameStop:
		tr.execHook(tr.config.PlayLog.OnGameStop, target)
		actionLabel = "game stopped"
	}

	tr.logger.Info("%s: %s (%ds)", actionLabel, target, totalTime)
}

func (tr *tracker) stopCore() bool {
	if tr.activeCore != "" {
		if ct, ok := tr.coreTimes[tr.activeCore]; ok && ct.time > 0 {
			tr.db.updateCore(ct)
		}

		tr.addEvent(eventActionCoreStop, tr.activeCore)
		tr.activeCore = ""

		return true
	} else {
		return false
	}
}

// Load the current running core and set it as active.
func (tr *tracker) loadCore() {
	tr.mu.Lock()
	defer tr.mu.Unlock()

	data, err := os.ReadFile(config.CoreNameFile)
	coreName := string(data)

	if err != nil {
		tr.logger.Error("error reading core name: %s", err)
		tr.stopCore()
		return
	}

	if coreName == config.MenuCore {
		mister.SetActiveGame("")
		coreName = ""
	}

	if coreName != tr.activeCore {
		tr.stopCore()

		tr.activeCore = coreName

		if coreName == "" {
			return
		}

		if _, ok := tr.coreTimes[coreName]; !ok {
			ct, err := tr.db.getCore(coreName)
			if noResults(err) {
				tr.coreTimes[coreName] = coreTime{
					name: coreName,
					time: 0,
				}
			} else if err != nil {
				tr.logger.Error("error loading core time: %s", err)
			} else {
				tr.coreTimes[coreName] = ct
			}
		}

		tr.addEvent(eventActionCoreStart, coreName)

		if !strings.HasPrefix(tr.activeGame, tr.activeCore) {
			tr.stopGame()
		}
	}
}

func (tr *tracker) stopGame() bool {
	if tr.activeGame != "" {
		if gt, ok := tr.gameTimes[tr.activeGame]; ok && gt.time > 0 {
			tr.db.updateGame(gt)
		}

		tr.addEvent(eventActionGameStop, tr.activeGame)
		tr.activeGame = ""
		return true
	} else {
		return false
	}
}

// Load the current running game and set it as active.
func (tr *tracker) loadGame() {
	tr.mu.Lock()
	defer tr.mu.Unlock()

	activeGame, err := mister.GetActiveGame()
	if err != nil {
		tr.logger.Error("error getting active game: %s", err)
		tr.stopGame()
		return
	} else if activeGame == "" {
		tr.stopGame()
		return
	}

	path := mister.ResolvePath(activeGame)
	filename := filepath.Base(path)
	name := strings.TrimSuffix(filename, filepath.Ext(filename))

	systems := games.FolderToSystems(path)
	var folder string
	if len(systems) > 0 && len(systems[0].Folder) > 0 {
		folder = systems[0].Folder[0]
	}

	id := fmt.Sprintf("%s/%s", folder, filename)

	if id != tr.activeGame {
		tr.stopGame()

		tr.activeGame = id

		if _, ok := tr.gameTimes[id]; !ok {
			gt, err := tr.db.getGame(id)
			if noResults(err) {
				tr.gameTimes[id] = gameTime{
					id:     id,
					path:   path,
					name:   name,
					folder: folder,
					time:   0,
				}
			} else if err != nil {
				tr.logger.Error("error loading game time: %s", err)
			} else {
				tr.gameTimes[id] = gt
			}
		}

		tr.addEvent(eventActionGameStart, id)
	}
}

func (tr *tracker) stopAll() {
	tr.mu.Lock()
	defer tr.mu.Unlock()
	tr.stopCore()
	tr.stopGame()
}

// Increment time of active core and game.
func (tr *tracker) tick(saveInterval int) {
	tr.mu.Lock()
	defer tr.mu.Unlock()

	saveSeconds := saveInterval * 60

	if tr.activeCore != "" {
		if ct, ok := tr.coreTimes[tr.activeCore]; ok {
			ct.time++

			if saveInterval > 0 && ct.time%saveSeconds == 0 {
				tr.logger.Info("saving core time: %s (%ds)", ct.name, ct.time)
				err := tr.db.updateCore(ct)
				if err != nil {
					tr.logger.Error("error updating core time: %s", err)
				}
			}

			tr.coreTimes[tr.activeCore] = ct
		}
	}

	if tr.activeGame != "" {
		if gt, ok := tr.gameTimes[tr.activeGame]; ok {
			gt.time++

			if saveInterval > 0 && gt.time%saveSeconds == 0 {
				tr.logger.Info("saving game time: %s (%ds)", gt.id, gt.time)
				err := tr.db.updateGame(gt)
				if err != nil {
					tr.logger.Error("error updating game time: %s", err)
				}
			}

			tr.gameTimes[tr.activeGame] = gt
		}
	}
}

// Start thread for updating core/game play times.
func (tr *tracker) startTicker(saveInterval int) {
	tr.logger.Info("starting ticker with save interval %dm", saveInterval)
	ticker := time.NewTicker(time.Second)
	go func() {
		count := 0
		for range ticker.C {
			tr.tick(saveInterval)
			count++
		}
	}()
}
