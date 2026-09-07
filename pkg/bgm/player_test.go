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
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"
	"time"
)

func newTestPlayer(t *testing.T, paths *Paths, cfg *Config) *Player {
	t.Helper()
	logger, _ := newTestLogger(paths)
	player := NewPlayer(paths, logger, cfg)
	player.randIndex = func(int) int { return 0 }
	player.sleep = func(time.Duration) {}
	player.radioRetryDelay = 0
	t.Cleanup(func() { player.radioClient.CloseIdleConnections() })
	t.Cleanup(player.StopPlaylist)
	return player
}

func touchTracks(t *testing.T, folder string, names ...string) {
	t.Helper()
	for _, name := range names {
		writeFile(t, filepath.Join(folder, name), "")
	}
}

func relativeTracks(t *testing.T, paths *Paths, tracks []string) []string {
	t.Helper()
	relative := make([]string, 0, len(tracks))
	for _, track := range tracks {
		rel, err := filepath.Rel(paths.MusicFolder, track)
		if err != nil {
			t.Fatal(err)
		}
		relative = append(relative, filepath.ToSlash(rel))
	}
	slices.Sort(relative)
	return relative
}

func TestPlaylistPathFallsBackToMusicFolder(t *testing.T) {
	paths := newTestPaths(t)
	player := newTestPlayer(t, &paths, ptr(DefaultConfig()))
	if err := os.MkdirAll(filepath.Join(paths.MusicFolder, "chip"), 0o755); err != nil {
		t.Fatal(err)
	}
	if got := player.PlaylistPath(Playlist{}); got != paths.MusicFolder {
		t.Fatalf("none -> %s", got)
	}
	if got := player.PlaylistPath(NamedPlaylist("all")); got != paths.MusicFolder {
		t.Fatalf("all -> %s", got)
	}
	if got := player.PlaylistPath(NamedPlaylist("chip")); got != filepath.Join(paths.MusicFolder, "chip") {
		t.Fatalf("chip -> %s", got)
	}
	if got := player.PlaylistPath(NamedPlaylist("missing")); got != paths.MusicFolder {
		t.Fatalf("missing -> %s", got)
	}
}

func layoutMusic(t *testing.T, paths *Paths) {
	t.Helper()
	touchTracks(t, paths.MusicFolder, "top.mp3", "_boot.mp3", "radio.pls", "notes.txt")
	touchTracks(t, filepath.Join(paths.MusicFolder, "chip"), "one.ogg", "_intro.wav", "stream.pls")
	touchTracks(t, filepath.Join(paths.MusicFolder, "chip", "deep"), "two.vgm")
	touchTracks(t, paths.BootFolder, "global.wav")
	touchTracks(t, filepath.Join(paths.BootFolder, "SNES"), "snes.mp3")
	if err := os.MkdirAll(filepath.Join(paths.MusicFolder, "folder.mp3"), 0o755); err != nil {
		t.Fatal(err)
	}
}

func TestBootRotationPreservesSelectedPlaylist(t *testing.T) {
	paths := newTestPaths(t)
	layoutMusic(t, &paths)
	cfg := DefaultConfig()
	cfg.Playlist = NamedPlaylist("chip")
	player := newTestPlayer(t, &paths, &cfg)
	before := player.Tracks(cfg.Playlist, false)
	player.SetBootInPlaylist(true)
	got := relativeTracks(t, &paths, player.Tracks(cfg.Playlist, false))
	for _, name := range []string{"chip/_intro.wav", "chip/one.ogg", "chip/deep/two.vgm", "boot/global.wav"} {
		if !slices.Contains(got, name) {
			t.Fatalf("missing %s in %v", name, got)
		}
	}
	for _, name := range []string{"_boot.mp3", "top.mp3", "boot/SNES/snes.mp3"} {
		if slices.Contains(got, name) {
			t.Fatalf("unrelated track %s in %v", name, got)
		}
	}
	if player.Playlist() != cfg.Playlist {
		t.Fatal("selected playlist changed")
	}
	player.SetBootInPlaylist(false)
	if !slices.Equal(before, player.Tracks(cfg.Playlist, false)) {
		t.Fatal("disabling did not restore tracks")
	}
	player.SetBootInPlaylist(true)
	player.playlist = NamedPlaylist("all")
	all := player.Tracks(player.Playlist(), false)
	count := 0
	for _, name := range all {
		if name == filepath.Join(paths.BootFolder, "global.wav") {
			count++
		}
		if IsPLS(name) {
			t.Fatalf("all unexpectedly includes radio: %s", name)
		}
	}
	if count != 1 {
		t.Fatalf("global boot track appears %d times", count)
	}
}

