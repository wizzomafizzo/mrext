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

package mister

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/wizzomafizzo/mrext/pkg/config"
)

func TestStorageRootForPrefersTheLongestMatch(t *testing.T) {
	previous := config.StorageRoots
	config.StorageRoots = []string{"/media/usb0", "/media/network", "/media/fat/cifs"}
	t.Cleanup(func() { config.StorageRoots = previous })

	cases := map[string]string{
		"/media/usb0/games/NES/game.nes": "/media/usb0",
		"/media/usb0":                    "/media/usb0",
		"/media/network/games/x.bin":     "/media/network",
		// The SD card is always present, so nothing under it is removable.
		"/media/fat/games/NES/game.nes": "",
		"/media/fat":                    "",
		// cifs lives under the SD card; the longer root has to win.
		"/media/fat/cifs/games/x.bin": "/media/fat/cifs",
		// A sibling that merely shares a prefix is not on that storage.
		"/media/usb00/games/x.bin": "",
		"/media/usb0extra":         "",
	}
	for path, want := range cases {
		if got := StorageRootFor(path); got != want {
			t.Errorf("StorageRootFor(%q) = %q, want %q", path, got, want)
		}
	}
}

func TestTargetAvailableDistinguishesDetachedFromDeleted(t *testing.T) {
	media := t.TempDir()
	attached := filepath.Join(media, "usb0")
	unmounted := filepath.Join(media, "usb1")
	absent := filepath.Join(media, "usb2")

	if err := os.MkdirAll(filepath.Join(attached, "games"), 0o750); err != nil {
		t.Fatal(err)
	}
	// An unmounted mount point is left behind as an empty directory.
	if err := os.Mkdir(unmounted, 0o750); err != nil {
		t.Fatal(err)
	}

	previous := config.StorageRoots
	config.StorageRoots = []string{attached, unmounted, absent}
	t.Cleanup(func() { config.StorageRoots = previous })

	if !TargetAvailable(filepath.Join(attached, "games", "gone.nes")) {
		t.Error("a deleted file on attached storage must count as available")
	}
	if TargetAvailable(filepath.Join(unmounted, "games", "game.nes")) {
		t.Error("an empty mount point means the drive is not attached")
	}
	if TargetAvailable(filepath.Join(absent, "games", "game.nes")) {
		t.Error("a missing mount point means the drive is not attached")
	}
	if !TargetAvailable("/media/fat/games/NES/gone.nes") {
		t.Error("the SD card is always available")
	}
}
