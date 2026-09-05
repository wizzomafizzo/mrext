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
	"testing"
)

// Expected output captured from Python 3's ConfigParser.write().
func TestParseINIMatchesConfigParser(t *testing.T) {
	source := "[bgm]\nPlayback = Random\nplaylist: none\n  ; indented comment\n# c2\nkeep_me = x\nmulti = a\n  b\n" +
		"[Other]\nz = 1\n"
	doc := ParseINI(source)
	doc.Set("bgm", "menuvolume", "3")
	want := "[bgm]\nplayback = Random\nplaylist = none\nkeep_me = x\nmulti = a\n\tb\nmenuvolume = 3\n\n" +
		"[Other]\nz = 1\n\n"
	if got := doc.String(); got != want {
		t.Fatalf("rendered:\n%q\nwant:\n%q", got, want)
	}
	if value, ok := doc.Get("bgm", "PLAYBACK"); !ok || value != "Random" {
		t.Fatalf("case-insensitive key lookup failed: %q %v", value, ok)
	}
	if value, ok := doc.Get("bgm", "multi"); !ok || value != "a\nb" {
		t.Fatalf("continuation value = %q", value)
	}
}

func TestDefaultINIRoundTripAddsOnlyTrailingBlankLine(t *testing.T) {
	if got := ParseINI(DefaultINI).String(); got != DefaultINI+"\n" {
		t.Fatalf("round trip changed the default file:\n%q", got)
	}
}

func TestParseINISectionsAreCaseSensitiveAndValuesMayBeEmpty(t *testing.T) {
	doc := ParseINI("[BGM]\nplayback = loop\n[bgm]\nplaylist =\n")
	if _, ok := doc.Get("bgm", "playback"); ok {
		t.Fatal("[BGM] must not satisfy a lookup for [bgm]")
	}
	value, ok := doc.Get("bgm", "playlist")
	if !ok || value != "" {
		t.Fatalf("empty value = %q ok=%v", value, ok)
	}
	if got := doc.String(); got != "[BGM]\nplayback = loop\n\n[bgm]\nplaylist = \n\n" {
		t.Fatalf("rendered %q", got)
	}
}

func TestParseINISkipsLinesPythonWouldRejectAndKeepsLastDuplicate(t *testing.T) {
	doc := ParseINI("orphan = 1\n[bgm]\nno delimiter here\n= novalue\nplayback = random\nplayback = loop\n")
	if value, _ := doc.Get("bgm", "playback"); value != "loop" {
		t.Fatalf("duplicate key should keep last value, got %q", value)
	}
	if got := doc.String(); got != "[bgm]\nplayback = loop\n\n" {
		t.Fatalf("rendered %q", got)
	}
}

func TestParseINIPercentEscapes(t *testing.T) {
	doc := ParseINI("[bgm]\nplaylist = 100%% Hits\n")
	if value, _ := doc.Get("bgm", "playlist"); value != "100% Hits" {
		t.Fatalf("unescaped value = %q", value)
	}
	doc.Set("bgm", "playlist", "50% Off")
	if got := doc.String(); got != "[bgm]\nplaylist = 50%% Off\n\n" {
		t.Fatalf("rendered %q", got)
	}
}

func TestParseINIBlankLinesInsideValues(t *testing.T) {
	doc := ParseINI("[bgm]\nmulti = a\n\n  b\n\n\nnext = c\n")
	if value, _ := doc.Get("bgm", "multi"); value != "a\n\nb" {
		t.Fatalf("multi = %q", value)
	}
	if value, _ := doc.Get("bgm", "next"); value != "c" {
		t.Fatalf("next = %q", value)
	}
}

func TestReadINIMissingFileIsEmptyAndWriteFileIsAtomic(t *testing.T) {
	path := filepath.Join(t.TempDir(), "bgm.ini")
	doc, err := ReadINI(path)
	if err != nil {
		t.Fatal(err)
	}
	if doc.HasSection("bgm") {
		t.Fatal("missing file produced a section")
	}
	doc.Set("bgm", "playback", "loop")
	if writeErr := doc.WriteFile(path); writeErr != nil {
		t.Fatal(writeErr)
	}
	if got := readFile(t, path); got != "[bgm]\nplayback = loop\n\n" {
		t.Fatalf("written %q", got)
	}
	entries, err := os.ReadDir(filepath.Dir(path))
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 1 {
		t.Fatalf("temporary file left behind: %d entries", len(entries))
	}
}
