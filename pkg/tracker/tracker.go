package tracker

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
	EventActionCoreStart = iota
	EventActionCoreStop
	EventActionGameStart
	EventActionGameStop
)

type EventAction struct {
	Timestamp  time.Time
	Action     int
	Target     string
	TargetPath string
	TotalTime  int // for recovery from power loss
}

type CoreTime struct {
	Name string
	Time int
}

type GameTime struct {
	Id     string
	Path   string
	Name   string
	Folder string
	Time   int
}

type Db interface {
	FixPowerLoss() (bool, error)
	AddEvent(ev EventAction) error
	UpdateCore(ct CoreTime) error
	GetCore(name string) (CoreTime, error)
	UpdateGame(gt GameTime) error
	GetGame(id string) (GameTime, error)
	NoResults(err error) bool
}

type Tracker struct {
	Logger     *service.Logger
	Config     *config.UserConfig
	Db         Db
	mu         sync.Mutex
	ActiveCore string
	ActiveGame string
	Events     []EventAction
	CoreTimes  map[string]CoreTime
	GameTimes  map[string]GameTime
}

func NewTracker(logger *service.Logger, cfg *config.UserConfig, db Db) (*Tracker, error) {
	logger.Info("starting tracker")

	fixed, err := db.FixPowerLoss()
	if err != nil {
		return nil, err
	} else if fixed {
		logger.Warn("fixed missing events from power loss")
	}

	return &Tracker{
		Logger:     logger,
		Config:     cfg,
		Db:         db,
		ActiveCore: "",
		ActiveGame: "",
		Events:     []EventAction{},
		CoreTimes:  map[string]CoreTime{},
		GameTimes:  map[string]GameTime{},
	}, nil
}

func (tr *Tracker) execHook(bin string, arg string) {
	if bin == "" {
		return
	}

	go func() {
		cmd := exec.Command(bin, arg)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		tr.Logger.Info("executing hook: %s %s", bin, arg)
		err := cmd.Run()
		if err != nil {
			tr.Logger.Error("error running hook: %s", err)
		}
	}()
}

func (tr *Tracker) addEvent(action int, target string) {
	totalTime := 0

	if action == EventActionCoreStart || action == EventActionCoreStop {
		if ct, ok := tr.CoreTimes[target]; ok {
			totalTime = ct.Time
		}
	} else if action == EventActionGameStart || action == EventActionGameStop {
		if gt, ok := tr.GameTimes[target]; ok {
			totalTime = gt.Time
		}
	}

	ev := EventAction{
		Timestamp: time.Now(),
		Action:    action,
		Target:    target,
		TotalTime: totalTime,
	}

	targetTime, ok := tr.GameTimes[target]
	if ok {
		ev.TargetPath = targetTime.Path
	}

	tr.Events = append(tr.Events, ev)
	err := tr.Db.AddEvent(ev)
	if err != nil {
		tr.Logger.Error("error saving event: %s", err)
	}

	actionLabel := ""
	switch action {
	case EventActionCoreStart:
		tr.execHook(tr.Config.PlayLog.OnCoreStart, target)
		actionLabel = "core started"
	case EventActionCoreStop:
		tr.execHook(tr.Config.PlayLog.OnCoreStop, target)
		actionLabel = "core stopped"
	case EventActionGameStart:
		tr.execHook(tr.Config.PlayLog.OnGameStart, target)
		actionLabel = "game started"
	case EventActionGameStop:
		tr.execHook(tr.Config.PlayLog.OnGameStop, target)
		actionLabel = "game stopped"
	}

	tr.Logger.Info("%s: %s (%ds)", actionLabel, target, totalTime)
}

func (tr *Tracker) stopCore() bool {
	if tr.ActiveCore != "" {
		if ct, ok := tr.CoreTimes[tr.ActiveCore]; ok && ct.Time > 0 {
			err := tr.Db.UpdateCore(ct)
			if err != nil {
				tr.Logger.Error("error saving core time: %s", err)
			}
		}

		tr.addEvent(EventActionCoreStop, tr.ActiveCore)
		tr.ActiveCore = ""

		return true
	} else {
		return false
	}
}

