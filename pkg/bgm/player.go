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

package bgm

import (
	"bufio"
	"context"
	"fmt"
	"io/fs"
	"math"
	"math/rand/v2"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

// HistorySize is the ratio of total tracks kept in the recently-played list.
const HistorySize = 0.2

// Player is a one-to-one port of the Python Player class. CmdMu is the
// script's self.mutex, serialising socket commands against the core-change
// loop; mu protects the fields below it.
//
//nolint:govet // Field order mirrors the Python class for easier comparison.
type Player struct {
	paths  Paths
	logger *Logger

	CmdMu sync.Mutex

	mu             sync.Mutex
	proc           *playerProcess
	current        *playState
	playback       string
	playlist       Playlist
	playInCore     bool
	bootInPlaylist bool
	history        []string
	endPlaylist    bool
	playlistDone   chan struct{}
	nextToken      uint64

	// randIndex and sleep are replaced by tests.
	randIndex func(n int) int
	sleep     func(time.Duration)
}

type playState struct {
	filename   string
	token      uint64
	inPlaylist bool
}

type playerProcess struct {
	cmd    *exec.Cmd
	reader *os.File
}

// NewPlayer initialises playback settings from the configuration.
func NewPlayer(paths *Paths, logger *Logger, cfg *Config) *Player {
	return &Player{
		paths:          *paths,
		logger:         logger,
		playback:       cfg.Playback,
		playlist:       cfg.Playlist,
		playInCore:     cfg.PlayInCore,
		bootInPlaylist: cfg.BootInPlaylist,
		randIndex:      rand.IntN,
		sleep:          time.Sleep,
	}
}

// Playback returns the current playback type.
func (p *Player) Playback() string {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.playback
}

// SetPlayback stores any playback string, as the socket protocol allows.
func (p *Player) SetPlayback(playback string) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.playback = playback
}

// Playlist returns the current playlist.
func (p *Player) Playlist() Playlist {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.playlist
}

// PlayInCore reports whether music keeps playing inside cores.
func (p *Player) PlayInCore() bool {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.playInCore
}

// SetPlayInCore updates the play-in-core flag.
func (p *Player) SetPlayInCore(enabled bool) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.playInCore = enabled
}

// BootInPlaylist reports whether boot sounds participate in normal playback.
func (p *Player) BootInPlaylist() bool {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.bootInPlaylist
}

// SetBootInPlaylist applies the rotation preference without interrupting a track.
func (p *Player) SetBootInPlaylist(enabled bool) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.bootInPlaylist = enabled
}

// Playing returns the path of the track being played, if any.
func (p *Player) Playing() (string, bool) {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.current == nil {
		return "", false
	}
	return p.current.filename, true
}

// InPlaylist mirrors in_playlist(): true until a playlist has been stopped.
func (p *Player) InPlaylist() bool {
	p.mu.Lock()
	defer p.mu.Unlock()
	return !p.endPlaylist
}

// Status renders the tab-separated reply to the "status" socket command.
func (p *Player) Status() string {
	p.mu.Lock()
	defer p.mu.Unlock()
	isPlaying, filename := "no", ""
	if p.current != nil {
		isPlaying = "yes"
		filename = filepath.Base(p.current.filename)
	}
	return fmt.Sprintf("%s\t%s\t%s\t%s", isPlaying, p.playback, p.playlist.String(), filename)
}

// PlaylistPath mirrors get_playlist_path(): the music folder for no playlist
// or "all", otherwise the named subfolder, falling back to the music folder
// when it does not exist.
func (p *Player) PlaylistPath(playlist Playlist) string {
	if playlist.IsNone() || playlist.IsAll() {
		return p.paths.MusicFolder
	}
	folder := filepath.Join(p.paths.MusicFolder, playlist.Name())
	if _, err := os.Stat(folder); err != nil {
		p.logger.Logf("Playlist folder does not exist: %s", folder)
		return p.paths.MusicFolder
	}
	return folder
}

// filterTracks mirrors filter_tracks(), including its use of the current
// playlist rather than the requested one for the .pls rule.
func (p *Player) filterTracks(names []string, includeBoot bool) []string {
	current := p.Playlist()
	tracks := make([]string, 0, len(names))
	for _, name := range names {
		if !IsValidFile(name) {
			continue
		}
		if includeBoot {
			tracks = append(tracks, name)
			continue
		}
		if strings.HasPrefix(name, "_") && !p.BootInPlaylist() {
			continue
		}
		if current.IsAll() && IsPLS(name) {
			continue
		}
		tracks = append(tracks, name)
	}
	return tracks
}

