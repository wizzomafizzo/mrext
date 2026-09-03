// mrext
// Copyright (c) 2026 mrext contributors.
// SPDX-License-Identifier: GPL-3.0-or-later
//
// This file is part of mrext.
//
// mrext is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// mrext is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with mrext. If not, see <http://www.gnu.org/licenses/>.

package control

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/bendahl/uinput"
	"github.com/gorilla/mux"
	"github.com/wizzomafizzo/mrext/pkg/input"
	"github.com/wizzomafizzo/mrext/pkg/service"
)

func wrapKeyboard(action string, err error) error {
	if err != nil {
		return fmt.Errorf("%s: %w", action, err)
	}
	return nil
}

func SendRawKeyboard(kbd input.Keyboard, code int) error {
	if code < 0 {
		return wrapKeyboard("send shifted raw key", kbd.Combo(uinput.KeyLeftshift, -code))
	}
	return wrapKeyboard("send raw key", kbd.Press(code))
}

func SendRawKeyboardDown(kbd input.Keyboard, code int) error {
	return wrapKeyboard("press raw key", kbd.KeyDown(code))
}

func SendRawKeyboardUp(kbd input.Keyboard, code int) error {
	return wrapKeyboard("release raw key", kbd.KeyUp(code))
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

		if err := SendRawKeyboard(kbd, key); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			logger.Error("failed to send raw keyboard input: %s", err)
		}
	}
}

func SendKeyboard(kbd input.Keyboard, key string) error {
	switch key {
	case "up":
		return wrapKeyboard("send up key", kbd.Up())
	case "down":
		return wrapKeyboard("send down key", kbd.Down())
	case "left":
		return wrapKeyboard("send left key", kbd.Left())
	case "right":
		return wrapKeyboard("send right key", kbd.Right())
	case "volume_up":
		return wrapKeyboard("raise volume", kbd.VolumeUp())
	case "volume_down":
		return wrapKeyboard("lower volume", kbd.VolumeDown())
	case "volume_mute":
		return wrapKeyboard("toggle mute", kbd.VolumeMute())
	case "menu":
		return wrapKeyboard("send menu key", kbd.Menu())
	case "back":
		return wrapKeyboard("send back key", kbd.Back())
	case "confirm":
		return wrapKeyboard("send confirm key", kbd.Confirm())
	case "cancel":
		return wrapKeyboard("send cancel key", kbd.Cancel())
	case "osd":
		return wrapKeyboard("send OSD key", kbd.OSD())
	case "screenshot":
		return wrapKeyboard("capture screenshot", kbd.Screenshot())
	case "raw_screenshot":
		return wrapKeyboard("capture raw screenshot", kbd.RawScreenshot())
	case "pair_bluetooth":
		return wrapKeyboard("pair Bluetooth", kbd.PairBluetooth())
	case "change_background":
		return wrapKeyboard("change background", kbd.ChangeBackground())
	case "core_select":
		return wrapKeyboard("select core", kbd.CoreSelect())
	case "user":
		return wrapKeyboard("send user key", kbd.User())
	case "reset":
		return wrapKeyboard("reset core", kbd.Reset())
	case "toggle_core_dates":
		return wrapKeyboard("toggle core dates", kbd.ToggleCoreDates())
	case "console":
		return wrapKeyboard("open console", kbd.Console())
	case "exit_console":
		return wrapKeyboard("exit console", kbd.ExitConsole())
	case "computer_osd":
		return wrapKeyboard("send computer OSD key", kbd.ComputerOSD())
	default:
		return fmt.Errorf("unknown key: %s", key)
	}
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
