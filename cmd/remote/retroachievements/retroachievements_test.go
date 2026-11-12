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
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

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