// Load the current running core and set it as active.
func (tr *Tracker) LoadCore() {
	tr.mu.Lock()
	defer tr.mu.Unlock()

	data, err := os.ReadFile(config.CoreNameFile)
	coreName := string(data)

	if err != nil {
		tr.Logger.Error("error reading core name: %s", err)
		tr.stopCore()
		return
	}

	if coreName == config.MenuCore {
		err := mister.SetActiveGame("")
		if err != nil {
			tr.Logger.Error("error setting active game: %s", err)
		}
		coreName = ""
	}

	if coreName != tr.ActiveCore {
		tr.stopCore()

		tr.ActiveCore = coreName

		if coreName == "" {
			return
		}

		if _, ok := tr.CoreTimes[coreName]; !ok {
			ct, err := tr.Db.GetCore(coreName)
			if tr.Db.NoResults(err) {
				tr.CoreTimes[coreName] = CoreTime{
					Name: coreName,
					Time: 0,
				}
			} else if err != nil {
				tr.Logger.Error("error loading core time: %s", err)
			} else {
				tr.CoreTimes[coreName] = ct
			}
		}

		tr.addEvent(EventActionCoreStart, coreName)

		if !strings.HasPrefix(tr.ActiveGame, tr.ActiveCore) {
			tr.stopGame()
		}
	}
}

func (tr *Tracker) stopGame() bool {
	if tr.ActiveGame != "" {
		if gt, ok := tr.GameTimes[tr.ActiveGame]; ok && gt.Time > 0 {
			err := tr.Db.UpdateGame(gt)
			if err != nil {
				tr.Logger.Error("error saving game time: %s", err)
			}
		}

		tr.addEvent(EventActionGameStop, tr.ActiveGame)
		tr.ActiveGame = ""
		return true
	} else {
		return false
	}
}

// Load the current running game and set it as active.
func (tr *Tracker) loadGame() {
	tr.mu.Lock()
	defer tr.mu.Unlock()

	activeGame, err := mister.GetActiveGame()
	if err != nil {
		tr.Logger.Error("error getting active game: %s", err)
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

	if id != tr.ActiveGame {
		tr.stopGame()

		tr.ActiveGame = id

		if _, ok := tr.GameTimes[id]; !ok {
			gt, err := tr.Db.GetGame(id)
			if tr.Db.NoResults(err) {
				tr.GameTimes[id] = GameTime{
					Id:     id,
					Path:   path,
					Name:   name,
					Folder: folder,
					Time:   0,
				}
			} else if err != nil {
				tr.Logger.Error("error loading game time: %s", err)
			} else {
				tr.GameTimes[id] = gt
			}
		}

		tr.addEvent(EventActionGameStart, id)
	}
}

func (tr *Tracker) StopAll() {
	tr.mu.Lock()
	defer tr.mu.Unlock()
	tr.stopCore()
	tr.stopGame()
}

// Increment time of active core and game.
func (tr *Tracker) tick(saveInterval int) {
	tr.mu.Lock()
	defer tr.mu.Unlock()

	saveSeconds := saveInterval * 60

	if tr.ActiveCore != "" {
		if ct, ok := tr.CoreTimes[tr.ActiveCore]; ok {
			ct.Time++

			if saveInterval > 0 && ct.Time%saveSeconds == 0 {
				tr.Logger.Info("saving core time: %s (%ds)", ct.Name, ct.Time)
				err := tr.Db.UpdateCore(ct)
				if err != nil {
					tr.Logger.Error("error updating core time: %s", err)
				}
			}

			tr.CoreTimes[tr.ActiveCore] = ct
		}
	}

	if tr.ActiveGame != "" {
		if gt, ok := tr.GameTimes[tr.ActiveGame]; ok {
			gt.Time++

			if saveInterval > 0 && gt.Time%saveSeconds == 0 {
				tr.Logger.Info("saving game time: %s (%ds)", gt.Id, gt.Time)
				err := tr.Db.UpdateGame(gt)
				if err != nil {
					tr.Logger.Error("error updating game time: %s", err)
				}
			}

			tr.GameTimes[tr.ActiveGame] = gt
		}
	}
}

// StartTicker Start thread for updating core/game play times.
func (tr *Tracker) StartTicker(saveInterval int) {
	tr.Logger.Info("starting ticker with save interval %dm", saveInterval)
	ticker := time.NewTicker(time.Second)
	go func() {
		count := 0
		for range ticker.C {
			tr.tick(saveInterval)
			count++
		}
	}()
}
