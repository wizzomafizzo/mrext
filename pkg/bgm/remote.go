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
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"os"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"
)

// MessageSize is the maximum socket payload, matching the Python script.
const MessageSize = 4096

// idleReadTimeout bounds how long the accept loop waits for a connected
// client to send its command, so Close never hangs on an idle client.
var idleReadTimeout = 5 * time.Second

// Remote serves the /tmp/bgm.sock line protocol: one command per connection,
// an optional reply, then the connection is closed.
type Remote struct {
	listener  net.Listener
	logger    *Logger
	player    *Player
	done      chan struct{}
	closeOnce sync.Once
}

// StartRemote binds the socket and serves commands in the background.
func StartRemote(paths *Paths, logger *Logger, player *Player) (*Remote, error) {
	listener, err := (&net.ListenConfig{}).Listen(context.Background(), "unix", paths.SocketFile)
	if err != nil {
		return nil, fmt.Errorf("bind BGM socket: %w", err)
	}
	if unixListener, ok := listener.(*net.UnixListener); ok {
		// Python left the file behind after "quit"; cleanup removes it.
		unixListener.SetUnlinkOnClose(false)
	}
	remote := &Remote{listener: listener, logger: logger, player: player, done: make(chan struct{})}
	logger.Log("Starting remote...")
	go remote.serve()
	return remote, nil
}

// Done is closed once the accept loop has stopped.
func (r *Remote) Done() <-chan struct{} { return r.done }

// Close stops accepting commands and waits for the loop to exit.
func (r *Remote) Close() {
	r.closeOnce.Do(func() { _ = r.listener.Close() })
	<-r.done
}

func (r *Remote) serve() {
	defer close(r.done)
	for {
		conn, err := r.listener.Accept()
		if err != nil {
			break
		}
		buffer := make([]byte, MessageSize)
		_ = conn.SetReadDeadline(time.Now().Add(idleReadTimeout))
		count, _ := conn.Read(buffer)
		if count == 0 {
			// A liveness probe connected without sending a command, or an
			// idle client never did.
			_ = conn.Close()
			continue
		}
		command := string(buffer[:count])
		if command == "quit" {
			_ = conn.Close()
			break
		}
		if reply, ok := r.Handle(command); ok {
			_, _ = conn.Write([]byte(reply))
		}
		_ = conn.Close()
	}
	r.closeOnce.Do(func() { _ = r.listener.Close() })
	r.logger.Log("Remote stopped")
}

// Handle runs one command under the command mutex. ok=false means Python
// returned None and sent nothing.
func (r *Remote) Handle(command string) (reply string, ok bool) {
	player := r.player
	player.CmdMu.Lock()
	defer player.CmdMu.Unlock()

	switch {
	case command == "stop":
		player.StopPlaylist()
	case command == "play":
		player.StartCurrentPlaylist()
	case command == "skip":
		player.Stop()
	case command == "pid":
		return strconv.Itoa(os.Getpid()), true
	case command == "status":
		return player.Status(), true
	case strings.HasPrefix(command, "set playlist"):
		if parts := strings.SplitN(command, " ", 3); len(parts) > 2 {
			player.ChangePlaylist(parts[2])
		}
	case strings.HasPrefix(command, "set playback"):
		if parts := strings.SplitN(command, " ", 3); len(parts) > 2 {
			player.SetPlayback(parts[2])
			if player.InPlaylist() {
				player.StartCurrentPlaylist()
			}
		}
	case command == "set bootinplaylist yes":
		player.SetBootInPlaylist(true)
	case command == "set bootinplaylist no":
		player.SetBootInPlaylist(false)
	case command == "set playincore yes":
		player.SetPlayInCore(true)
	case command == "set playincore no":
		player.SetPlayInCore(false)
	case strings.HasPrefix(command, "get"):
		return r.handleGet(command)
	default:
		r.logger.Logf("Unknown command: %s", command)
	}
	return "", false
}

func (r *Remote) handleGet(command string) (reply string, ok bool) {
	parts := strings.SplitN(command, " ", 2)
	if len(parts) < 2 {
		return "", true
	}
	switch parts[1] {
	case "playlist":
		playlist := r.player.Playlist()
		if playlist.IsNone() {
			return "", false
		}
		return playlist.Name(), true
	case "playback":
		return r.player.Playback(), true
	case "playincore":
		if r.player.PlayInCore() {
			return "yes", true
		}
		return "no", true
	default:
		return "", true
	}
}

// Send mirrors send_socket(): nothing happens without a socket file, and an
// empty reply is reported as no reply.
func Send(paths *Paths, message string) (reply string, replied bool, err error) {
	if _, statErr := os.Stat(paths.SocketFile); statErr != nil {
		return "", false, nil
	}
	dialer := net.Dialer{Timeout: 5 * time.Second}
	conn, err := dialer.DialContext(context.Background(), "unix", paths.SocketFile)
	if err != nil {
		return "", false, fmt.Errorf("connect to BGM service: %w", err)
	}
	defer func() { _ = conn.Close() }()
	if _, writeErr := conn.Write([]byte(message)); writeErr != nil {
		return "", false, fmt.Errorf("send BGM command: %w", writeErr)
	}
	buffer := make([]byte, MessageSize)
	count, err := conn.Read(buffer)
	if err != nil && !errors.Is(err, io.EOF) {
		return "", false, fmt.Errorf("read BGM reply: %w", err)
	}
	if count == 0 {
		return "", false, nil
	}
	return string(buffer[:count]), true, nil
}

// SocketExists reports whether the socket file is present.
func SocketExists(paths *Paths) bool {
	_, err := os.Stat(paths.SocketFile)
	return err == nil
}

// SocketStale reports a socket file nobody is listening on, which the Python
// script could never recover from without a reboot.
func SocketStale(paths *Paths) bool {
	if !SocketExists(paths) {
		return false
	}
	dialer := net.Dialer{Timeout: time.Second}
	conn, err := dialer.DialContext(context.Background(), "unix", paths.SocketFile)
	if err != nil {
		return errors.Is(err, syscall.ECONNREFUSED) || errors.Is(err, syscall.ENOENT)
	}
	_ = conn.Close()
	return false
}
