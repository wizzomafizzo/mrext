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

package settings

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/wizzomafizzo/mrext/pkg/mister"
)

func TestExpectedFilenameGuard(t *testing.T) {
	mi := mister.MisterIni{Id: 4, Filename: "MiSTer_ultrawide.ini"}
	for _, expected := range []string{"", mi.Filename, "MiSTer_auto.ini"} {
		request := httptest.NewRequestWithContext(t.Context(), http.MethodPut, "/settings/inis/4", http.NoBody)
		request.Header.Set(mister.IniFilenameHeader, expected)
		response := httptest.NewRecorder()
		accepted := checkIniFilename(response, request, mi)
		want := expected == "" || expected == mi.Filename
		if accepted != want || (!want && response.Code != http.StatusConflict) {
			t.Fatalf("expected=%q accepted=%t status=%d", expected, accepted, response.Code)
		}
	}
}

func TestRelaunchOnlyAfterSuccessfulINISave(t *testing.T) {
	root := t.TempDir()
	mi := &mister.MisterIni{Filename: mister.DefaultIniFilename, Path: filepath.Join(root, mister.DefaultIniFilename)}
	if err := mi.Load(); err != nil {
		t.Fatal(err)
	}
	if err := mi.SetKey("vscale_mode", "2"); err != nil {
		t.Fatal(err)
	}
	calls := 0
	relaunch := func() error {
		calls++
		if _, err := os.Stat(mi.Path); err != nil {
			t.Fatalf("relaunch before save: %v", err)
		}
		return nil
	}
	if err := saveAndRelaunch(mi, relaunch); err != nil || calls != 1 {
		t.Fatalf("successful save: calls=%d err=%v", calls, err)
	}
	mi.Path = filepath.Join(root, "missing-directory", mister.DefaultIniFilename)
	if err := saveAndRelaunch(mi, relaunch); err == nil || calls != 1 {
		t.Fatalf("failed save triggered relaunch: calls=%d err=%v", calls, err)
	}
}
