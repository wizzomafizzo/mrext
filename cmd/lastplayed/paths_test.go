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

//nolint:gosec // All filesystem operations use temporary fixture roots.
package main

import (
	"bytes"
	"encoding/xml"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/wizzomafizzo/mrext/pkg/config"
)

func TestLastPlayedPathsDoNotDependOnLauncherFolder(t *testing.T) {
	for _, fixture := range []struct{ system, game, folder string }{
		{system: "NES", game: "Mario & Luigi's (USA) [!].nes"},
		{system: "PSX", game: "A Game (Disc 1) [!].chd", folder: "History & [Recent]!"},
	} {
		t.Run(fixture.system, func(t *testing.T) {
			root := t.TempDir()
			iniPath := filepath.Join(root, "lastplayed.ini")
			text := "[lastplayed]\nrecent_folder_name = " + fixture.folder + "\n"
			if err := os.WriteFile(iniPath, []byte(text), 0o600); err != nil {
				t.Fatal(err)
			}
			cfg, err := config.LoadUserConfigAt(iniPath, filepath.Join(root, "lastplayed.sh"), &config.UserConfig{})
			if err != nil {
				t.Fatal(err)
			}
			cfg.Systems.GamesFolder = []string{root}
			game := filepath.Join(root, "games", fixture.system, "Deep", "More games", fixture.game)
			if err = os.MkdirAll(filepath.Dir(game), 0o755); err != nil {
				t.Fatal(err)
			}
			if err = os.WriteFile(game, []byte("fixture"), 0o600); err != nil {
				t.Fatal(err)
			}
			if err = createLastPlayedMgl(cfg, game, root); err != nil {
				t.Fatal(err)
			}
			if err = addToRecentFolder(cfg, game, root); err != nil {
				t.Fatal(err)
			}
			folder := fixture.folder
			if folder == "" {
				folder = defaultRecentFolderName
			}
			recentDir := filepath.Join(root, "_"+folder)
			files, err := os.ReadDir(recentDir)
			if err != nil {
				t.Fatal(err)
			}
			var entries []string
			for _, file := range files {
				if filepath.Ext(file.Name()) == ".mgl" {
					entries = append(entries, filepath.Join(recentDir, file.Name()))
				}
			}
			if len(entries) != 1 {
				t.Fatalf("entries=%v", entries)
			}
			last, err := os.ReadFile(filepath.Join(root, "Last Played.mgl"))
			if err != nil {
				t.Fatal(err)
			}
			recent, err := os.ReadFile(entries[0])
			if err != nil {
				t.Fatal(err)
			}
			if !bytes.Equal(last, recent) {
				t.Fatal("root and recent launchers differ")
			}
			assertMGLTarget(t, last, fixture.system, game)

			// A replay regenerates an old malformed entry without touching user launchers.
			if err = os.WriteFile(entries[0], []byte("old invalid launcher"), 0o600); err != nil {
				t.Fatal(err)
			}
			old := time.Unix(1, 0)
			if err = os.Chtimes(entries[0], old, old); err != nil {
				t.Fatal(err)
			}
			userFile := filepath.Join(recentDir, "Keep.mgl")
			if err = os.WriteFile(userFile, []byte("user content"), 0o600); err != nil {
				t.Fatal(err)
			}
			if err = addToRecentFolder(cfg, game, root); err != nil {
				t.Fatal(err)
			}
			repaired, err := os.ReadFile(entries[0])
			if err != nil || !bytes.Equal(last, repaired) {
				t.Fatalf("replay did not regenerate entry: %s %v", repaired, err)
			}
			kept, err := os.ReadFile(userFile)
			if err != nil || string(kept) != "user content" {
				t.Fatal("unrelated launcher modified")
			}
			if fixture.folder != "" {
				if _, statErr := os.Stat(filepath.Join(root, "_Recently Played")); !os.IsNotExist(statErr) {
					t.Fatal("configured folder was ignored")
				}
			}
		})
	}
}

type mglMediaFixture struct {
	Path string `xml:"path,attr"`
}

func assertMGLTarget(t *testing.T, data []byte, system, target string) {
	t.Helper()
	var mgl struct {
		RBF   string            `xml:"rbf"`
		Files []mglMediaFixture `xml:"file"`
	}
	if err := xml.Unmarshal(data, &mgl); err != nil {
		t.Fatal(err)
	}
	if mgl.RBF != "_Console/"+system {
		t.Fatalf("RBF %q", mgl.RBF)
	}
	if len(mgl.Files) != 1 {
		t.Fatalf("media slots: %+v", mgl.Files)
	}
	// Main_MiSTer 915ca339: menu.cpp resolves media against HomeDir, which
	// user_io.h maps to user_io_get_core_path. MGL destination is not a base.
	// These are lexical checks only; no real /media paths are read or written.
	for _, home := range []string{"/media/fat/games/" + system, "/media/usb0/games/" + system} {
		resolved := filepath.Clean(home + "/" + mgl.Files[0].Path)
		if resolved != target {
			t.Fatalf("home=%s media=%s resolved=%s want=%s", home, mgl.Files[0].Path, resolved, target)
		}
	}
}
