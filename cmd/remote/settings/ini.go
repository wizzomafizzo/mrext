package settings

import (
	"encoding/json"
	"github.com/wizzomafizzo/mrext/pkg/mister"
	"github.com/wizzomafizzo/mrext/pkg/service"
	"net/http"
)

func HandleSaveIni(logger *service.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		//iniFile, err := mister.LoadMisterIni()
		//if err != nil {
		//	http.Error(w, err.Error(), http.StatusInternalServerError)
		//	logger.Error("load mister.ini: %s", err)
		//	return
		//}
		//
		//err = mister.SaveMisterIni(iniFile)
		//if err != nil {
		//	http.Error(w, err.Error(), http.StatusInternalServerError)
		//	logger.Error("save mister.ini: %s", err)
		//	return
		//}
	}
}

type IniResponse struct {
	Active int              `json:"active"`
	Inis   []mister.IniFile `json:"inis"`
}

func HandleListInis(logger *service.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		inis, err := mister.ListMisterInis()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			logger.Error("list mister.inis: %s", err)
			return
		}

		activeIni, err := mister.GetCurrentIni()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			logger.Error("get current mister.ini: %s", err)
			return
		}

		iniResponse := IniResponse{
			Active: activeIni,
			Inis:   inis,
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

		availableInis, err := mister.ListMisterInis()
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

		err = mister.SetCurrentIni(args.Ini)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			logger.Error("set active mister.ini: %s", err)
			return
		}
	}
}