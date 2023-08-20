package systems

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/wizzomafizzo/mrext/cmd/remote/menu"
	"github.com/wizzomafizzo/mrext/pkg/service"

	"github.com/gorilla/mux"

	"github.com/wizzomafizzo/mrext/pkg/games"
	"github.com/wizzomafizzo/mrext/pkg/mister"
	"github.com/wizzomafizzo/mrext/pkg/utils"
)

type System struct {
	Id            string `json:"id"`
	Name          string `json:"name"`
	Category      string `json:"category"`
	LocalisedName string `json:"localisedName"`
}

var ignoreSystems = []string{
	"Arcade",
	"NESMusic",
	"SNESMusic",
}

func ListSystems(logger *service.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var systems []System

		existingSystems := utils.MapKeys(games.SystemsWithRbf())

		for _, system := range games.Systems {
			if utils.Contains(ignoreSystems, system.Id) {
				continue
			}

			if !utils.Contains(existingSystems, system.Id) {
				continue
			}

			localisedName, err := menu.GetNamesTxt(system.Name, "")

			if err != nil {
				localisedName = ""
			}

			systems = append(systems, System{
				Id:   system.Id,
				Name: system.Name,
				// TODO: error checking
				Category:      strings.Split(system.Rbf, "/")[0][1:],
				LocalisedName: localisedName,
			})
		}

		err := json.NewEncoder(w).Encode(systems)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			logger.Error("list systems: during encode: %s", err)
			return
		}
	}
}

func LaunchCore(logger *service.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)

		id := vars["id"]
		system, err := games.GetSystem(id)
		if err != nil {
			http.NotFound(w, r)
			return
		}

		err = mister.LaunchCore(*system)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			logger.Error("launch core: during launch: %s", err)
			return
		}
	}
}