func TestBootRotationAvoidsImmediateRepeat(t *testing.T) {
	paths := newTestPaths(t)
	touchTracks(t, paths.MusicFolder, "_boot.wav", "song.wav")
	cfg := DefaultConfig()
	cfg.BootInPlaylist = true
	player := newTestPlayer(t, &paths, &cfg)
	player.randIndex = func(int) int { return 0 }
	boot := filepath.Join(paths.MusicFolder, "_boot.wav")
	song := filepath.Join(paths.MusicFolder, "song.wav")
	player.addHistory(boot)
	for range 4 {
		if got, ok := player.randomTrack(); !ok || got != song {
			t.Fatalf("after boot: %q %v", got, ok)
		}
		player.addHistory(song)
		if got, ok := player.randomTrack(); !ok || got != boot {
			t.Fatalf("after song: %q %v", got, ok)
		}
		player.addHistory(boot)
	}
	if err := os.Remove(song); err != nil {
		t.Fatal(err)
	}
	if got, ok := player.randomTrack(); !ok || got != boot {
		t.Fatalf("single track: %q %v", got, ok)
	}
}

func TestTracksNoneIsTopLevelOnly(t *testing.T) {
	paths := newTestPaths(t)
	layoutMusic(t, &paths)
	player := newTestPlayer(t, &paths, ptr(DefaultConfig()))
	got := relativeTracks(t, &paths, player.Tracks(Playlist{}, false))
	want := []string{"folder.mp3", "radio.pls", "top.mp3"}
	if !slices.Equal(got, want) {
		t.Fatalf("tracks %v want %v", got, want)
	}
	withBoot := relativeTracks(t, &paths, player.Tracks(Playlist{}, true))
	if !slices.Contains(withBoot, "_boot.mp3") || len(withBoot) != 4 {
		t.Fatalf("include_boot tracks %v", withBoot)
	}
}

func TestTracksAllIsRecursiveWithoutPLS(t *testing.T) {
	paths := newTestPaths(t)
	layoutMusic(t, &paths)
	cfg := DefaultConfig()
	cfg.Playlist = NamedPlaylist("all")
	player := newTestPlayer(t, &paths, &cfg)
	got := relativeTracks(t, &paths, player.Tracks(player.Playlist(), false))
	want := []string{"boot/SNES/snes.mp3", "boot/global.wav", "chip/deep/two.vgm", "chip/one.ogg", "top.mp3"}
	if !slices.Equal(got, want) {
		t.Fatalf("tracks %v want %v", got, want)
	}
}

func TestTracksNamedPlaylistKeepsPLSAndUsesCurrentPlaylistRule(t *testing.T) {
	paths := newTestPaths(t)
	layoutMusic(t, &paths)
	cfg := DefaultConfig()
	cfg.Playlist = NamedPlaylist("chip")
	player := newTestPlayer(t, &paths, &cfg)
	got := relativeTracks(t, &paths, player.Tracks(player.Playlist(), false))
	want := []string{"chip/deep/two.vgm", "chip/one.ogg", "chip/stream.pls"}
	if !slices.Equal(got, want) {
		t.Fatalf("tracks %v want %v", got, want)
	}
	// A non-existent playlist falls back to the music folder recursively and
	// still keeps .pls files because the current playlist is not "all".
	missing := relativeTracks(t, &paths, player.Tracks(NamedPlaylist("missing"), false))
	if !slices.Contains(missing, "radio.pls") || !slices.Contains(missing, "boot/global.wav") {
		t.Fatalf("missing playlist tracks %v", missing)
	}
	// The .pls rule follows the *current* playlist, not the requested one.
	player.playlist = NamedPlaylist("all")
	chipUnderAll := relativeTracks(t, &paths, player.Tracks(NamedPlaylist("chip"), false))
	if slices.Contains(chipUnderAll, "chip/stream.pls") {
		t.Fatalf("pls should be excluded while current playlist is all: %v", chipUnderAll)
	}
}

