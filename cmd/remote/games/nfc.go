package games

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/exec"

	"github.com/wizzomafizzo/mrext/pkg/config"
	"github.com/wizzomafizzo/mrext/pkg/service"
)

type NfcState struct {
	Available bool   `json:"installed"`
	Running   bool   `json:"running"`
	Name      string `json:"name"`
}

func getNfcState() NfcState {
	state := NfcState{}

	bin := config.ScriptsFolder + "/tapto.sh"
	pid := "/tmp/tapto/tapto.pid"

	if _, err := os.Stat(bin); err == nil {
		state.Available = true
	} else {
		return state
	}

	state.Name = "tapto"

	if _, err := os.Stat(pid); err == nil {
		state.Running = true
	}

	return state
}

func NfcStatus(logger *service.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, _ *http.Request) {
		payload := getNfcState()

		err := json.NewEncoder(w).Encode(payload)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			logger.Error("tapto status: encoding response: %s", err)
			return
		}
	}
}

func NfcWrite(logger *service.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var args struct {
			Path string `json:"path"`
		}

		err := json.NewDecoder(r.Body).Decode(&args)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			logger.Error("tapto write: decoding request: %s", err)
			return
		}

		state := getNfcState()

		if !state.Available {
			http.Error(w, "tapto app not found", http.StatusInternalServerError)
			return
		}

		if !state.Running {
			http.Error(w, "tapto service not running", http.StatusInternalServerError)
			return
		}

		nfcBin := fmt.Sprintf(config.ScriptsFolder+"/%s.sh", state.Name)

		cmd := exec.Command(nfcBin, "-write", args.Path)
		err = cmd.Run()

		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			logger.Error("tapto write: run command: %s", err)
			return
		}
	}
}

func NfcCancel(logger *service.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, _ *http.Request) {
		state := getNfcState()

		if !state.Available {
			http.Error(w, "tapto app not found", http.StatusInternalServerError)
			return
		}

		if !state.Running {
			return
		}

		nfcBin := fmt.Sprintf(config.ScriptsFolder+"/%s.sh", state.Name)

		cmd := exec.Command(nfcBin, "-service", "restart")
		err := cmd.Run()

		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			logger.Error("tapto cancel: run command: %s", err)
			return
		}
	}
}
