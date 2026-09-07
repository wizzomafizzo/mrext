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
package mister

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestUpdateConfiguredMacAddressRejectsInjection(t *testing.T) {
	// The value arrives from an unauthenticated HTTP request and is written
	// into the file U-Boot reads, so a newline used to append arbitrary boot
	// parameters.
	rejected := []string{
		"02:00:00:00:00:01\nbootargs=init=/bin/sh",
		"not a mac",
		"",
		"02:00:00:00:00",
		"02-00-00-00-00-01-extra",
	}
	for _, value := range rejected {
		if err := UpdateConfiguredMacAddress(value); err == nil {
			t.Errorf("accepted %q", value)
		}
	}
}

func TestUpdateUBootParamKeepsCommentsAndOrder(t *testing.T) {
	original := "# board settings\nbootargs=console=ttyS0\nethaddr=02:00:00:00:00:01\n" +
		"; keep me\nvideo=1280x720\n"
	path := filepath.Join(t.TempDir(), "u-boot.txt")
	if err := os.WriteFile(path, []byte(original), 0o644); err != nil {
		t.Fatal(err)
	}

	previous := ubootPath
	ubootPath = path
	t.Cleanup(func() { ubootPath = previous })

	if err := UpdateUBootParam("ethaddr", "02:00:00:00:00:02"); err != nil {
		t.Fatal(err)
	}

	saved := readFileString(t, path)
	want := "# board settings\nbootargs=console=ttyS0\nethaddr=02:00:00:00:00:02\n" +
		"; keep me\nvideo=1280x720\n"
	if saved != want {
		t.Fatalf("saved:\n%q\nwant:\n%q", saved, want)
	}

	// The original is kept once, and not overwritten by later saves.
	backup := readFileString(t, path+".backup")
	if backup != original {
		t.Fatalf("backup = %q, want the original", backup)
	}
	if err := UpdateUBootParam("ethaddr", "02:00:00:00:00:03"); err != nil {
		t.Fatal(err)
	}
	if again := readFileString(t, path+".backup"); again != original {
		t.Fatalf("a second save replaced the backup: %q", again)
	}
}

func TestUpdateUBootParamAppendsAMissingKey(t *testing.T) {
	path := filepath.Join(t.TempDir(), "u-boot.txt")
	if err := os.WriteFile(path, []byte("bootargs=quiet\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	previous := ubootPath
	ubootPath = path
	t.Cleanup(func() { ubootPath = previous })

	if err := UpdateUBootParam("ethaddr", "02:00:00:00:00:01"); err != nil {
		t.Fatal(err)
	}
	saved := readFileString(t, path)
	if !strings.HasPrefix(saved, "bootargs=quiet\n") || !strings.Contains(saved, "ethaddr=02:00:00:00:00:01") {
		t.Fatalf("saved = %q", saved)
	}
}

func readFileString(t *testing.T, path string) string {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	return string(data)
}