func TestHistoryTrimsToOneMoreThanTwentyPercent(t *testing.T) {
	paths := newTestPaths(t)
	player := newTestPlayer(t, &paths, ptr(DefaultConfig()))
	touchTracks(t, paths.MusicFolder, "a.mp3", "b.mp3", "c.mp3", "d.mp3")
	player.addHistory("a.mp3")
	if len(player.History()) != 0 {
		t.Fatal("fewer than five tracks must keep no history")
	}
	for _, name := range []string{"e", "f", "g", "h", "i", "j"} {
		touchTracks(t, paths.MusicFolder, name+".mp3")
	}
	for _, name := range []string{"1", "2", "3", "4", "5"} {
		player.addHistory(name)
	}
	if got := player.History(); !slices.Equal(got, []string{"3", "4", "5"}) {
		t.Fatalf("history %v", got)
	}
}

func TestRandomTrackAvoidsHistoryAndFallsBackWhenExhausted(t *testing.T) {
	paths := newTestPaths(t)
	player := newTestPlayer(t, &paths, ptr(DefaultConfig()))
	touchTracks(t, paths.MusicFolder, "a.mp3", "b.mp3")
	player.history = []string{filepath.Join(paths.MusicFolder, "a.mp3")}
	track, ok := player.randomTrack()
	if !ok || filepath.Base(track) != "b.mp3" {
		t.Fatalf("expected b.mp3, got %q %v", track, ok)
	}
	player.history = append(player.history, filepath.Join(paths.MusicFolder, "b.mp3"))
	if _, ok := player.randomTrack(); !ok {
		t.Fatal("exhausted history must still return a track")
	}
	_ = os.Remove(filepath.Join(paths.MusicFolder, "a.mp3"))
	_ = os.Remove(filepath.Join(paths.MusicFolder, "b.mp3"))
	if _, ok := player.randomTrack(); ok {
		t.Fatal("no tracks must return ok=false")
	}
}

func TestPlayLoopsAndKillsOnFinished(t *testing.T) {
	paths := newTestPaths(t)
	logPath := installStubPlayers(t)
	t.Setenv("BGM_STUB_FINISH", "1")
	player := newTestPlayer(t, &paths, ptr(DefaultConfig()))
	track := filepath.Join(paths.MusicFolder, "X03_song.mp3")
	touchTracks(t, paths.MusicFolder, "X03_song.mp3")

	start := time.Now()
	player.Play(track)
	if time.Since(start) > 5*time.Second {
		t.Fatal("mpg123 was not killed after reporting finished")
	}
	invocations := stubInvocations(t, logPath)
	if len(invocations) != 3 {
		t.Fatalf("expected three loops, got %v", invocations)
	}
	if invocations[0] != "mpg123 --no-control "+track {
		t.Fatalf("argv %q", invocations[0])
	}
	if _, playing := player.Playing(); playing {
		t.Fatal("playing must be cleared after the track ends")
	}
}

func TestPlayX00PlaysNothingButRecordsHistory(t *testing.T) {
	paths := newTestPaths(t)
	logPath := installStubPlayers(t)
	player := newTestPlayer(t, &paths, ptr(DefaultConfig()))
	for _, name := range []string{"X00_a.mp3", "b.mp3", "c.mp3", "d.mp3", "e.mp3"} {
		touchTracks(t, paths.MusicFolder, name)
	}
	track := filepath.Join(paths.MusicFolder, "X00_a.mp3")
	player.Play(track)
	if len(stubInvocations(t, logPath)) != 0 {
		t.Fatal("X00_ must not start a player")
	}
	if !slices.Contains(player.History(), track) {
		t.Fatal("X00_ tracks are still added to the history")
	}
	player.Play(filepath.Join(paths.MusicFolder, "notes.txt"))
	if len(stubInvocations(t, logPath)) != 0 {
		t.Fatal("invalid files must be ignored")
	}
}

