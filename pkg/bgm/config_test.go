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
	"testing"
)

func TestParsePythonBool(t *testing.T) {
	accepted := map[string]bool{
		"1": true, "yes": true, "TRUE": true, "On": true,
		"0": false, "No": false, "false": false, "OFF": false,
	}
	for value, want := range accepted {
		got, valid := parsePythonBool(value)
		if !valid || got != want {
			t.Fatalf("%q -> %v valid=%v", value, got, valid)
		}
	}
	for _, value := range []string{"y", "t", "maybe", ""} {
		if _, valid := parsePythonBool(value); valid {
			t.Fatalf("%q should be rejected like configparser", value)
		}
	}
}

func TestParsePythonNumbers(t *testing.T) {
	ints := []struct {
		value string
		want  int
	}{{" 7 ", 7}, {"+3", 3}, {"-1", -1}, {"1_0", 10}}
	for _, item := range ints {
		got, valid := parsePythonInt(item.value)
		if !valid || got != item.want {
			t.Fatalf("int %q -> %d valid=%v", item.value, got, valid)
		}
	}
	for _, value := range []string{"3.0", "", "x", "1__0", "_1"} {
		if _, valid := parsePythonInt(value); valid {
			t.Fatalf("int %q should be rejected", value)
		}
	}
	floats := []struct {
		value string
		want  float64
	}{{"0", 0}, {"0.5", 0.5}, {".5", 0.5}, {"2.", 2}, {"1e1", 10}, {" 1_000.5 ", 1000.5}}
	for _, item := range floats {
		got, valid := parsePythonFloat(item.value)
		if !valid || got != item.want {
			t.Fatalf("float %q -> %v valid=%v", item.value, got, valid)
		}
	}
	for _, value := range []string{"inf", "nan", "", "abc", "."} {
		if _, valid := parsePythonFloat(value); valid {
			t.Fatalf("float %q should be rejected", value)
		}
	}
}

func TestConfigFromDocumentAppliesFallbacks(t *testing.T) {
	doc := ParseINI("[bgm]\nplayback = Loop\nplaylist = none\nstartup = maybe\nplayincore = ON\n" +
		"corebootdelay = 1.5\nmenuvolume = seven\ndefaultvolume = 4\n[tui]\ntheme = nord\nmouse = no\n")
	cfg := ConfigFromDocument(doc)
	if cfg.Playback != "Loop" {
		t.Fatalf("playback must keep its case, got %q", cfg.Playback)
	}
	if !cfg.Playlist.IsNone() {
		t.Fatal("playlist none should parse as no playlist")
	}
	if !cfg.Startup {
		t.Fatal("invalid startup should fall back to the default true")
	}
	if !cfg.PlayInCore || cfg.CoreBootDelay != 1.5 || cfg.MenuVolume != -1 || cfg.DefaultVolume != 4 {
		t.Fatalf("unexpected config %+v", cfg)
	}
	if cfg.TUI.Theme != "nord" || cfg.TUI.Mouse || !cfg.TUI.CRTMode {
		t.Fatalf("unexpected tui options %+v", cfg.TUI)
	}
	empty := ConfigFromDocument(ParseINI("[bgm]\nplaylist =\n"))
	if empty.Playlist.IsNone() || empty.Playlist.Name() != "" || empty.Playlist.String() != "" {
		t.Fatalf("empty playlist must be a distinct named playlist: %+v", empty.Playlist)
	}
	if ParsePlaylist("None").IsNone() {
		t.Fatal("only lowercase none means no playlist")
	}
}

func TestLoadConfigWritesDefaultOnlyWhenMusicFolderExists(t *testing.T) {
	paths := newTestPaths(t)
	cfg, err := LoadConfig(&paths)
	if err != nil {
		t.Fatal(err)
	}
	if got := readFile(t, paths.IniFile); got != DefaultINI {
		t.Fatalf("default ini = %q", got)
	}
	if cfg != DefaultConfig() {
		t.Fatalf("config %+v != defaults", cfg)
	}

	missing := RootedPaths(t.TempDir())
	if _, err := LoadConfig(&missing); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(missing.IniFile); !os.IsNotExist(err) {
		t.Fatal("ini must not be created without a music folder")
	}
}

