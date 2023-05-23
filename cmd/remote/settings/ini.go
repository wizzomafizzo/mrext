package settings

import (
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