func TestPlayerCommandsAndArguments(t *testing.T) {
	paths := newTestPaths(t)
	logPath := installStubPlayers(t)
	t.Setenv("BGM_STUB_EXIT", "1")
	player := newTestPlayer(t, &paths, ptr(DefaultConfig()))
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "audio/mpeg")
		_, _ = w.Write([]byte("test audio"))
	}))
	t.Cleanup(server.Close)
	for _, name := range []string{"a.ogg", "b.wav", "c.mid", "d.vgz", "radio.pls"} {
		touchTracks(t, paths.MusicFolder, name)
	}
	writeFile(t, filepath.Join(paths.MusicFolder, "radio.pls"), "File1="+server.URL+"\n")
	for _, name := range []string{"a.ogg", "b.wav", "c.mid", "d.vgz", "radio.pls"} {
		player.Play(filepath.Join(paths.MusicFolder, name))
	}
	got := stubInvocations(t, logPath)
	want := []string{
		"ogg123 " + filepath.Join(paths.MusicFolder, "a.ogg"),
		"aplay " + filepath.Join(paths.MusicFolder, "b.wav"),
		"aplaymidi " + filepath.Join(paths.MusicFolder, "c.mid") + " --port=128:0",
		"vgmplay " + filepath.Join(paths.MusicFolder, "d.vgz"),
		"mpg123 --no-control -",
	}
	if !slices.Equal(got, want) {
		t.Fatalf("invocations %v want %v", got, want)
	}
}

func TestStopKillsRunningPlayerAndAbandonsLoops(t *testing.T) {
	paths := newTestPaths(t)
	logPath := installStubPlayers(t)
	player := newTestPlayer(t, &paths, ptr(DefaultConfig()))
	track := filepath.Join(paths.MusicFolder, "X05_long.wav")
	touchTracks(t, paths.MusicFolder, "X05_long.wav")

	done := make(chan struct{})
	go func() {
		defer close(done)
		player.Play(track)
	}()
	waitFor(t, func() bool { return len(stubInvocations(t, logPath)) == 1 })
	waitFor(t, func() bool { _, playing := player.Playing(); return playing })
	if status := player.Status(); status != "yes\trandom\tnone\tX05_long.wav" {
		t.Fatalf("status %q", status)
	}
	player.Stop()
	select {
	case <-done:
	case <-time.After(5 * time.Second):
		t.Fatal("Play did not return after Stop")
	}
	if len(stubInvocations(t, logPath)) != 1 {
		t.Fatal("remaining loops must be abandoned after Stop")
	}
	if status := player.Status(); status != "no\trandom\tnone\t" {
		t.Fatalf("status %q", status)
	}
	player.Stop()
}

func TestStartPlaylistSemantics(t *testing.T) {
	paths := newTestPaths(t)
	installStubPlayers(t)
	t.Setenv("BGM_STUB_EXIT", "1")
	player := newTestPlayer(t, &paths, ptr(DefaultConfig()))
	touchTracks(t, paths.MusicFolder, "a.wav")
	if !player.InPlaylist() {
		t.Fatal("in_playlist() is true before any playlist starts")
	}
	player.StartPlaylist(PlaybackDisabled)
	if player.InPlaylist() || player.Playback() != PlaybackDisabled {
		t.Fatal("disabled must stop and store the playback type")
	}
	player.StartPlaylist("shuffle")
	if player.Playback() != PlaybackDisabled || !player.InPlaylist() {
		t.Fatal("unknown playback plays random without changing the stored value")
	}
	player.StopPlaylist()
	player.StartPlaylist(PlaybackLoop)
	if player.Playback() != PlaybackLoop {
		t.Fatal("loop must be stored")
	}
	player.StopPlaylist()
	if player.InPlaylist() {
		t.Fatal("stopping ends the playlist")
	}
}

func TestLoopPlaylistRepeatsTheSameTrack(t *testing.T) {
	paths := newTestPaths(t)
	logPath := installStubPlayers(t)
	t.Setenv("BGM_STUB_EXIT", "1")
	player := newTestPlayer(t, &paths, ptr(DefaultConfig()))
	touchTracks(t, paths.MusicFolder, "a.wav", "b.wav")
	calls := 0
	player.randIndex = func(n int) int {
		calls++
		return (calls * 7) % n
	}
	player.StartPlaylist(PlaybackLoop)
	waitFor(t, func() bool { return len(stubInvocations(t, logPath)) >= 3 })
	player.StopPlaylist()
	invocations := stubInvocations(t, logPath)
	for _, invocation := range invocations[1:] {
		if invocation != invocations[0] {
			t.Fatalf("loop playlist changed track: %v", invocations)
		}
	}
}

