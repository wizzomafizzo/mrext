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
package main

import (
	"context"
	"net"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/wizzomafizzo/mrext/pkg/bgm"
)

type syncBuffer struct {
	buf strings.Builder
	mu  sync.Mutex
}

func (b *syncBuffer) Write(data []byte) (int, error) {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.buf.Write(data) //nolint:wrapcheck // test helper
}

func (b *syncBuffer) String() string {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.buf.String()
}

// newTestApp builds an app on a temporary root with recording spawn/TUI
// hooks and a short socket path.
func newTestApp(t *testing.T) (*app, *syncBuffer) {
	t.Helper()
	out := &syncBuffer{}
	application, err := newApp(cliOptions{root: t.TempDir()}, out)
	if err != nil {
		t.Fatal(err)
	}
	socketDir, err := os.MkdirTemp("", "bgm")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.RemoveAll(socketDir) })
	application.paths.SocketFile = filepath.Join(socketDir, bgm.SocketFilename)
	application.logger = bgm.NewLoggerTo(&application.paths, out)
	application.spawn = func() error { return nil }
	application.runTUI = func(*bgm.Config) error { return nil }
	return application, out
}

func writeTestFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

// startFakeService listens on the app socket with a stopped player so the
// CLI and TUI see a running service without spawning audio players.
func startFakeService(t *testing.T, application *app) *bgm.Player {
	t.Helper()
	cfg, _ := bgm.LoadConfig(&application.paths)
	player := bgm.NewPlayer(&application.paths, application.logger, &cfg)
	player.StopPlaylist()
	remote, err := bgm.StartRemote(&application.paths, application.logger, player)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		remote.Close()
		player.StopPlaylist()
		_ = os.Remove(application.paths.SocketFile)
	})
	return player
}

func TestParseCLI(t *testing.T) {
	for _, command := range []string{"exec", "start", "stop", "restart"} {
		options, err := parseCLI([]string{"--root", "/tmp/x", command})
		if err != nil || options.command != command || options.root != "/tmp/x" {
			t.Fatalf("parseCLI(%s) = %+v %v", command, options, err)
		}
	}
	options, err := parseCLI(nil)
	if err != nil || options.command != "" {
		t.Fatalf("no arguments should select the interactive path: %+v %v", options, err)
	}
	for _, args := range [][]string{{"status"}, {"start", "now"}, {"--bogus"}} {
		if _, err := parseCLI(args); err == nil {
			t.Fatalf("expected an error for %v", args)
		}
	}
}

func TestStartHonoursStartupFlag(t *testing.T) {
	application, out := newTestApp(t)
	spawned := 0
	application.spawn = func() error { spawned++; return nil }
	writeTestFile(t, application.paths.IniFile, "[bgm]\nstartup = no\n")
	if err := application.dispatch(commandStart); err != nil {
		t.Fatal(err)
	}
	if spawned != 0 || out.String() != "Auto-start is disabled in configuration\n" {
		t.Fatalf("spawned=%d output %q", spawned, out.String())
	}
	writeTestFile(t, application.paths.IniFile, bgm.DefaultINI)
	if err := application.dispatch(commandStart); err != nil {
		t.Fatal(err)
	}
	if spawned != 1 {
		t.Fatalf("spawned=%d", spawned)
	}
}

func TestStopAndRestartWithoutService(t *testing.T) {
	application, out := newTestApp(t)
	spawned := 0
	application.spawn = func() error { spawned++; return nil }
	writeTestFile(t, application.paths.IniFile, bgm.DefaultINI)
	if err := application.dispatch(commandStop); err != nil {
		t.Fatal(err)
	}
	if out.String() != "BGM service is not running\n" {
		t.Fatalf("output %q", out.String())
	}
	if err := application.dispatch(commandRestart); err != nil {
		t.Fatal(err)
	}
	if spawned != 1 {
		t.Fatalf("restart should start the service, spawned=%d", spawned)
	}
}

