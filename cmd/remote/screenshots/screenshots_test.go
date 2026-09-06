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

package screenshots

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
	"time"

	"github.com/bendahl/uinput"
	"github.com/wizzomafizzo/mrext/cmd/remote/control"
	"github.com/wizzomafizzo/mrext/pkg/input"
)

type keyEvent struct {
	key  int
	down bool
}

type screenshotKeyboard struct {
	uinput.Keyboard
	events []keyEvent
}

func (k *screenshotKeyboard) KeyDown(key int) error {
	k.events = append(k.events, keyEvent{key: key, down: true})
	return nil
}

func (k *screenshotKeyboard) KeyUp(key int) error {
	k.events = append(k.events, keyEvent{key: key})
	return nil
}

func TestTakeScreenshotMatchesControlShortcut(t *testing.T) {
	device := &screenshotKeyboard{}
	keyboard := input.Keyboard{Device: device}
	var delay time.Duration
	var logged bool
	handler := screenshotHandler(keyboard.Screenshot, func(string, ...any) { logged = true },
		func(duration time.Duration) { delay += duration })
	response := httptest.NewRecorder()
	request := httptest.NewRequestWithContext(t.Context(), http.MethodPost, "/screenshots", http.NoBody)
	handler.ServeHTTP(response, request)
	if response.Code != http.StatusOK || delay != time.Second || logged {
		t.Fatalf("status=%d delay=%v logged=%t", response.Code, delay, logged)
	}
	var payload ScreenshotPayload
	if err := json.Unmarshal(response.Body.Bytes(), &payload); err != nil {
		t.Fatal(err)
	}
	if payload != (ScreenshotPayload{}) {
		t.Fatalf("response shape changed: %+v", payload)
	}
	controlDevice := &screenshotKeyboard{}
	if err := control.SendKeyboard(input.Keyboard{Device: controlDevice}, "screenshot"); err != nil {
		t.Fatal(err)
	}
	want := []keyEvent{
		{key: uinput.KeyLeftalt, down: true},
		{key: uinput.KeyScrolllock, down: true},
		{key: uinput.KeyLeftalt},
		{key: uinput.KeyScrolllock},
	}
	if !reflect.DeepEqual(device.events, want) || !reflect.DeepEqual(device.events, controlDevice.events) {
		t.Fatalf("API events=%v Control events=%v", device.events, controlDevice.events)
	}
}

func TestTakeScreenshotReportsCaptureFailure(t *testing.T) {
	calls, logs, waits := 0, 0, 0
	handler := screenshotHandler(func() error {
		calls++
		return errors.New("keyboard unavailable")
	}, func(string, ...any) { logs++ }, func(time.Duration) { waits++ })
	response := httptest.NewRecorder()
	request := httptest.NewRequestWithContext(t.Context(), http.MethodPost, "/screenshots", http.NoBody)
	handler.ServeHTTP(response, request)
	if response.Code != http.StatusInternalServerError || calls != 1 || logs != 1 || waits != 0 {
		t.Fatalf("status=%d calls=%d logs=%d waits=%d", response.Code, calls, logs, waits)
	}
	if json.Valid(response.Body.Bytes()) {
		t.Fatal("failure returned a success payload")
	}
}
