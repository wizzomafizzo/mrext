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

package menu

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"github.com/wizzomafizzo/mrext/pkg/config"
	"github.com/wizzomafizzo/mrext/pkg/service"
)

func TestResolveMenuPathRejectsTraversal(t *testing.T) {
	// Every one of these resolved outside the SD root before the fix, because
	// cleanPath cleaned the relative path first and then joined it, so leading
	// ".." survived into the join.
	escapes := []string{
		"../../etc/passwd",
		"..",
		"../",
		"_Arcade/../../../etc/passwd",
		"/etc/passwd",
		"/media/usb0/games",
		"./../../etc/passwd",
		config.SdFolder + "/../../etc/passwd",
	}
	for _, path := range escapes {
		resolved, err := resolveMenuPath(path)
		if err == nil {
			t.Errorf("%q resolved to %q, want rejection", path, resolved)
			continue
		}
		if !errors.Is(err, errOutsideMenuRoot) {
			t.Errorf("%q: got %v, want an outside-root error", path, err)
		}
	}
}

func TestResolveMenuPathRejectsProtectedPaths(t *testing.T) {
	// Deleting any of these leaves a MiSTer that will not boot, or wipes every
	// core configuration. Only the MiSTer prefix was guarded before, and only
	// in the delete handler.
	protected := []string{
		"MiSTer",
		"MiSTer.ini",
		"MiSTer_20240101.rbf",
		"menu.rbf",
		"u-boot.img",
		"u-boot.txt",
		"linux",
		"linux/user-startup.sh",
		"config",
		"config/NES.cfg",
		"MISTER.INI",
		config.SdFolder + "/linux/user-startup.sh",
	}
	for _, path := range protected {
		resolved, err := resolveMenuPath(path)
		if err == nil {
			t.Errorf("%q resolved to %q, want rejection", path, resolved)
			continue
		}
		if !errors.Is(err, errProtectedPath) {
			t.Errorf("%q: got %v, want a protected-path error", path, err)
		}
	}
}

func TestResolveMenuPathAcceptsMenuPaths(t *testing.T) {
	accepted := map[string]string{
		"":                           config.SdFolder,
		".":                          config.SdFolder,
		"_Arcade":                    filepath.Join(config.SdFolder, "_Arcade"),
		"_Arcade/Pac-Man.mra":        filepath.Join(config.SdFolder, "_Arcade", "Pac-Man.mra"),
		"_Console/_NES/Game [!].mgl": filepath.Join(config.SdFolder, "_Console", "_NES", "Game [!].mgl"),
		config.SdFolder:              config.SdFolder,
		config.SdFolder + "/_Arcade": filepath.Join(config.SdFolder, "_Arcade"),
		"_Games/_NES/../_SNES":       filepath.Join(config.SdFolder, "_Games", "_SNES"),
		"configuration-notes.mgl":    filepath.Join(config.SdFolder, "configuration-notes.mgl"),
		"linux-tips.mgl":             filepath.Join(config.SdFolder, "linux-tips.mgl"),
	}
	for path, want := range accepted {
		resolved, err := resolveMenuPath(path)
		if err != nil {
			t.Errorf("%q: unexpected error %v", path, err)
			continue
		}
		if resolved != want {
			t.Errorf("%q resolved to %q, want %q", path, resolved, want)
		}
	}
}

func TestRelativeMenuPathKeepsListingLinksStable(t *testing.T) {
	// ListFolder puts this value straight into the parent, next and up links,
	// so an empty path must stay empty rather than becoming ".".
	cases := map[string]string{
		"":                           "",
		".":                          ".",
		"_Arcade":                    "_Arcade",
		config.SdFolder:              "",
		config.SdFolder + "/_Arcade": "_Arcade",
		"_Arcade/":                   "_Arcade",
	}
	for path, want := range cases {
		if got := relativeMenuPath(path); got != want {
			t.Errorf("relativeMenuPath(%q) = %q, want %q", path, got, want)
		}
	}
}

func TestDeleteHandlerRejectsTraversalBeforeTouchingDisk(t *testing.T) {
	// The handler used to reach os.RemoveAll for any ordinary file anywhere on
	// the device. The traversal targets here point into a temp dir rather than
	// a real system path, so a regression in the guard deletes a throwaway file
	// instead of something that matters.
	logger := service.NewLogger("mrext-menu-path-test")
	handler := HandleDeleteFile(logger)

	bait := filepath.Join(t.TempDir(), "bait")
	if err := os.WriteFile(bait, []byte("keep me"), 0o600); err != nil {
		t.Fatal(err)
	}
	escape, err := filepath.Rel(config.SdFolder, bait)
	if err != nil {
		t.Fatal(err)
	}

	paths := []string{escape, bait, "u-boot.img", "linux/user-startup.sh", "MiSTer.ini"}
	for _, path := range paths {
		body := strings.NewReader(`{"path":` + strconv.Quote(path) + `}`)
		request := httptest.NewRequestWithContext(
			t.Context(), http.MethodPost, "/api/menu/files/delete", body)
		recorder := httptest.NewRecorder()

		handler(recorder, request)

		if recorder.Code != http.StatusBadRequest {
			t.Errorf("delete %q returned %d, want %d", path, recorder.Code, http.StatusBadRequest)
		}
	}

	if _, err := os.Stat(bait); err != nil {
		t.Fatalf("delete request escaped the menu root: %v", err)
	}
}
