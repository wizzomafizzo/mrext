package games

import (
	"encoding/json"
	"github.com/wizzomafizzo/mrext/cmd/remote/websocket"
	"github.com/wizzomafizzo/mrext/pkg/config"
	"github.com/wizzomafizzo/mrext/pkg/mister"
	"github.com/wizzomafizzo/mrext/pkg/service"
	"github.com/wizzomafizzo/mrext/pkg/tracker"
	"net/http"
)

type fakeDb struct {
	logger *service.Logger
}

func (f *fakeDb) FixPowerLoss() (bool, error) {
	return false, nil
}

func (f *fakeDb) AddEvent(ev tracker.EventAction) error {
	switch ev.Action {
	case tracker.EventActionCoreStart:
		websocket.Broadcast(f.logger, "coreRunning:"+ev.Target)
	case tracker.EventActionCoreStop:
		websocket.Broadcast(f.logger, "coreRunning:")
	case tracker.EventActionGameStart:
		websocket.Broadcast(f.logger, "gameRunning:"+ev.Target)
	case tracker.EventActionGameStop:
		websocket.Broadcast(f.logger, "gameRunning:")
	}

	return nil
}

func (f *fakeDb) UpdateCore(ct tracker.CoreTime) error {
	return nil
}

func (f *fakeDb) GetCore(name string) (tracker.CoreTime, error) {
	return tracker.CoreTime{}, nil
}

func (f *fakeDb) UpdateGame(gt tracker.GameTime) error {
	return nil
}

func (f *fakeDb) GetGame(id string) (tracker.GameTime, error) {
	return tracker.GameTime{}, nil
}

func (f *fakeDb) NoResults(err error) bool {
	return true
}

func StartTracker(logger *service.Logger, cfg *config.UserConfig) (*tracker.Tracker, func() error, error) {
	tr, err := tracker.NewTracker(logger, cfg, &fakeDb{
		logger: logger,
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
	Core string `json:"core"`
	Game string `json:"game"`
}

func HandlePlaying(tr *tracker.Tracker) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		playing := PlayingPayload{
			Core: tr.ActiveCore,
			Game: tr.ActiveGame,
		}

		err := json.NewEncoder(w).Encode(playing)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
}
