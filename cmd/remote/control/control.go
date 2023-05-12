package control

import (
	"fmt"
	"github.com/bendahl/uinput"
	"github.com/gorilla/mux"
	"github.com/wizzomafizzo/mrext/pkg/input"
	"github.com/wizzomafizzo/mrext/pkg/service"
	"net/http"
	"strconv"
)

func SendRawKeyboard(kbd input.Keyboard, code int) error {
	if code < 0 {
		kbd.Combo(uinput.KeyLeftshift, -code)
	} else {
		kbd.Press(code)
	}

	return nil
}

func HandleRawKeyboard(kbd input.Keyboard, logger *service.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		keyQ := vars["key"]

		key, err := strconv.Atoi(keyQ)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			logger.Error("raw keyboard input (%s) is invalid: %s", keyQ, err)
			return
		}

		_ = SendRawKeyboard(kbd, key)
	}
}

func SendKeyboard(kbd input.Keyboard, key string) error {
	switch key {
	case "up":
		kbd.Up()
	case "down":
		kbd.Down()
	case "left":
		kbd.Left()
	case "right":
		kbd.Right()
	case "volume_up":
		kbd.VolumeUp()
	case "volume_down":
		kbd.VolumeDown()
	case "volume_mute":
		kbd.VolumeMute()
	case "menu":
		kbd.Menu()
	case "back":
		kbd.Back()
	case "confirm":
		kbd.Confirm()
	case "cancel":
		kbd.Cancel()
	case "osd":
		kbd.Osd()
	case "screenshot":
		kbd.Screenshot()
	case "raw_screenshot":
		kbd.RawScreenshot()
	case "pair_bluetooth":
		kbd.PairBluetooth()
	case "change_background":
		kbd.ChangeBackground()
	case "core_select":
		kbd.CoreSelect()
	case "user":
		kbd.User()
	case "reset":
		kbd.Reset()
	case "toggle_core_dates":
		kbd.ToggleCoreDates()
	case "console":
		kbd.Console()
	case "computer_osd":
		kbd.ComputerOsd()
	default:
		return fmt.Errorf("unknown key: %s", key)
	}

	return nil
}

func HandleKeyboard(kbd input.Keyboard) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		key := vars["key"]

		err := SendKeyboard(kbd, key)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
}