func TestInteractiveFirstRunPreparesFilesAndAsksForMusic(t *testing.T) {
	application, out := newTestApp(t)
	application.paths.AppPath = "/media/fat/Scripts/bgm.sh"
	spawned := false
	application.spawn = func() error { spawned = true; return nil }
	if err := application.dispatch(""); err != nil {
		t.Fatal(err)
	}
	want := "Created music folder.\nAdded service to startup script.\n" +
		"Add music files to " + application.paths.MusicFolder + " and re-run this script to start.\n"
	if out.String() != want {
		t.Fatalf("output %q", out.String())
	}
	if spawned {
		t.Fatal("service must not start without music")
	}
	data, err := os.ReadFile(application.paths.IniFile)
	if err != nil || string(data) != bgm.DefaultINI {
		t.Fatalf("default ini not created: %q %v", data, err)
	}
	startup, err := os.ReadFile(application.paths.StartupScript)
	hook := "[[ -e /media/fat/Scripts/bgm.sh ]] && /media/fat/Scripts/bgm.sh $1"
	if err != nil || !strings.Contains(string(startup), hook) {
		t.Fatalf("startup script %q %v", startup, err)
	}
}

func TestInteractiveStartsServiceThenOpensTUI(t *testing.T) {
	application, out := newTestApp(t)
	writeTestFile(t, filepath.Join(application.paths.MusicFolder, "song.mp3"), "")
	opened := 0
	application.runTUI = func(*bgm.Config) error { opened++; return nil }
	spawned := 0
	application.spawn = func() error {
		spawned++
		startFakeService(t, application)
		return nil
	}
	if err := application.dispatch(""); err != nil {
		t.Fatal(err)
	}
	if spawned != 1 || opened != 1 {
		t.Fatalf("spawned=%d opened=%d", spawned, opened)
	}
	if !strings.Contains(out.String(), "Starting BGM service...\n") {
		t.Fatalf("output %q", out.String())
	}
	// With the service already running the TUI opens directly.
	if err := application.dispatch(""); err != nil {
		t.Fatal(err)
	}
	if spawned != 1 || opened != 2 {
		t.Fatalf("spawned=%d opened=%d", spawned, opened)
	}
}

func TestInteractiveReplacesStaleSocket(t *testing.T) {
	application, out := newTestApp(t)
	writeTestFile(t, filepath.Join(application.paths.MusicFolder, "song.mp3"), "")
	listener, err := (&net.ListenConfig{}).Listen(context.Background(), "unix", application.paths.SocketFile)
	if err != nil {
		t.Fatal(err)
	}
	if unixListener, ok := listener.(*net.UnixListener); ok {
		unixListener.SetUnlinkOnClose(false)
	}
	_ = listener.Close()
	spawned, opened := 0, 0
	application.spawn = func() error {
		spawned++
		startFakeService(t, application)
		return nil
	}
	application.runTUI = func(*bgm.Config) error { opened++; return nil }
	if err := application.dispatch(""); err != nil {
		t.Fatal(err)
	}
	if spawned != 1 || opened != 1 || !strings.Contains(out.String(), "Starting BGM service...\n") {
		t.Fatalf("spawned=%d opened=%d output %q", spawned, opened, out.String())
	}
}

func TestInteractiveGivesUpWhenServiceNeverAppears(t *testing.T) {
	application, out := newTestApp(t)
	previous := serviceTimeout
	serviceTimeout = 100 * time.Millisecond
	t.Cleanup(func() { serviceTimeout = previous })
	writeTestFile(t, filepath.Join(application.paths.MusicFolder, "song.mp3"), "")
	opened := false
	application.runTUI = func(*bgm.Config) error { opened = true; return nil }
	if err := application.dispatch(""); err != nil {
		t.Fatal(err)
	}
	if opened || !strings.HasSuffix(out.String(), "Starting BGM service...\n") {
		t.Fatalf("opened=%v output %q", opened, out.String())
	}
}

func TestExecRefusesWhenServiceRunning(t *testing.T) {
	application, out := newTestApp(t)
	writeTestFile(t, application.paths.IniFile, bgm.DefaultINI)
	startFakeService(t, application)
	if err := application.dispatch(commandExec); err != nil {
		t.Fatal(err)
	}
	if out.String() != "BGM service is already running, exiting...\n" {
		t.Fatalf("output %q", out.String())
	}
}