// Tracks mirrors get_tracks(): the top level only for no playlist, otherwise
// a recursive walk of the playlist folder.
func (p *Player) Tracks(playlist Playlist, includeBoot bool) []string {
	folder := p.PlaylistPath(playlist)
	var tracks []string
	if playlist.IsNone() {
		entries, err := os.ReadDir(folder)
		if err != nil {
			return nil
		}
		names := make([]string, 0, len(entries))
		for _, entry := range entries {
			names = append(names, entry.Name())
		}
		for _, name := range p.filterTracks(names, includeBoot) {
			tracks = append(tracks, filepath.Join(folder, name))
		}
		return p.withGlobalBootTracks(tracks, includeBoot)
	}
	byFolder := make(map[string][]string)
	var order []string
	_ = filepath.WalkDir(folder, func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return nil //nolint:nilerr // os.walk silently skips unreadable entries.
		}
		if path == folder || entry.IsDir() || isSymlinkToDir(path, entry) {
			return nil
		}
		parent := filepath.Dir(path)
		if _, seen := byFolder[parent]; !seen {
			order = append(order, parent)
		}
		byFolder[parent] = append(byFolder[parent], entry.Name())
		return nil
	})
	for _, parent := range order {
		for _, name := range p.filterTracks(byFolder[parent], includeBoot) {
			tracks = append(tracks, filepath.Join(parent, name))
		}
	}
	return p.withGlobalBootTracks(tracks, includeBoot)
}

func (p *Player) withGlobalBootTracks(tracks []string, includeBoot bool) []string {
	if !p.BootInPlaylist() {
		return tracks
	}
	entries, err := os.ReadDir(p.paths.BootFolder)
	if err != nil {
		return tracks
	}
	var names []string
	for _, entry := range entries {
		if !entry.IsDir() {
			names = append(names, entry.Name())
		}
	}
	for _, name := range p.filterTracks(names, includeBoot) {
		path := filepath.Join(p.paths.BootFolder, name)
		if !containsString(tracks, path) {
			tracks = append(tracks, path)
		}
	}
	return tracks
}

// isSymlinkToDir mirrors os.walk placing symlinked folders in dirs rather
// than files.
func isSymlinkToDir(path string, entry fs.DirEntry) bool {
	if entry.Type()&fs.ModeSymlink == 0 {
		return false
	}
	info, err := os.Stat(path)
	return err == nil && info.IsDir()
}

// TotalTracks counts the tracks Tracks would return.
func (p *Player) TotalTracks(playlist Playlist, includeBoot bool) int {
	return len(p.Tracks(playlist, includeBoot))
}

func (p *Player) addHistory(filename string) {
	size := int(math.Floor(float64(p.TotalTracks(p.Playlist(), false)) * HistorySize))
	if p.BootInPlaylist() {
		size = max(size, 1)
	}
	if size < 1 {
		return
	}
	p.mu.Lock()
	defer p.mu.Unlock()
	for len(p.history) > size {
		p.history = p.history[1:]
	}
	p.history = append(p.history, filename)
}

// History returns a copy of the recently played tracks.
func (p *Player) History() []string {
	p.mu.Lock()
	defer p.mu.Unlock()
	return append([]string(nil), p.history...)
}

// Stop mirrors stop(): only when a player process exists is the current
// track cleared and the process killed.
func (p *Player) Stop() {
	p.mu.Lock()
	proc := p.proc
	if proc != nil {
		p.current = nil
		p.proc = nil
	}
	p.mu.Unlock()
	if proc != nil {
		killPlayer(proc.cmd)
		_ = proc.reader.Close()
	}
}

// Play mirrors play(): blocking playback honouring the X##_ loop prefix.
func (p *Player) Play(filename string) {
	p.playTrack(filename, false)
}

// playTrack runs one track. Playlist tracks additionally stop as soon as the
// playlist has ended, so StopPlaylist can never leave a stray track running.
func (p *Player) playTrack(filename string, inPlaylist bool) {
	p.Stop()
	if !IsValidFile(filename) {
		return
	}
	p.mu.Lock()
	p.nextToken++
	state := &playState{filename: filename, token: p.nextToken, inPlaylist: inPlaylist}
	p.current = state
	p.mu.Unlock()
	p.addHistory(filename)
	p.logger.Logf("Now playing: %s", filename)

	for loop := LoopAmount(filename); loop > 0 && p.shouldContinue(state); loop-- {
		p.logger.Logf("Loop #%d", loop)
		switch {
		case IsMP3(filename), IsPLS(filename):
			p.playMP3(state, filename)
		case IsOGG(filename):
			p.playFile(state, "ogg123", filename)
		case IsWAV(filename):
			p.playFile(state, "aplay", filename)
		case IsMID(filename):
			p.playFile(state, "aplaymidi", filename, "--port="+MIDIPort)
		case IsVGM(filename):
			p.playFile(state, "vgmplay", filename)
		}
	}

	p.mu.Lock()
	if p.current == state {
		p.current = nil
	}
	p.mu.Unlock()
}

