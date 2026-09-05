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
	"path/filepath"
	"testing"
)

func TestFileTypePredicates(t *testing.T) {
	valid := []string{"a.MP3", "b.pls", "c.Ogg", "d.wav", "e.mid", "f.vgm", "g.VGZ", "h.vgm.gz", "dir.mp3"}
	for _, name := range valid {
		if !IsValidFile(name) {
			t.Fatalf("%s should be valid", name)
		}
	}
	for _, name := range []string{"a.midi", "b.gz", "c.flac", "d.mp3.bak", "e"} {
		if IsValidFile(name) {
			t.Fatalf("%s should be rejected", name)
		}
	}
	if !IsVGM("x.vgm.gz") || IsVGM("x.gz") || IsMP3("x.pls") {
		t.Fatal("suffix checks are wrong")
	}
}

func TestLoopAmount(t *testing.T) {
	cases := map[string]int{
		"X05_song.mp3": 5, "X23_song.mp3": 23, "X00_song.mp3": 0, "X5_song.mp3": 1, "x05_song.mp3": 1,
		"song.mp3": 1, "X05song.mp3": 1, filepath.Join("X07_folder", "song.mp3"): 1,
	}
	for name, want := range cases {
		if got := LoopAmount(name); got != want {
			t.Fatalf("LoopAmount(%q) = %d want %d", name, got, want)
		}
	}
}

func TestPLSURL(t *testing.T) {
	paths := newTestPaths(t)
	logger, out := newTestLogger(&paths)
	writeINI(t, &paths, "[bgm]\ndebug = yes\n")
	radio := filepath.Join(paths.MusicFolder, "radio.pls")
	writeFile(t, radio, "[playlist]\nFile1=http://example.com/stream\r\nTitle1=Example\n")
	if got := PLSURL(radio, logger); got != "http://example.com/stream\r" {
		t.Fatalf("url = %q", got)
	}
	empty := filepath.Join(paths.MusicFolder, "empty.pls")
	writeFile(t, empty, "[playlist]\nTitle1=Example\n")
	if got := PLSURL(empty, logger); got != "" {
		t.Fatalf("expected no url, got %q", got)
	}
	if out.String() != "Playlist URL not found\n" {
		t.Fatalf("log output %q", out.String())
	}
}
