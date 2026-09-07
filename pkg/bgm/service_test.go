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
	"context"
	"errors"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func newTestService(t *testing.T, paths *Paths) (*Service, *syncBuffer) {
	t.Helper()
	out := &syncBuffer{}
	logger := NewLoggerTo(paths, out)
	service := NewService(paths, logger)
	service.Watcher().RetryInterval = time.Millisecond
	service.Watcher().Settle = 20 * time.Millisecond
	service.Player().randIndex = func(int) int { return 0 }
	service.Player().sleep = func(time.Duration) {}
	return service, out
}

func runService(t *testing.T, service *Service) (cancel context.CancelFunc, errs <-chan error) {
	t.Helper()
	ctx, cancel := context.WithCancel(context.Background())
	results := make(chan error, 1)
	go func() { results <- service.Run(ctx) }()
	t.Cleanup(func() {
		cancel()
		service.Cleanup()
	})
	return cancel, results
}

func setCore(t *testing.T, paths *Paths, core string) {
	t.Helper()
	writeFile(t, paths.CoreNameFile, core)
}

func TestStartupDelayPrecedesBootAndOrdinaryAudio(t *testing.T) {
	for _, boot := range []bool{true, false} {
		name := "playlist"
		if boot {
			name = "boot"
		}
		t.Run(name, func(t *testing.T) {
			paths := newTestPaths(t)
			logPath := installStubPlayers(t)
			playback := "loop"
			track := "song.wav"
			if boot {
				playback = "disabled"
				track = "_boot.wav"
				t.Setenv("BGM_STUB_EXIT", "1")
			}
			writeINI(t, &paths, "[bgm]\nbootdelay = 0.1\nplayback = "+playback+"\n")
			touchTracks(t, paths.MusicFolder, track)
			setCore(t, &paths, MenuCore)
			service, _ := newTestService(t, &paths)
			started := time.Now()
			cancel, errs := runService(t, service)
			waitFor(t, func() bool { return len(stubInvocations(t, logPath)) > 0 })
			if time.Since(started) < 100*time.Millisecond {
				t.Fatal("audio started before delay")
			}
			cancel()
			if err := <-errs; !errors.Is(err, context.Canceled) {
				t.Fatalf("shutdown: %v", err)
			}
		})
	}
}

func TestStartupDelayCancelsWithoutPlaying(t *testing.T) {
	for _, cleanup := range []bool{false, true} {
		name := "context"
		if cleanup {
			name = "cleanup"
		}
		t.Run(name, func(t *testing.T) {
			paths := newTestPaths(t)
			logPath := installStubPlayers(t)
			writeINI(t, &paths, "[bgm]\nbootdelay = 3600\ndebug = yes\n")
			touchTracks(t, paths.MusicFolder, "_boot.wav")
			service, out := newTestService(t, &paths)
			cancel, errs := runService(t, service)
			waitFor(t, func() bool { return strings.Contains(out.String(), "before startup audio") })
			if cleanup {
				service.Cleanup()
			} else {
				cancel()
			}
			select {
			case err := <-errs:
				if !errors.Is(err, context.Canceled) {
					t.Fatalf("shutdown: %v", err)
				}
			case <-time.After(2 * time.Second):
				t.Fatal("shutdown waited for startup timer")
			}
			if len(stubInvocations(t, logPath)) != 0 {
				t.Fatal("audio started during canceled delay")
			}
		})
	}
}

func TestServiceFollowsCoreChanges(t *testing.T) {
	paths := newTestPaths(t)
	logPath := installStubPlayers(t)
	t.Setenv("BGM_STUB_EXIT", "1")
	writeINI(t, &paths, "[bgm]\nmenuvolume = 5\ndefaultvolume = 2\ndebug = yes\n")
	touchTracks(t, paths.MusicFolder, "_boot.wav", "song.wav")
	touchTracks(t, filepath.Join(paths.BootFolder, "SNES"), "snes.wav")
	setCore(t, &paths, MenuCore)

	service, out := newTestService(t, &paths)
	cancel, errs := runService(t, service)

	waitFor(t, func() bool { return SocketExists(&paths) })
	waitFor(t, func() bool { return strings.Contains(out.String(), "Starting random playlist...") })
	invocations := stubInvocations(t, logPath)
	if len(invocations) == 0 || !strings.HasSuffix(invocations[0], "_boot.wav") {
		t.Fatalf("boot track should play first: %v", invocations)
	}
	if got := readFile(t, paths.CmdInterface); got != "volume 5\n" {
		t.Fatalf("menu volume expected after boot, got %q", got)
	}
	if !strings.Contains(out.String(), "Setting volume to 2") {
		t.Fatal("default volume must be set after the boot track")
	}

	// Ensure the watch is established before triggering a change.
	time.Sleep(50 * time.Millisecond)
	setCore(t, &paths, "SNES")
	waitFor(t, func() bool { return strings.Contains(out.String(), "Playing core boot track...") })
	waitFor(t, func() bool { return strings.Contains(out.String(), "Releasing mutex") })
	if service.Player().InPlaylist() {
		t.Fatal("entering a core stops the playlist")
	}
	if got := readFile(t, paths.CmdInterface); got != "volume 2\n" {
		t.Fatalf("default volume expected in core, got %q", got)
	}
	if !strings.Contains(strings.Join(stubInvocations(t, logPath), "\n"), "snes.wav") {
		t.Fatal("core boot track did not play")
	}

	time.Sleep(50 * time.Millisecond)
	setCore(t, &paths, "snes")
	waitFor(t, func() bool { return strings.Contains(out.String(), "core is the same") })

	time.Sleep(50 * time.Millisecond)
	setCore(t, &paths, MenuCore)
	waitFor(t, func() bool { return strings.Contains(out.String(), "Switched to menu core") })
	waitFor(t, func() bool { return service.Player().InPlaylist() })
	waitFor(t, func() bool { return readFile(t, paths.CmdInterface) == "volume 5\n" })

	cancel()
	if err := <-errs; !errors.Is(err, context.Canceled) {
		t.Fatal(err)
	}
	service.Cleanup()
	if SocketExists(&paths) {
		t.Fatal("cleanup must remove the socket")
	}
	if !strings.Contains(out.String(), "Remote stopped") {
		t.Fatal("remote must be stopped")
	}
}

