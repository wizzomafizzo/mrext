package settings

import (
	"encoding/json"
	"github.com/wizzomafizzo/mrext/pkg/mister"
	"github.com/wizzomafizzo/mrext/pkg/service"
	"io"
	"net/http"
	"os"
	"strings"
)

const (
	hostnameKey   = "__hostname"
	macAddressKey = "__ethernetMacAddress"
)

type SaveIniRequest = map[string]string

func HandleSaveIni(logger *service.Logger, reqId int) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		logger.Info("save ini request: %d", reqId)

		var args SaveIniRequest

		err := json.NewDecoder(r.Body).Decode(&args)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			logger.Error("decode save ini request: %s", err)
			return
		}

		mi, err := mister.GetMisterIni(reqId)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			logger.Error("get mister.ini: %s", err)
			return
		}

		err = mi.Load()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			logger.Error("load mister.ini: %s", err)
			return
		}

		for key, value := range args {
			// custom internal setting
			if strings.HasPrefix(key, "__") {
				if key == hostnameKey {
					err = mister.UpdateHostname(value, false)
					if err != nil {
						logger.Error("set hostname: %s", err)
					}
				} else if key == macAddressKey {
					err = mister.UpdateConfiguredMacAddress(value)
					if err != nil {
						logger.Error("set mac address: %s", err)
					}
				}
			}

			err := mi.SetKey(key, value)
			if err != nil {
				logger.Error("update mister.ini: %s", err)
			}
			logger.Info("update mister.ini: %s=%s", key, value)
		}

		err = mi.Save()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			logger.Error("save mister.ini: %s", err)
			return
		}

		err = mister.RelaunchIfInMenu()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			logger.Error("relaunch mister: %s", err)
			return
		}
	}
}

func HandleLoadIni(logger *service.Logger, reqId int) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		logger.Info("load ini request: %d", reqId)

		mi, err := mister.GetMisterIni(reqId)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			logger.Error("get mister.ini: %s", err)
			return
		}

		err = mi.Load()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			logger.Error("load mister.ini: %s", err)
			return
		}

		section, err := mi.File.GetSection(mister.MainIniSection)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			logger.Error("get mister.ini section: %s", err)
			return
		}

		payload := make(map[string]string)
		for _, key := range section.Keys() {
			payload[key.Name()] = key.Value()
		}

		hostname, err := os.Hostname()
		if err != nil {
			hostname = ""
		}

		payload[hostnameKey] = hostname

		macAddress, err := mister.GetConfiguredMacAddress()
		if err != nil {
			macAddress = ""
		}

		payload[macAddressKey] = macAddress

		err = json.NewEncoder(w).Encode(payload)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			logger.Error("encode mister.ini: %s", err)
			return
		}
	}
}

type IniResponse struct {
	Active int                `json:"active"`
	Inis   []mister.MisterIni `json:"inis"`
}

func HandleListInis(logger *service.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		inis, err := mister.GetAllWithDefaultMisterIni()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			logger.Error("list mister.inis: %s", err)
			return
		}

		activeIni, err := mister.GetActiveIni()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			logger.Error("get current mister.ini: %s", err)
			return
		}

		inisList := make([]mister.MisterIni, 0)
		for _, ini := range inis {
			inisList = append(inisList, ini)
		}

		iniResponse := IniResponse{
			Active: activeIni,
			Inis:   inisList,
		}

		err = json.NewEncoder(w).Encode(iniResponse)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			logger.Error("failed to encode inis: %s", err)
			return
		}
	}
}

type SetActiveIniRequest struct {
	Ini int `json:"ini"`
}

func HandleSetActiveIni(logger *service.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var args SetActiveIniRequest

		err := json.NewDecoder(r.Body).Decode(&args)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			logger.Error("decode set active ini request: %s", err)
			return
		}

		if args.Ini < 1 || args.Ini > 4 {
			http.Error(w, "ini must be between 1 and 4", http.StatusInternalServerError)
			logger.Error("ini must be between 1 and 4")
			return
		}

		availableInis, err := mister.GetAllWithDefaultMisterIni()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			logger.Error("list mister.inis: %s", err)
			return
		}

		if args.Ini > len(availableInis) {
			http.Error(w, "ini does not exist", http.StatusInternalServerError)
			logger.Error("ini does not exist")
			return
		}

		err = mister.SetActiveIni(args.Ini)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			logger.Error("set active mister.ini: %s", err)
			return
		}
	}
}

type SetMenuBackgroundModeRequest struct {
	Mode int `json:"mode"`
}

func HandleSetMenuBackgroundMode(logger *service.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var args SetMenuBackgroundModeRequest

		err := json.NewDecoder(r.Body).Decode(&args)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			logger.Error("decode set menu background mode request: %s", err)
			return
		}

		err = mister.SetMenuBackgroundMode(args.Mode)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			logger.Error("set menu background mode: %s", err)
			return
		}

		err = mister.RelaunchIfInMenu()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			logger.Error("relaunch if in menu: %s", err)
			return
		}
	}
}

func HandleDownloadRemoteLog(logger *service.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		logger.Info("download remote log")

		// TODO: don't hardcode this path
		file, err := os.Open("/tmp/remote.log")
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			logger.Error("open remote log: %s", err)
			return
		}

		w.Header().Set("Content-Disposition", "attachment; filename=remote.log")
		w.Header().Set("Content-Type", "text/plain")

		_, err = io.Copy(w, file)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			logger.Error("download remote log: %s", err)
			return
		}
	}
}
