package games

import (
	"encoding/json"
	"github.com/wizzomafizzo/mrext/pkg/config"
	"github.com/wizzomafizzo/mrext/pkg/service"
	"net/http"
	"os"
	"os/exec"
)

// TODO: is it possible to not rely on a specific path to the nfc app?
const nfcBin = config.ScriptsFolder + "/nfc.sh"

func nfcExists() bool {
	_, err := os.Stat(nfcBin)
	return !os.IsNotExist(err)
}

func NfcStatus(logger *service.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, _ *http.Request) {
		payload := struct {
			Available bool `json:"available"`
			Running   bool `json:"running"`
		}{}

		if nfcExists() {
			payload.Available = true
		}

		_, err := os.Stat("/tmp/nfc.pid")
		if err == nil {
			payload.Running = true
		}

		err = json.NewEncoder(w).Encode(payload)
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

		if !nfcExists() {
			http.Error(w, "nfc script not found", http.StatusInternalServerError)
			return
		}

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
		if !nfcExists() {
			http.Error(w, "nfc script not found", http.StatusInternalServerError)
			return
		}

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