// shouldContinue reports whether state is still the current track and, for
// playlist tracks, whether the playlist is still running.
func (p *Player) shouldContinue(state *playState) bool {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.current == state && (!state.inPlaylist || !p.endPlaylist)
}

// playMP3 runs mpg123, killing it as soon as it reports the track finished
// to work around the hang the Python script describes.
func (p *Player) playMP3(state *playState, filename string) {
	if IsPLS(filename) {
		filename = PLSURL(filename, p.logger)
	}
	proc, err := p.startProcess(state, "mpg123", "--no-control", filename)
	if err != nil {
		p.logger.Log(err.Error())
		return
	}
	scanner := bufio.NewScanner(proc.reader)
	for scanner.Scan() {
		line := strings.TrimRight(scanner.Text(), " \t\r\n\f\v")
		p.logger.Log(line)
		if strings.Contains(line, "finished.") {
			break
		}
	}
	p.finishProcess(proc)
}

// playFile runs a player until it exits, logging its output.
func (p *Player) playFile(state *playState, name string, args ...string) {
	proc, err := p.startProcess(state, name, args...)
	if err != nil {
		p.logger.Log(err.Error())
		return
	}
	scanner := bufio.NewScanner(proc.reader)
	for scanner.Scan() {
		p.logger.Log(strings.TrimRight(scanner.Text(), " \t\r\n\f\v"))
	}
	p.finishProcess(proc)
}

func (p *Player) startProcess(state *playState, name string, args ...string) (*playerProcess, error) {
	reader, writer, err := os.Pipe()
	if err != nil {
		return nil, fmt.Errorf("create %s output pipe: %w", name, err)
	}
	// #nosec G204 -- players are fixed MiSTer tools; arguments are music file paths.
	cmd := exec.CommandContext(context.Background(), name, args...)
	cmd.Stdout = writer
	cmd.Stderr = writer
	cmd.SysProcAttr = playerAttributes()
	if err := cmd.Start(); err != nil {
		_ = reader.Close()
		_ = writer.Close()
		return nil, fmt.Errorf("start %s: %w", name, err)
	}
	_ = writer.Close()
	proc := &playerProcess{cmd: cmd, reader: reader}
	p.mu.Lock()
	p.proc = proc
	ended := state.inPlaylist && p.endPlaylist
	p.mu.Unlock()
	if ended {
		// The playlist stopped while this player was starting.
		killPlayer(cmd)
	}
	return proc, nil
}

// finishProcess mirrors kill_player() for the process this call started.
func (p *Player) finishProcess(proc *playerProcess) {
	p.mu.Lock()
	if p.proc == proc {
		p.proc = nil
	}
	p.mu.Unlock()
	killPlayer(proc.cmd)
	_ = proc.reader.Close()
	_ = proc.cmd.Wait()
}

// randomTrack mirrors get_random_track() with a guard against the Python
// infinite loop when every remaining track is in the history.
func (p *Player) randomTrack() (string, bool) {
	tracks := p.Tracks(p.Playlist(), false)
	if len(tracks) == 0 {
		return "", false
	}
	history := p.History()
	candidates := make([]string, 0, len(tracks))
	for _, track := range tracks {
		if !containsString(history, track) {
			candidates = append(candidates, track)
		}
	}
	if len(candidates) == 0 {
		candidates = tracks
		// Exhausted history must not immediately repeat a boot sound when
		// another track exists, including in very small playlists.
		if p.BootInPlaylist() && len(history) > 0 && len(tracks) > 1 {
			candidates = make([]string, 0, len(tracks))
			for _, track := range tracks {
				if track != history[len(history)-1] {
					candidates = append(candidates, track)
				}
			}
		}
	}
	return candidates[p.randIndex(len(candidates))], true
}

func containsString(values []string, value string) bool {
	for _, candidate := range values {
		if candidate == value {
			return true
		}
	}
	return false
}

func (p *Player) beginPlaylist() chan struct{} {
	done := make(chan struct{})
	p.mu.Lock()
	p.endPlaylist = false
	p.playlistDone = done
	p.mu.Unlock()
	return done
}

func (p *Player) startRandomPlaylist() {
	p.logger.Log("Starting random playlist...")
	done := p.beginPlaylist()
	go func() {
		defer close(done)
		for p.InPlaylist() {
			track, ok := p.randomTrack()
			if !ok {
				break
			}
			p.playTrack(track, true)
		}
		p.logger.Log("Random playlist ended")
	}()
}