func TestRandomPlaylistEndsWithoutTracks(t *testing.T) {
	paths := newTestPaths(t)
	installStubPlayers(t)
	logger, out := newTestLogger(&paths)
	writeINI(t, &paths, "[bgm]\ndebug = yes\n")
	defaults := DefaultConfig()
	player := NewPlayer(&paths, logger, &defaults)
	player.StartPlaylist(PlaybackRandom)
	waitFor(t, func() bool { return strings.Contains(out.String(), "Random playlist ended") })
	player.StopPlaylist()
}

func TestChangePlaylistQuirks(t *testing.T) {
	paths := newTestPaths(t)
	installStubPlayers(t)
	logger, out := newTestLogger(&paths)
	writeINI(t, &paths, "[bgm]\ndebug = yes\n")
	cfg := DefaultConfig()
	cfg.Playlist = NamedPlaylist("chip")
	player := NewPlayer(&paths, logger, &cfg)
	player.randIndex = func(int) int { return 0 }
	t.Cleanup(player.StopPlaylist)
	touchTracks(t, filepath.Join(paths.MusicFolder, "chip"), "one.ogg")
	if err := os.MkdirAll(filepath.Join(paths.MusicFolder, "empty"), 0o755); err != nil {
		t.Fatal(err)
	}
	player.StopPlaylist()

	player.ChangePlaylist("empty")
	if player.Playlist().Name() != "chip" {
		t.Fatal("an empty folder must be rejected")
	}
	// "none" counts the *current* playlist, so it is accepted although the
	// top level has no files.
	player.ChangePlaylist("none")
	if !player.Playlist().IsNone() {
		t.Fatal("none must be accepted when the current playlist has tracks")
	}
	if !strings.Contains(out.String(), "Changed playlist: None") {
		t.Fatalf("log %q", out.String())
	}
	player.ChangePlaylist("chip")
	// Missing folders fall back to the music folder and are accepted.
	player.ChangePlaylist("does not exist")
	if player.Playlist().Name() != "does not exist" {
		t.Fatalf("playlist %q", player.Playlist().Name())
	}
	if strings.Contains(out.String(), "Starting random playlist") {
		t.Fatal("changing a playlist while stopped must not start music")
	}
}

func TestChangePlaylistRestartsWhenPlaying(t *testing.T) {
	paths := newTestPaths(t)
	installStubPlayers(t)
	t.Setenv("BGM_STUB_EXIT", "1")
	logger, out := newTestLogger(&paths)
	writeINI(t, &paths, "[bgm]\ndebug = yes\n")
	defaults := DefaultConfig()
	player := NewPlayer(&paths, logger, &defaults)
	player.randIndex = func(int) int { return 0 }
	t.Cleanup(player.StopPlaylist)
	touchTracks(t, paths.MusicFolder, "a.wav")
	touchTracks(t, filepath.Join(paths.MusicFolder, "chip"), "one.wav")
	player.StartPlaylist(PlaybackRandom)
	player.ChangePlaylist("chip")
	if strings.Count(out.String(), "Starting random playlist...") != 2 {
		t.Fatalf("expected a restart, log %q", out.String())
	}
	player.StopPlaylist()
}

func TestBootTrackSources(t *testing.T) {
	paths := newTestPaths(t)
	layoutMusic(t, &paths)
	cfg := DefaultConfig()
	cfg.Playlist = NamedPlaylist("chip")
	player := newTestPlayer(t, &paths, &cfg)
	seen := map[string]bool{}
	for index := range 2 {
		player.randIndex = func(int) int { return index }
		track, ok := player.BootTrack()
		if !ok {
			t.Fatal("expected a boot track")
		}
		seen[filepath.Base(track)] = true
	}
	if !seen["_intro.wav"] || !seen["global.wav"] {
		t.Fatalf("boot candidates %v", seen)
	}
	player.playlist = NamedPlaylist("all")
	player.randIndex = func(int) int { return 0 }
	if track, _ := player.BootTrack(); filepath.Base(track) != "_boot.mp3" {
		t.Fatalf("all should use top-level underscore files, got %q", track)
	}
	if err := os.RemoveAll(paths.BootFolder); err != nil {
		t.Fatal(err)
	}
	player.playlist = NamedPlaylist("chip")
	touchTracks(t, filepath.Join(paths.MusicFolder, "chip"), "_only.wav")
	_ = os.Remove(filepath.Join(paths.MusicFolder, "chip", "_intro.wav"))
	if track, ok := player.BootTrack(); !ok || filepath.Base(track) != "_only.wav" {
		t.Fatalf("boot track %q %v", track, ok)
	}
}

