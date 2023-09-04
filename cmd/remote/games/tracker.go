package games

import (
	"bytes"
	"encoding/json"
	"github.com/wizzomafizzo/mrext/cmd/remote/websocket"
	"github.com/wizzomafizzo/mrext/pkg/config"
	"github.com/wizzomafizzo/mrext/pkg/mister"
	"github.com/wizzomafizzo/mrext/pkg/service"
	"github.com/wizzomafizzo/mrext/pkg/tracker"
	"net/http"
	"os"
	"path/filepath"
)

type fakeDb struct {
	logger *service.Logger
	cfg    *config.UserConfig
}

func (f *fakeDb) FixPowerLoss() (bool, error) {
	return false, nil
}

func (f *fakeDb) AddEvent(ev tracker.EventAction) error {
	switch ev.Action {
	case tracker.EventActionCoreStart:
		websocket.Broadcast(f.logger, "coreRunning:"+ev.Target)
		SendAnnounceGame(f.cfg, f.logger, &ev)
	case tracker.EventActionCoreStop:
		websocket.Broadcast(f.logger, "coreRunning:")
		SendAnnounceGame(f.cfg, f.logger, &ev)
	case tracker.EventActionGameStart:
		websocket.Broadcast(f.logger, "gameRunning:"+ev.Target)
		SendAnnounceGame(f.cfg, f.logger, &ev)
	case tracker.EventActionGameStop:
		websocket.Broadcast(f.logger, "gameRunning:")
		SendAnnounceGame(f.cfg, f.logger, &ev)
	}

	return nil
}

func (f *fakeDb) UpdateCore(_ tracker.CoreTime) error {
	return nil
}

func (f *fakeDb) GetCore(_ string) (tracker.CoreTime, error) {
	return tracker.CoreTime{}, nil
}

func (f *fakeDb) UpdateGame(_ tracker.GameTime) error {
	return nil
}

func (f *fakeDb) GetGame(_ string) (tracker.GameTime, error) {
	return tracker.GameTime{}, nil
}

func (f *fakeDb) NoResults(_ error) bool {
	return true
}

func StartTracker(logger *service.Logger, cfg *config.UserConfig) (*tracker.Tracker, func() error, error) {
	tr, err := tracker.NewTracker(logger, cfg, &fakeDb{
		logger: logger,
		cfg:    cfg,
	})
	if err != nil {
		logger.Error("failed to start tracker: %s", err)
		return nil, nil, err
	}

	tr.LoadCore()
	if !mister.ActiveGameEnabled() {
		err := mister.SetActiveGame("")
		if err != nil {
			tr.Logger.Error("error setting active game: %s", err)
		}
	}

	watcher, err := tracker.StartFileWatch(tr)
	if err != nil {
		tr.Logger.Error("error starting file watch: %s", err)
		return nil, nil, err
	}

	tr.StartTicker(0)

	return tr, func() error {
		err := watcher.Close()
		if err != nil {
			tr.Logger.Error("error closing file watcher: %s", err)
		}
		tr.StopAll()
		return nil
	}, nil
}

type PlayingPayload struct {
	Core       string `json:"core"`
	System     string `json:"system"`
	SystemName string `json:"systemName"`
	Game       string `json:"game"`
	GameName   string `json:"gameName"`
}

func HandlePlaying(tr *tracker.Tracker) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		playing := PlayingPayload{
			Core:       tr.ActiveCore,
			System:     tr.ActiveSystem,
			SystemName: tr.ActiveSystemName,
			Game:       tr.ActiveGame,
			GameName:   tr.ActiveGameName,
		}

		err := json.NewEncoder(w).Encode(playing)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
}

type AnnounceGamePayload struct {
	Platform     string `json:"platform"`
	Hostname     string `json:"hostname"`
	Core         string `json:"core"`
	System       string `json:"system"`
	SystemName   string `json:"systemName"`
	GamePath     string `json:"gamePath"`
	GameFilename string `json:"gameFilename"`
	GameName     string `json:"gameName"`
}

func SendAnnounceGame(cfg *config.UserConfig, logger *service.Logger, ev *tracker.EventAction) {
	hostname, err := os.Hostname()
	if err != nil {
		hostname = ""
	}

	announce := AnnounceGamePayload{
		Platform:     "MiSTer",
		Hostname:     hostname,
		Core:         ev.ActiveCore.Core,
		System:       ev.ActiveCore.System,
		SystemName:   ev.ActiveCore.SystemName,
		GamePath:     ev.ActiveGame.Path,
		GameFilename: filepath.Base(ev.ActiveGame.Path),
		GameName:     ev.ActiveGame.Name,
	}

	url := cfg.Remote.AnnounceGameUrl
	data, err := json.Marshal(announce)
	if err != nil {
		logger.Error("error marshalling announce payload: %s", err)
		return
	}

	if url != "" {
		resp, err := http.Post(url, "application/json", bytes.NewBuffer(data))
		if err != nil {
			logger.Error("error sending announce payload: %s", err)
			return
		}
		defer resp.Body.Close()
	}
}
