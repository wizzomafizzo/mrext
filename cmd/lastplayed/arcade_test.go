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

//nolint:gosec // Tests operate only on temporary fixture roots.
package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/wizzomafizzo/mrext/pkg/config"
	"github.com/wizzomafizzo/mrext/pkg/tracker"
)

func TestArcadeEventCreatesLastAndRecentLinks(t *testing.T) {
	root := t.TempDir()
	folder := filepath.Join(root, "_Arcade", "nested")
	if err := os.MkdirAll(folder, 0o755); err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(folder, "Pooyan.mra")
	if err := os.WriteFile(path, []byte("<misterromdescription/>"), 0o600); err != nil {
		t.Fatal(err)
	}
	db := &fakeDb{config: &config.UserConfig{}, sdRoot: root}
	event := &tracker.EventAction{Action: tracker.EventActionGameStart, Target: "pooyan", TargetPath: path}
	for range 2 {
		if err := db.AddEvent(event); err != nil {
			t.Fatal(err)
		}
	}
	for _, link := range []string{
		filepath.Join(root, "Last Played.mra"),
		filepath.Join(root, "_Recently Played", "01 Pooyan [Arcade].mra"),
	} {
		target, err := os.Readlink(link)
		if err != nil || target != path {
			t.Fatalf("link=%s target=%s err=%v", link, target, err)
		}
	}
	entries, err := os.ReadDir(filepath.Join(root, "_Recently Played"))
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 2 { // One MRA and the shared cores link.
		t.Fatalf("duplicate recent entries: %v", entries)
	}
	for _, unresolved := range []string{"", "   "} {
		event.TargetPath = unresolved
		if eventErr := db.AddEvent(event); eventErr != nil {
			t.Fatal(eventErr)
		}
	}
	event.TargetPath = filepath.Join(root, "missing.mra")
	if eventErr := db.AddEvent(event); eventErr == nil {
		t.Fatal("missing MRA was accepted")
	}
	target, err := os.Readlink(filepath.Join(root, "Last Played.mra"))
	if err != nil || target != path {
		t.Fatalf("failed resolution replaced valid launcher: %q %v", target, err)
	}
}

func TestRecentFolderKeepsMixedGamesAcrossReplays(t *testing.T) {
	root := t.TempDir()
	cfg := &config.UserConfig{}
	cfg.Systems.GamesFolder = []string{root}
	arcade := filepath.Join(root, "_Arcade", "Pooyan.mra")
	console := filepath.Join(root, "games", "NES", "Mario.nes")
	for _, path := range []string{arcade, console} {
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, nil, 0o600); err != nil {
			t.Fatal(err)
		}
	}
	for _, path := range []string{arcade, console, arcade, console} {
		if err := addToRecentFolder(cfg, path, root); err != nil {
			t.Fatal(err)
		}
	}
	recent := filepath.Join(root, "_Recently Played")
	entries, err := os.ReadDir(recent)
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 3 {
		t.Fatalf("expected two launchers and cores link: %v", entries)
	}
	for _, entry := range entries {
		path := filepath.Join(recent, entry.Name())
		switch filepath.Ext(path) {
		case ".mra":
			target, readErr := os.Readlink(path)
			if readErr != nil || target != arcade {
				t.Fatalf("broken arcade link: %q %v", target, readErr)
			}
		case ".mgl":
			data, readErr := os.ReadFile(path)
			if readErr != nil || !strings.Contains(string(data), console) {
				t.Fatalf("broken console launcher: %s %v", data, readErr)
			}
		}
	}
}

func TestEmptyTargetCannotCreateLaunchers(t *testing.T) {
	root := t.TempDir()
	cfg := &config.UserConfig{}
	if err := createLastPlayedMgl(cfg, "", root); err == nil {
		t.Fatal("empty last-played target accepted")
	}
	if err := addToRecentFolder(cfg, "", root); err == nil {
		t.Fatal("empty recent target accepted")
	}
	entries, err := os.ReadDir(root)
	if err != nil || len(entries) != 0 {
		t.Fatalf("empty target mutated menu: %v %v", entries, err)
	}
}