func TestPlayCoreBoot(t *testing.T) {
	paths := newTestPaths(t)
	logPath := installStubPlayers(t)
	t.Setenv("BGM_STUB_EXIT", "1")
	writeINI(t, &paths, "[bgm]\ncorebootdelay = 1.5\n")
	player := newTestPlayer(t, &paths, ptr(DefaultConfig()))
	var slept time.Duration
	player.sleep = func(duration time.Duration) { slept += duration }
	touchTracks(t, filepath.Join(paths.BootFolder, "Genesis"), "start.wav")
	if err := os.MkdirAll(filepath.Join(paths.BootFolder, "Empty"), 0o755); err != nil {
		t.Fatal(err)
	}
	touchTracks(t, paths.BootFolder, "NES.mp3")

	player.PlayCoreBoot("genesis")
	if got := stubInvocations(t, logPath); len(got) != 1 || !strings.HasSuffix(got[0], "start.wav") {
		t.Fatalf("invocations %v", got)
	}
	if slept != 1500*time.Millisecond {
		t.Fatalf("slept %v", slept)
	}
	player.PlayCoreBoot("empty")
	player.PlayCoreBoot("NES")
	player.PlayCoreBoot("")
	if got := stubInvocations(t, logPath); len(got) != 1 {
		t.Fatalf("empty folders and files must not play: %v", got)
	}
}

func TestPlayCoreBootDefaultFallback(t *testing.T) {
	paths := newTestPaths(t)
	logPath := installStubPlayers(t)
	t.Setenv("BGM_STUB_EXIT", "1")
	writeINI(t, &paths, "[bgm]\ncorebootdelay = 1.5\n")
	player := newTestPlayer(t, &paths, ptr(DefaultConfig()))
	var slept time.Duration
	player.sleep = func(duration time.Duration) { slept += duration }
	touchTracks(t, filepath.Join(paths.BootFolder, "DeFaUlT"), "fallback.wav")
	touchTracks(t, filepath.Join(paths.BootFolder, "Genesis"), "specific.wav")
	touchTracks(t, filepath.Join(paths.BootFolder, "Empty"), "notes.txt")
	touchTracks(t, paths.BootFolder, "startup.wav")

	player.PlayCoreBoot("SNES")
	got := stubInvocations(t, logPath)
	if len(got) != 1 || !strings.HasSuffix(got[0], "fallback.wav") {
		t.Fatalf("missing core folder must use fallback: %v", got)
	}
	if slept != 1500*time.Millisecond {
		t.Fatalf("fallback delay = %v", slept)
	}
	player.PlayCoreBoot("genesis")
	got = stubInvocations(t, logPath)
	if len(got) != 2 || !strings.HasSuffix(got[1], "specific.wav") {
		t.Fatalf("core-specific track must take priority: %v", got)
	}
	player.PlayCoreBoot("empty")
	player.PlayCoreBoot("")
	if got := stubInvocations(t, logPath); len(got) != 2 {
		t.Fatalf("empty core folders and unknown core state must remain silent: %v", got)
	}
	if slept != 3*time.Second {
		t.Fatalf("silent core boots must not delay: %v", slept)
	}
	if err := os.Remove(filepath.Join(paths.BootFolder, "DeFaUlT", "fallback.wav")); err != nil {
		t.Fatal(err)
	}
	player.PlayCoreBoot("SNES")
	if got := stubInvocations(t, logPath); len(got) != 2 {
		t.Fatalf("empty fallback must not reuse startup tracks: %v", got)
	}
}

func TestPlaylistTrackStartedAfterStopIsKilledImmediately(t *testing.T) {
	paths := newTestPaths(t)
	logPath := installStubPlayers(t)
	player := newTestPlayer(t, &paths, ptr(DefaultConfig()))
	track := filepath.Join(paths.MusicFolder, "long.wav")
	touchTracks(t, paths.MusicFolder, "long.wav")
	player.StopPlaylist()

	start := time.Now()
	player.playTrack(track, true)
	if time.Since(start) > 5*time.Second {
		t.Fatal("a playlist track started after StopPlaylist must not run")
	}
	if len(stubInvocations(t, logPath)) != 0 {
		t.Fatal("no player may start once the playlist has ended")
	}
	if _, playing := player.Playing(); playing {
		t.Fatal("playing must be cleared")
	}
}
