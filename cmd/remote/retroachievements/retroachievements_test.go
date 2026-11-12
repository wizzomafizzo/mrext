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

//nolint:gosec // Tests only operate on temporary fixture paths.
package retroachievements

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/wizzomafizzo/mrext/pkg/service"
)

const loginFixture = "username=someone\npassword=secret\nhardcore=0\n"

func testModes(t *testing.T, content string) *Modes {
	t.Helper()
	root := t.TempDir()
	path := filepath.Join(root, "retroachievements.cfg")
	if content != "" {
		if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	modes := New(service.NewLogger("retroachievements-test"))
	modes.path = path
	modes.coreName = filepath.Join(root, "CORENAME")
	modes.pollInterval = 2 * time.Millisecond
	modes.settleDelay = 5 * time.Millisecond
	modes.startTimeout = time.Second
	if err := os.WriteFile(modes.coreName, []byte("MENU"), 0o644); err != nil {
		t.Fatal(err)
	}
	return modes
}

func serve(t *testing.T, handler http.HandlerFunc, method, body string) *httptest.ResponseRecorder {
	t.Helper()
	recorder := httptest.NewRecorder()
	handler(recorder, httptest.NewRequestWithContext(
		t.Context(), method, "/api/settings/retroachievements", strings.NewReader(body),
	))
	if strings.Contains(recorder.Body.String(), "secret") || strings.Contains(recorder.Body.String(), "someone") {
		t.Fatalf("response leaks the login: %q", recorder.Body.String())
	}
	return recorder
}

func decodeStatus(t *testing.T, recorder *httptest.ResponseRecorder) Status {
	t.Helper()
	var status Status
	if err := json.NewDecoder(recorder.Body).Decode(&status); err != nil {
		t.Fatal(err)
	}
	return status
}

func TestStatusWithoutTheFile(t *testing.T) {
	modes := testModes(t, "")
	recorder := serve(t, modes.HandleStatus(), http.MethodGet, "")
	if recorder.Code != http.StatusOK {
		t.Fatalf("code = %d", recorder.Code)
	}
	if status := decodeStatus(t, recorder); status != (Status{}) {
		t.Fatalf("status = %+v, want all false", status)
	}
}

func TestStatusReadsOnlyTheHardcoreKeys(t *testing.T) {
	modes := testModes(t, loginFixture+"force_hardcore=1\n")
	recorder := serve(t, modes.HandleStatus(), http.MethodGet, "")
	want := Status{Present: true, Hardcore: false, Forced: true}
	if status := decodeStatus(t, recorder); status != want {
		t.Fatalf("status = %+v, want %+v", status, want)
	}
}

func TestSetHardcore(t *testing.T) {
	modes := testModes(t, loginFixture)
	recorder := serve(t, modes.HandleSetHardcore(), http.MethodPut, `{"hardcore":true}`)
	if recorder.Code != http.StatusOK {
		t.Fatalf("code = %d: %s", recorder.Code, recorder.Body.String())
	}
	if status := decodeStatus(t, recorder); !status.Present || !status.Hardcore {
		t.Fatalf("status = %+v, want hardcore", status)
	}
	data, err := os.ReadFile(modes.path)
	if err != nil {
		t.Fatal(err)
	}
	if want := "username=someone\npassword=secret\nhardcore=1\n"; string(data) != want {
		t.Fatalf("saved %q, want %q", data, want)
	}
}

func TestSetHardcoreNeedsTheFile(t *testing.T) {
	modes := testModes(t, "")
	recorder := serve(t, modes.HandleSetHardcore(), http.MethodPut, `{"hardcore":true}`)
	if recorder.Code != http.StatusNotFound {
		t.Fatalf("code = %d, want 404", recorder.Code)
	}
	if _, err := os.Stat(modes.path); !os.IsNotExist(err) {
		t.Fatalf("the file was created: %v", err)
	}
}

func TestSetHardcoreRejectsBadBodies(t *testing.T) {
	modes := testModes(t, loginFixture)
	for _, body := range []string{"", "{}", `{"hardcore":"yes"}`, "not json"} {
		recorder := serve(t, modes.HandleSetHardcore(), http.MethodPut, body)
		if recorder.Code != http.StatusBadRequest {
			t.Errorf("%q: code = %d, want 400", body, recorder.Code)
		}
	}
}

func hardcoreInFile(t *testing.T, modes *Modes) bool {
	t.Helper()
	data, err := os.ReadFile(modes.path)
	if err != nil {
		t.Fatal(err)
	}
	return strings.Contains(string(data), "hardcore=1")
}

// startCore stands in for Main starting the launched core.
func startCore(t *testing.T, modes *Modes, name string) {
	t.Helper()
	later := time.Now().Add(time.Second)
	if err := os.WriteFile(modes.coreName, []byte(name), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.Chtimes(modes.coreName, later, later); err != nil {
		t.Fatal(err)
	}
}

func waitSettled(t *testing.T, modes *Modes) {
	t.Helper()
	deadline := time.Now().Add(2 * time.Second)
	for modes.pending.Load() {
		if time.Now().After(deadline) {
			t.Fatal("the default was never put back")
		}
		time.Sleep(time.Millisecond)
	}
}

func on() *bool  { value := true; return &value }
func off() *bool { value := false; return &value }

func TestLaunchInTheOtherMode(t *testing.T) {
	modes := testModes(t, loginFixture)
	modes.startTimeout = time.Minute

	var seen bool
	err := modes.Launch(on(), func() error {
		seen = hardcoreInFile(t, modes)
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	if !seen {
		t.Fatal("the core load did not see hardcore")
	}
	if !hardcoreInFile(t, modes) {
		t.Fatal("the default came back before the core started")
	}
	if status := decodeStatus(t, serve(t, modes.HandleStatus(), http.MethodGet, "")); status.Hardcore {
		t.Fatal("status shows the launch's mode as the default")
	}

	startCore(t, modes, "RA_SNES")
	waitSettled(t, modes)
	if hardcoreInFile(t, modes) {
		t.Fatal("the default was not put back")
	}
}

func TestLaunchInTheDefaultModeWritesNothing(t *testing.T) {
	modes := testModes(t, loginFixture)
	before, err := os.Stat(modes.path)
	if err != nil {
		t.Fatal(err)
	}
	if launchErr := modes.Launch(off(), func() error { return nil }); launchErr != nil {
		t.Fatal(launchErr)
	}
	if modes.pending.Load() {
		t.Fatal("a restore is waiting for nothing")
	}
	after, err := os.Stat(modes.path)
	if err != nil {
		t.Fatal(err)
	}
	if !after.ModTime().Equal(before.ModTime()) {
		t.Fatal("the file was written")
	}
}

func TestLaunchWithoutTheFile(t *testing.T) {
	modes := testModes(t, "")
	launched := false
	if err := modes.Launch(on(), func() error { launched = true; return nil }); err != nil {
		t.Fatal(err)
	}
	if !launched {
		t.Fatal("the game was not launched")
	}
	if _, err := os.Stat(modes.path); !os.IsNotExist(err) {
		t.Fatalf("the file was created: %v", err)
	}
}

func TestFailedLaunchPutsTheDefaultBack(t *testing.T) {
	modes := testModes(t, loginFixture)
	modes.startTimeout = time.Minute
	failure := errors.New("no core")
	if err := modes.Launch(on(), func() error { return failure }); !errors.Is(err, failure) {
		t.Fatalf("err = %v", err)
	}
	if modes.pending.Load() || hardcoreInFile(t, modes) {
		t.Fatal("the launch's mode stayed in the file")
	}
}

func TestPlainLaunchNeverInheritsAMode(t *testing.T) {
	modes := testModes(t, loginFixture)
	modes.startTimeout = time.Minute
	if err := modes.Launch(on(), func() error { return nil }); err != nil {
		t.Fatal(err)
	}

	var seen bool
	if err := modes.Launch(nil, func() error { seen = hardcoreInFile(t, modes); return nil }); err != nil {
		t.Fatal(err)
	}
	if seen {
		t.Fatal("a launch without a mode started in the previous launch's mode")
	}

	recorder := httptest.NewRecorder()
	request := httptest.NewRequestWithContext(t.Context(), http.MethodPost, "/", http.NoBody)
	modes.Settled(func(http.ResponseWriter, *http.Request) {})(recorder, request)
	if modes.pending.Load() {
		t.Fatal("still waiting after a plain launch")
	}
}

func TestDefaultChangedWhileALaunchWaits(t *testing.T) {
	modes := testModes(t, loginFixture)
	modes.startTimeout = time.Minute
	if err := modes.Launch(on(), func() error { return nil }); err != nil {
		t.Fatal(err)
	}
	recorder := serve(t, modes.HandleSetHardcore(), http.MethodPut, `{"hardcore":true}`)
	if status := decodeStatus(t, recorder); !status.Hardcore {
		t.Fatalf("status = %+v, want the new default", status)
	}

	startCore(t, modes, "RA_SNES")
	waitSettled(t, modes)
	if !hardcoreInFile(t, modes) {
		t.Fatal("the new default was lost when the launch's mode came out")
	}
}

func TestDefaultComesBackWhenTheCoreNeverStarts(t *testing.T) {
	modes := testModes(t, loginFixture)
	modes.startTimeout = 20 * time.Millisecond
	if err := modes.Launch(on(), func() error { return nil }); err != nil {
		t.Fatal(err)
	}
	waitSettled(t, modes)
	if hardcoreInFile(t, modes) {
		t.Fatal("the launch's mode stayed in the file")
	}
}