func TestBootDelayValidationAndDefaults(t *testing.T) {
	for _, value := range []string{"-1", "NaN", "Inf", "1e99", "9223372036.854776", "invalid"} {
		if _, err := ParseBootDelay(value); err == nil {
			t.Fatalf("accepted %q", value)
		}
		cfg := ConfigFromDocument(ParseINI("[bgm]\nbootdelay = " + value + "\n"))
		if cfg.BootDelay != 0 {
			t.Fatalf("invalid delay did not fall back: %v", cfg.BootDelay)
		}
	}
	cfg := ConfigFromDocument(ParseINI("[bgm]\nbootdelay = 1.25\ncorebootdelay = 3\n"))
	if cfg.BootDelay != 1.25 || cfg.CoreBootDelay != 3 {
		t.Fatalf("independent delays: %+v", cfg)
	}
	settings := SettingsFromConfig(&cfg)
	settings.BootDelay = -1
	if err := settings.Validate(); err == nil {
		t.Fatal("invalid staged startup delay accepted")
	}
	settings.BootDelay = 0.5
	settings.ApplyTo(&cfg)
	if cfg.BootDelay != 0.5 || cfg.CoreBootDelay != 3 {
		t.Fatal("startup delay changed core delay")
	}
}

func TestSaveSettingsPreservesUnknownEntriesAndFormat(t *testing.T) {
	paths := newTestPaths(t)
	writeINI(t, &paths, "[bgm]\nplayback = loop\ncustom = keep\nstartup = yes\n[extra]\nfoo = bar\n")
	settings := Settings{
		Theme: "nord", CoreBootDelay: 0.5, BootDelay: 1.25, MenuVolume: 5, DefaultVolume: -1,
		Startup: false, PlayInCore: true, Debug: true, Mouse: false, CRTMode: true, OnScreenKeyboard: true,
	}
	if err := SaveSettings(paths.IniFile, &settings); err != nil {
		t.Fatal(err)
	}
	want := "[bgm]\nplayback = loop\ncustom = keep\nstartup = no\nplayincore = yes\ncorebootdelay = 0.5\n" +
		"bootdelay = 1.25\nmenuvolume = 5\ndefaultvolume = -1\ndebug = yes\n\n[extra]\nfoo = bar\n\n" +
		"[tui]\ntheme = nord\nmouse = no\ncrt_mode = yes\non_screen_keyboard = yes\n\n"
	if got := readFile(t, paths.IniFile); got != want {
		t.Fatalf("saved:\n%q\nwant:\n%q", got, want)
	}
	cfg, err := LoadConfig(&paths)
	if err != nil {
		t.Fatal(err)
	}
	if SettingsFromConfig(&cfg) != settings {
		t.Fatalf("round trip %+v != %+v", SettingsFromConfig(&cfg), settings)
	}
}

func TestSavePlaybackAndPlaylist(t *testing.T) {
	paths := newTestPaths(t)
	writeINI(t, &paths, DefaultINI)
	if err := SavePlayback(paths.IniFile, PlaybackLoop); err != nil {
		t.Fatal(err)
	}
	if err := SavePlaylist(paths.IniFile, NamedPlaylist("chip tunes")); err != nil {
		t.Fatal(err)
	}
	cfg, _ := LoadConfig(&paths)
	if cfg.Playback != PlaybackLoop || cfg.Playlist.Name() != "chip tunes" {
		t.Fatalf("config %+v", cfg)
	}
	if err := SavePlaylist(paths.IniFile, Playlist{}); err != nil {
		t.Fatal(err)
	}
	if doc, _ := ReadINI(paths.IniFile); doc != nil {
		if value, _ := doc.Get("bgm", "playlist"); value != "none" {
			t.Fatalf("no playlist must be written as none, got %q", value)
		}
	}
}

func TestSettingsValidateAndDelayParsing(t *testing.T) {
	defaults := DefaultConfig()
	settings := SettingsFromConfig(&defaults)
	if err := settings.Validate(); err != nil {
		t.Fatal(err)
	}
	settings.MenuVolume = 8
	if err := settings.Validate(); err == nil {
		t.Fatal("volume 8 must be rejected")
	}
	if _, err := ParseDelay("-1"); err == nil {
		t.Fatal("negative delay must be rejected")
	}
	if value, err := ParseDelay(" 2.50 "); err != nil || value != 2.5 {
		t.Fatalf("ParseDelay = %v %v", value, err)
	}
	if FormatDelay(2.5) != "2.5" || FormatDelay(0) != "0" || FormatDelay(3) != "3" {
		t.Fatal("FormatDelay must not add trailing zeros")
	}
	both := Config{MenuVolume: 0, DefaultVolume: 7}
	menuOnly := Config{MenuVolume: -1, DefaultVolume: 7}
	if !both.ShouldChangeVolume() || menuOnly.ShouldChangeVolume() {
		t.Fatal("ShouldChangeVolume requires both volumes enabled")
	}
}