func (p *Player) startLoopPlaylist() {
	p.logger.Log("Starting loop playlist...")
	done := p.beginPlaylist()
	track, ok := p.randomTrack()
	go func() {
		defer close(done)
		for p.InPlaylist() && ok {
			p.playTrack(track, true)
		}
		p.logger.Log("Loop playlist ended")
	}()
}

// StartPlaylist mirrors start_playlist(): known types are stored, "disabled"
// only stops, and anything else plays random tracks without changing the
// stored playback value.
func (p *Player) StartPlaylist(playback string) {
	p.StopPlaylist()
	switch playback {
	case PlaybackRandom:
		p.SetPlayback(PlaybackRandom)
		p.startRandomPlaylist()
	case PlaybackLoop:
		p.SetPlayback(PlaybackLoop)
		p.startLoopPlaylist()
	case PlaybackDisabled:
		p.SetPlayback(PlaybackDisabled)
	default:
		p.startRandomPlaylist()
	}
}

// StartCurrentPlaylist mirrors start_playlist() with no argument.
func (p *Player) StartCurrentPlaylist() {
	p.StartPlaylist(p.Playback())
}

// StopPlaylist mirrors stop_playlist() and additionally waits for the
// playlist goroutine so a stopped playlist can never start one more track.
func (p *Player) StopPlaylist() {
	p.mu.Lock()
	p.endPlaylist = true
	done := p.playlistDone
	p.mu.Unlock()
	p.Stop()
	if done != nil {
		<-done
	}
}

// ChangePlaylist mirrors change_playlist(), including accepting names whose
// folder does not exist because the count falls back to the music folder.
func (p *Player) ChangePlaylist(name string) {
	target := ParsePlaylist(name)
	check := target
	if target.IsNone() {
		check = p.Playlist()
	}
	p.PlaylistPath(target)
	if p.TotalTracks(check, true) == 0 {
		return
	}
	logName := target.Name()
	if target.IsNone() {
		logName = "None"
	}
	p.logger.Logf("Changed playlist: %s", logName)
	p.mu.Lock()
	p.playlist = target
	p.mu.Unlock()
	if p.InPlaylist() {
		p.StartCurrentPlaylist()
	}
}

// BootTrack mirrors get_boot_track(): underscore-prefixed files at the top of
// the current playlist folder plus every file at the top of music/boot.
func (p *Player) BootTrack() (string, bool) {
	var candidates []string
	folder := p.PlaylistPath(p.Playlist())
	if entries, err := os.ReadDir(folder); err == nil {
		for _, entry := range entries {
			if strings.HasPrefix(entry.Name(), "_") && IsValidFile(entry.Name()) {
				candidates = append(candidates, filepath.Join(folder, entry.Name()))
			}
		}
	}
	if entries, err := os.ReadDir(p.paths.BootFolder); err == nil {
		for _, entry := range entries {
			if IsValidFile(entry.Name()) {
				candidates = append(candidates, filepath.Join(p.paths.BootFolder, entry.Name()))
			}
		}
	}
	if len(candidates) == 0 {
		return "", false
	}
	return candidates[p.randIndex(len(candidates))], true
}

// PlayBoot plays a boot track synchronously when one exists.
func (p *Player) PlayBoot() {
	track, ok := p.BootTrack()
	if !ok {
		return
	}
	p.logger.Logf("Selected boot track: %s", track)
	p.Play(track)
}

// PlayCoreBoot mirrors play_core_boot(): a case-insensitive folder match in
// music/boot plays one random track after the configured delay. An empty
// matching folder ends the search; a played one lets it continue.
func (p *Player) PlayCoreBoot(core string) {
	entries, err := os.ReadDir(p.paths.BootFolder)
	if err != nil {
		return
	}
	for _, entry := range entries {
		if !strings.EqualFold(entry.Name(), core) {
			continue
		}
		folder := filepath.Join(p.paths.BootFolder, entry.Name())
		info, statErr := os.Stat(folder)
		if statErr != nil || !info.IsDir() {
			continue
		}
		files, readErr := os.ReadDir(folder)
		if readErr != nil {
			continue
		}
		var tracks []string
		for _, file := range files {
			if IsValidFile(file.Name()) {
				tracks = append(tracks, filepath.Join(folder, file.Name()))
			}
		}
		if len(tracks) == 0 {
			return
		}
		cfg, _ := LoadConfig(&p.paths)
		p.sleep(time.Duration(cfg.CoreBootDelay * float64(time.Second)))
		p.logger.Log("Playing core boot track...")
		p.Play(tracks[p.randIndex(len(tracks))])
	}
}
