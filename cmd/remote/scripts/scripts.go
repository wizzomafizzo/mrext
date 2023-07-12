package scripts

import (
	"encoding/json"
	"github.com/gorilla/mux"
	"github.com/wizzomafizzo/mrext/pkg/config"
	"github.com/wizzomafizzo/mrext/pkg/input"
	"github.com/wizzomafizzo/mrext/pkg/mister"
	"github.com/wizzomafizzo/mrext/pkg/service"
	"net/http"
	"os/exec"
	"path/filepath"
)

func HandleLaunchScript(logger *service.Logger, kbd input.Keyboard) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		filename := vars["filename"]

		logger.Info("launch script request")

		path := filepath.Join(config.ScriptsFolder, filename)
		logger.Info("running script: %s", path)

		go func() {
			err := mister.RunScript(kbd, path)
			if err != nil {
				logger.Error("error running script: %s", err)
			}
		}()
	}
}

func HandleListScripts(logger *service.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		logger.Info("list scripts request")

		files, err := mister.GetAllScripts()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			logger.Error("error listing scripts: %s", err)
			return
		}

		var payload struct {
			CanLaunch bool            `json:"canLaunch"`
			Scripts   []mister.Script `json:"scripts"`
		}

		payload.CanLaunch = mister.ScriptCanLaunch()
		payload.Scripts = files

		err = json.NewEncoder(w).Encode(payload)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			logger.Error("error encoding response: %s", err)
			return
		}
	}
}

func HandleOpenScriptsConsole(logger *service.Logger, kbd input.Keyboard) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		logger.Info("open scripts console request")

		err := mister.OpenConsole(kbd)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			logger.Error("error opening console: %s", err)
			return
		}

		if mister.IsScriptRunning() {
			err = exec.Command("chvt", "2").Run()
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				logger.Error("error changing vt: %s", err)
				return
			}
		}
	}
}

func HandleKillActiveScript(logger *service.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		logger.Info("kill active script request")

		err := mister.KillActiveScript()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			logger.Error("error killing active script: %s", err)
			return
		}
	}
}
