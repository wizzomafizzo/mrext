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

	name := "tapto"
	// TODO: is it possible to not rely on a specific path to the nfc app?
	binTemplate := config.ScriptsFolder + "/%s.sh"
	pidTemplate := "/tmp/%s.pid"

	if _, err := os.Stat(fmt.Sprintf(binTemplate, name)); err == nil {
		state.Available = true
	} else {
		name = "nfc"
		if _, err := os.Stat(fmt.Sprintf(binTemplate, name)); err == nil {
			state.Available = true
		} else {
			return state
		}
	}

	state.Name = name

	if _, err := os.Stat(fmt.Sprintf(pidTemplate, name)); err == nil {
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
			logger.Error("nfc status: encoding response: %s", err)
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
			logger.Error("nfc write: decoding request: %s", err)
			return
		}

		state := getNfcState()

		if !state.Available {
			http.Error(w, "nfc app not found", http.StatusInternalServerError)
			return
		}

		if !state.Running {
			http.Error(w, "nfc service not running", http.StatusInternalServerError)
			return
		}

		nfcBin := fmt.Sprintf(config.ScriptsFolder+"/%s.sh", state.Name)

		cmd := exec.Command(nfcBin, "-write", args.Path)
		cmd.Env = append(os.Environ(), config.UserAppPathEnv+"="+nfcBin)
		err = cmd.Run()

		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			logger.Error("nfc write: run command: %s", err)
			return
		}
	}
}

func NfcCancel(logger *service.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, _ *http.Request) {
		state := getNfcState()

		if !state.Available {
			http.Error(w, "nfc app not found", http.StatusInternalServerError)
			return
		}

		if !state.Running {
			return
		}

		nfcBin := fmt.Sprintf(config.ScriptsFolder+"/%s.sh", state.Name)

		cmd := exec.Command(nfcBin, "-service", "restart")
		cmd.Env = append(os.Environ(), config.UserAppPathEnv+"="+nfcBin)
		err := cmd.Run()

		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			logger.Error("nfc cancel: run command: %s", err)
			return
		}
	}
}
