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
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func writeRAFixture(t *testing.T, content string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "retroachievements.cfg")
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
	return path
}

func TestParseRAHardcoreMatchesMain(t *testing.T) {
	tests := []struct {
		name    string
		content string
		want    RAHardcore
	}{
		{"missing key is off", "username=someone\n", RAHardcore{}},
		{"on", "hardcore=1\n", RAHardcore{Hardcore: true}},
		{"off", "hardcore=0\n", RAHardcore{}},
		{"any non-zero number is on", "hardcore=2\n", RAHardcore{Hardcore: true}},
		{"negative is on", "hardcore=-1\n", RAHardcore{Hardcore: true}},
		{"leading spaces in the value", "hardcore=  \t1\n", RAHardcore{Hardcore: true}},
		{"digits then text", "hardcore=1 # on\n", RAHardcore{Hardcore: true}},
		{"text is off", "hardcore=yes\n", RAHardcore{}},
		{"key case does not matter", "HardCore=1\n", RAHardcore{Hardcore: true}},
		// Main does not trim the key, so this line is not the hardcore key.
		{"spaced key is ignored", "hardcore = 1\n", RAHardcore{}},
		{"last one wins", "hardcore=1\nhardcore=0\n", RAHardcore{}},
		{"comment in column 0", "#hardcore=1\n", RAHardcore{}},
		{"indented hash is not a comment", " #x=1\nhardcore=1\n", RAHardcore{Hardcore: true}},
		{"CRLF", "hardcore=1\r\nforce_hardcore=1\r\n", RAHardcore{Hardcore: true, Forced: true}},
		{"no final newline", "hardcore=1", RAHardcore{Hardcore: true}},
		{"forced", "force_hardcore=1\n", RAHardcore{Forced: true}},
		{
			"long line is split like fgets",
			"comment=" + strings.Repeat("x", raLineSize-len("comment=")) + "hardcore=1\n",
			RAHardcore{Hardcore: true},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := parseRAHardcore([]byte(tt.content)); got != tt.want {
				t.Fatalf("got %+v, want %+v", got, tt.want)
			}
		})
	}
}

func TestReadRAHardcoreMissingFile(t *testing.T) {
	_, err := ReadRAHardcore(filepath.Join(t.TempDir(), "retroachievements.cfg"))
	if !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("err = %v, want os.ErrNotExist", err)
	}
}

func TestSetRAHardcoreKeepsTheRestOfTheFile(t *testing.T) {
	tests := []struct {
		name      string
		content   string
		want      string
		enabled   bool
		unchanged bool
	}{
		{
			name:    "replaces only the value",
			content: "username=someone\npassword=secret\n# hardcore=1 is strict\nhardcore=0\ndebug=0\n",
			enabled: true,
			want:    "username=someone\npassword=secret\n# hardcore=1 is strict\nhardcore=1\ndebug=0\n",
		},
		{
			name:    "keeps CRLF and the key's spelling",
			content: "username=someone\r\nHardcore= 1\r\n",
			enabled: false,
			want:    "username=someone\r\nHardcore= 0\r\n",
		},
		{
			name:    "changes the line Main uses",
			content: "hardcore=0\nhardcore=0\n",
			enabled: true,
			want:    "hardcore=0\nhardcore=1\n",
		},
		{
			name:    "appends a missing key",
			content: "username=someone\npassword=secret",
			enabled: true,
			want:    "username=someone\npassword=secret\nhardcore=1\n",
		},
		{
			name:    "appends with CRLF",
			content: "username=someone\r\n",
			enabled: true,
			want:    "username=someone\r\nhardcore=1\r\n",
		},
		{
			name:    "keeps the length of the digits",
			content: "hardcore=+01 ; strict\n",
			enabled: false,
			want:    "hardcore=+00 ; strict\n",
		},
		{
			name:    "appends when the value has no digits",
			content: "hardcore=yes\n",
			enabled: true,
			want:    "hardcore=yes\nhardcore=1\n",
		},
		{
			name:    "keeps where a long line splits",
			content: "hardcore=0" + strings.Repeat("x", 501) + "force_hardcore=1\n",
			enabled: true,
			want:    "hardcore=1" + strings.Repeat("x", 501) + "force_hardcore=1\n",
		},
		{
			name:      "leaves a matching value alone",
			content:   "hardcore=1 ; on\n",
			enabled:   true,
			want:      "hardcore=1 ; on\n",
			unchanged: true,
		},
		{
			name:      "a missing key is already off",
			content:   "username=someone\n",
			enabled:   false,
			want:      "username=someone\n",
			unchanged: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			path := writeRAFixture(t, tt.content)
			changed, err := SetRAHardcore(path, tt.enabled)
			if err != nil {
				t.Fatal(err)
			}
			if changed == tt.unchanged {
				t.Fatalf("changed = %v", changed)
			}
			if got := readFileString(t, path); got != tt.want {
				t.Fatalf("saved:\n%q\nwant:\n%q", got, tt.want)
			}
			settings, err := ReadRAHardcore(path)
			if err != nil {
				t.Fatal(err)
			}
			if settings.Hardcore != tt.enabled {
				t.Fatalf("read back %v, want %v", settings.Hardcore, tt.enabled)
			}
			if want := parseRAHardcore([]byte(tt.content)).Forced; settings.Forced != want {
				t.Fatalf("force_hardcore read back %v, want %v", settings.Forced, want)
			}
		})
	}
}

func TestSetRAHardcoreNeverCreatesTheFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "retroachievements.cfg")
	if _, err := SetRAHardcore(path, true); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("err = %v, want os.ErrNotExist", err)
	}
	if _, err := os.Stat(path); !os.IsNotExist(err) {
		t.Fatalf("the file was created: %v", err)
	}
}

func TestSetRAHardcoreLeavesNoStagedFile(t *testing.T) {
	path := writeRAFixture(t, "hardcore=0\n")
	if _, err := SetRAHardcore(path, true); err != nil {
		t.Fatal(err)
	}
	entries, err := os.ReadDir(filepath.Dir(path))
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 1 {
		t.Fatalf("folder holds %d entries, want only the settings file", len(entries))
	}
}
