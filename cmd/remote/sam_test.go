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

package main

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
)

func TestSAMActivityUsesExistingFileOnly(t *testing.T) {
	root := t.TempDir()
	missing := filepath.Join(root, "absent", "SAM_Joy_Activity")
	if err := signalSAMActivity(missing); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Dir(missing)); !os.IsNotExist(err) {
		t.Fatalf("created absent SAM directory: %v", err)
	}
	path := filepath.Join(root, "SAM_Joy_Activity")
	if err := os.WriteFile(path, []byte("previous longer message"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := signalSAMActivity(path); err != nil {
		t.Fatal(err)
	}
	// #nosec G304 -- path is beneath t.TempDir.
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != "zaparoo\n" {
		t.Fatalf("unexpected activity token: %q", data)
	}
	if err := signalSAMActivity(root); err == nil {
		t.Fatal("accepted directory as activity file")
	}
	link := filepath.Join(root, "link")
	if err := os.Symlink(path, link); err != nil {
		t.Fatal(err)
	}
	if err := signalSAMActivity(link); err == nil {
		t.Fatal("accepted symlink as activity file")
	}
}

func TestSAMNotifiedOnlyAfterSuccessfulLaunch(t *testing.T) {
	for _, status := range []int{0, http.StatusOK, http.StatusBadRequest, http.StatusInternalServerError} {
		launched, notified := false, false
		handler := afterSuccessfulLaunch(func(w http.ResponseWriter, _ *http.Request) {
			launched = true
			if status != 0 {
				w.WriteHeader(status)
			}
		}, func() error {
			if !launched {
				t.Fatal("notified before launch")
			}
			notified = true
			return nil
		}, func(string, ...any) { t.Fatal("unexpected notification failure") })
		response := httptest.NewRecorder()
		request := httptest.NewRequestWithContext(t.Context(), http.MethodPost, "/launch", http.NoBody)
		handler.ServeHTTP(response, request)
		if notified != (status == 0 || status == http.StatusOK) {
			t.Fatalf("status=%d notified=%t", status, notified)
		}
	}
}

func TestSAMFailureDoesNotChangeLaunchResponse(t *testing.T) {
	logged := false
	handler := afterSuccessfulLaunch(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte("launched"))
	}, func() error { return errors.New("activity unavailable") }, func(string, ...any) { logged = true })
	response := httptest.NewRecorder()
	request := httptest.NewRequestWithContext(t.Context(), http.MethodPost, "/launch", http.NoBody)
	handler.ServeHTTP(response, request)
	if response.Code != http.StatusOK || response.Body.String() != "launched" || !logged {
		t.Fatalf("status=%d body=%q logged=%t", response.Code, response.Body.String(), logged)
	}
}
