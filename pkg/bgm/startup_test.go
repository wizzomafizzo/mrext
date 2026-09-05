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

//nolint:gosec // Tests operate only on paths rooted in temporary directories.
package bgm

import (
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"
)

func TestTryAddToStartupCreatesAndAppendsOnce(t *testing.T) {
	paths := newTestPaths(t)
	paths.AppPath = "/media/fat/Scripts/bgm.sh"
	logger, out := newTestLogger(&paths)
	if err := TryAddToStartup(&paths, logger); err != nil {
		t.Fatal(err)
	}
	want := "#!/bin/sh\n\n# Startup BGM\n[[ -e /media/fat/Scripts/bgm.sh ]] && /media/fat/Scripts/bgm.sh $1\n"
	if got := readFile(t, paths.StartupScript); got != want {
		t.Fatalf("startup script %q", got)
	}
	if out.String() != "Added service to startup script.\n" {
		t.Fatalf("output %q", out.String())
	}
	if err := TryAddToStartup(&paths, logger); err != nil {
		t.Fatal(err)
	}
	if got := readFile(t, paths.StartupScript); got != want {
		t.Fatal("entry must only be added once")
	}
}

func TestTryAddToStartupAppendsToExistingScript(t *testing.T) {
	paths := newTestPaths(t)
	paths.AppPath = filepath.Join(t.TempDir(), "my scripts", "bgm.sh")
	writeFile(t, paths.StartupScript, "#!/bin/sh\necho hi\n")
	logger, _ := newTestLogger(&paths)
	if err := TryAddToStartup(&paths, logger); err != nil {
		t.Fatal(err)
	}
	got := readFile(t, paths.StartupScript)
	if !strings.HasPrefix(got, "#!/bin/sh\necho hi\n\n# Startup BGM\n[[ -e \"") {
		t.Fatalf("startup script %q", got)
	}
	info, err := os.Stat(paths.StartupScript)
	if err != nil {
		t.Fatal(err)
	}
	if info.Mode().Perm()&0o100 == 0 {
		t.Log("existing permissions are kept as-is")
	}
}

func TestPlaylistsListsFoldersExceptBoot(t *testing.T) {
	paths := newTestPaths(t)
	for _, folder := range []string{"zeta", "alpha", "boot", "Mid"} {
		if err := os.MkdirAll(filepath.Join(paths.MusicFolder, folder), 0o755); err != nil {
			t.Fatal(err)
		}
	}
	touchTracks(t, paths.MusicFolder, "file.mp3")
	playlists, err := Playlists(&paths)
	if err != nil {
		t.Fatal(err)
	}
	if !slices.Equal(playlists, []string{"Mid", "alpha", "zeta"}) {
		t.Fatalf("playlists %v", playlists)
	}
}

func TestEnsureMusicFolder(t *testing.T) {
	paths := RootedPaths(t.TempDir())
	created, err := EnsureMusicFolder(&paths)
	if err != nil || !created {
		t.Fatalf("created=%v err=%v", created, err)
	}
	created, err = EnsureMusicFolder(&paths)
	if err != nil || created {
		t.Fatalf("second call created=%v err=%v", created, err)
	}
}