func TestServiceExitsWhenCoreNameNeverAppears(t *testing.T) {
	paths := newTestPaths(t)
	installStubPlayers(t)
	writeINI(t, &paths, "[bgm]\ndebug = yes\n")
	touchTracks(t, paths.MusicFolder, "song.wav")
	service, out := newTestService(t, &paths)
	service.Watcher().RetryCount = 2
	_, errs := runService(t, service)
	select {
	case err := <-errs:
		if err != nil {
			t.Fatal(err)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("service did not exit")
	}
	expected := []string{
		"CORENAME file does not exist, retrying...", "No CORENAME file found", "CORENAME file is missing, exiting...",
	}
	for _, want := range expected {
		if !strings.Contains(out.String(), want) {
			t.Fatalf("missing %q in %q", want, out.String())
		}
	}
	// Music still started because a missing CORENAME counts as the menu.
	if !strings.Contains(out.String(), "Starting random playlist...") {
		t.Fatal("playlist should start while CORENAME is missing")
	}
}

func TestServicePlayInCoreIgnoresCoreChanges(t *testing.T) {
	paths := newTestPaths(t)
	installStubPlayers(t)
	writeINI(t, &paths, "[bgm]\nplayincore = yes\ndebug = yes\n")
	touchTracks(t, paths.MusicFolder, "song.wav")
	setCore(t, &paths, "NES")
	service, out := newTestService(t, &paths)
	runService(t, service)
	waitFor(t, func() bool { return strings.Contains(out.String(), "Starting random playlist...") })
	time.Sleep(50 * time.Millisecond)
	setCore(t, &paths, "SNES")
	waitFor(t, func() bool { return strings.Contains(out.String(), "playincore is enabled") })
	if !service.Player().InPlaylist() {
		t.Fatal("music keeps playing inside cores")
	}
}

func TestServiceDoesNotStartInCore(t *testing.T) {
	paths := newTestPaths(t)
	installStubPlayers(t)
	writeINI(t, &paths, "[bgm]\ndebug = yes\n")
	touchTracks(t, paths.MusicFolder, "song.wav")
	setCore(t, &paths, "NES")
	service, out := newTestService(t, &paths)
	runService(t, service)
	waitFor(t, func() bool { return SocketExists(&paths) })
	time.Sleep(50 * time.Millisecond)
	if strings.Contains(out.String(), "Starting random playlist...") {
		t.Fatal("music must not start while a core is running")
	}
	if !service.Player().InPlaylist() {
		t.Fatal("in_playlist() stays true until a playlist is stopped")
	}
}

func TestVolumeSetClampsAndWritesCommand(t *testing.T) {
	paths := newTestPaths(t)
	logger, out := newTestLogger(&paths)
	writeINI(t, &paths, "[bgm]\ndebug = yes\n")
	VolumeSet(&paths, logger, 9)
	if got := readFile(t, paths.CmdInterface); got != "volume 7\n" {
		t.Fatalf("command %q", got)
	}
	VolumeSet(&paths, logger, -3)
	if got := readFile(t, paths.CmdInterface); got != "volume 0\n" {
		t.Fatalf("command %q", got)
	}
	if !strings.Contains(out.String(), "Setting volume to 7\nSetting volume to 0\n") {
		t.Fatalf("log %q", out.String())
	}
}

func TestGetCoreTrimsWhitespace(t *testing.T) {
	paths := newTestPaths(t)
	if _, ok := GetCore(&paths); ok {
		t.Fatal("missing file must report ok=false")
	}
	setCore(t, &paths, " MENU\n")
	if core, ok := GetCore(&paths); !ok || core != "MENU" {
		t.Fatalf("core %q %v", core, ok)
	}
}
