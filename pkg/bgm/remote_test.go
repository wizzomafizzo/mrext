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
	"strconv"
	"strings"
	"testing"
)

func startTestRemote(t *testing.T, paths *Paths, cfg *Config) (*Remote, *Player, *syncBuffer) {
	t.Helper()
	out := &syncBuffer{}
	logger := NewLoggerTo(paths, out)
	player := NewPlayer(paths, logger, cfg)
	player.randIndex = func(int) int { return 0 }
	remote, err := StartRemote(paths, logger, player)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		remote.Close()
		player.StopPlaylist()
	})
	return remote, player, out
}

func TestSendWithoutSocket(t *testing.T) {
	paths := newTestPaths(t)
	reply, replied, err := Send(&paths, "status")
	if err != nil || replied || reply != "" {
		t.Fatalf("Send = %q %v %v", reply, replied, err)
	}
	if SocketExists(&paths) || SocketStale(&paths) {
		t.Fatal("no socket file expected")
	}
}

func TestRemoteProtocol(t *testing.T) {
	paths := newTestPaths(t)
	installStubPlayers(t)
	t.Setenv("BGM_STUB_EXIT", "1")
	writeINI(t, &paths, "[bgm]\ndebug = yes\n")
	touchTracks(t, filepath.Join(paths.MusicFolder, "my music"), "a.wav")
	_, player, out := startTestRemote(t, &paths, ptr(DefaultConfig()))
	player.StopPlaylist()

	reply, replied, err := Send(&paths, "status")
	if err != nil || !replied || reply != "no\trandom\tnone\t" {
		t.Fatalf("status = %q %v %v", reply, replied, err)
	}
	reply, replied, _ = Send(&paths, "pid")
	if !replied || reply != strconv.Itoa(os.Getpid()) {
		t.Fatalf("pid = %q", reply)
	}
	if _, replied, _ = Send(&paths, "get playlist"); replied {
		t.Fatal("get playlist with no playlist must send nothing")
	}
	if reply, _, _ = Send(&paths, "get playback"); reply != "random" {
		t.Fatalf("get playback = %q", reply)
	}
	if reply, _, _ = Send(&paths, "get playincore"); reply != "no" {
		t.Fatalf("get playincore = %q", reply)
	}
	if _, replied, _ = Send(&paths, "get bogus"); replied {
		t.Fatal("unknown get sends an empty reply")
	}
	if _, replied, _ = Send(&paths, "set playincore yes"); replied {
		t.Fatal("set commands send nothing")
	}
	if reply, _, _ = Send(&paths, "get playincore"); reply != "yes" {
		t.Fatalf("get playincore = %q", reply)
	}
	_, _, _ = Send(&paths, "set playlist my music")
	if reply, _, _ = Send(&paths, "get playlist"); reply != "my music" {
		t.Fatalf("playlist names keep their spaces, got %q", reply)
	}
	_, _, _ = Send(&paths, "set playback shuffle")
	if reply, _, _ = Send(&paths, "status"); reply != "no\tshuffle\tmy music\t" {
		t.Fatalf("status = %q", reply)
	}
	if strings.Contains(out.String(), "Starting random playlist") {
		t.Fatal("set playback must not start music while stopped")
	}
	_, _, _ = Send(&paths, "play")
	if !player.InPlaylist() {
		t.Fatal("play starts the playlist")
	}
	_, _, _ = Send(&paths, "stop")
	if player.InPlaylist() {
		t.Fatal("stop ends the playlist")
	}
	_, _, _ = Send(&paths, "dance")
	if !strings.Contains(out.String(), "Unknown command: dance") {
		t.Fatalf("log %q", out.String())
	}
	if SocketStale(&paths) {
		t.Fatal("a live service must not look stale")
	}
	if strings.Contains(out.String(), "Unknown command: \n") {
		t.Fatal("liveness probes must not be logged as commands")
	}
}

func TestRemoteQuitStopsListeningButKeepsSocketFile(t *testing.T) {
	paths := newTestPaths(t)
	writeINI(t, &paths, "[bgm]\ndebug = yes\n")
	remote, _, out := startTestRemote(t, &paths, ptr(DefaultConfig()))
	if _, replied, err := Send(&paths, "quit"); replied || err != nil {
		t.Fatalf("quit reply %v %v", replied, err)
	}
	<-remote.Done()
	if !SocketExists(&paths) {
		t.Fatal("quit must leave the socket file behind like Python")
	}
	if !SocketStale(&paths) {
		t.Fatal("nobody listens after quit")
	}
	if !strings.Contains(out.String(), "Remote stopped") {
		t.Fatalf("log %q", out.String())
	}
	remote.Close()
}

func TestRemoteRefusesSecondBind(t *testing.T) {
	paths := newTestPaths(t)
	startTestRemote(t, &paths, ptr(DefaultConfig()))
	logger, _ := newTestLogger(&paths)
	defaults := DefaultConfig()
	if _, err := StartRemote(&paths, logger, NewPlayer(&paths, logger, &defaults)); err == nil {
		t.Fatal("binding an existing socket must fail")
	}
}
