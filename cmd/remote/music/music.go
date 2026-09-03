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

package music

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/gorilla/mux"
	"github.com/wizzomafizzo/mrext/pkg/config"
	"github.com/wizzomafizzo/mrext/pkg/service"
)

type Service struct {
	Playback string `json:"playback"`
	Playlist string `json:"playlist"`
	Track    string `json:"track"`
	Running  bool   `json:"running"`
	Playing  bool   `json:"playing"`
}

type Playlists []string

const (
	musicFolder  = config.SdFolder + "/music"
	musicSocket  = "/tmp/bgm.sock"
	socketBuffer = 4096
)

func sendCmd(cmd string) (string, error) {
	dialer := &net.Dialer{Timeout: 5 * time.Second}
	conn, err := dialer.DialContext(context.Background(), "unix", musicSocket)
	if err != nil {
		return "", fmt.Errorf("connect to music service: %w", err)
	}
	defer func(conn net.Conn) {
		_ = conn.Close()
	}(conn)

	_, err = conn.Write([]byte(cmd))
	if err != nil {
		return "", fmt.Errorf("write music command: %w", err)
	}

	buf := make([]byte, socketBuffer)
	_, err = conn.Read(buf)
	if err != nil && !errors.Is(err, io.EOF) {
		return "", fmt.Errorf("read music response: %w", err)
	}

	return string(bytes.Trim(buf, "\x00")), nil
}

func Status(logger *service.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, _ *http.Request) {
		var status Service

		_, err := os.Stat(musicSocket)
		if err != nil {
			status.Running = false
		} else {
			status.Running = true
		}

		if !status.Running {
			err = json.NewEncoder(w).Encode(status)
			if err != nil {
				logger.Error("failed to encode server status: %s", err)
			}
			return
		}

		resp, err := sendCmd("status")
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			logger.Error("bgm status: %s", err)
			return
		}

		states := strings.Split(resp, "\t")
		if len(states) < 4 {
			http.Error(w, "invalid response from bgm: "+resp, http.StatusInternalServerError)
			logger.Error("invalid response from bgm: %s", resp)
			return
		}

		status.Playing = states[0] == "yes"
		status.Playback = states[1]
		status.Playlist = states[2]
		status.Track = states[3]

		err = json.NewEncoder(w).Encode(status)
		if err != nil {
			logger.Error("failed to encode server status: %s", err)
		}
	}
}

func Play(logger *service.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, _ *http.Request) {
		_, err := sendCmd("play")
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			logger.Error("bgm play: %s", err)
			return
		}
		time.Sleep(500 * time.Millisecond)
	}
}

func Stop(logger *service.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, _ *http.Request) {
		_, err := sendCmd("stop")
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			logger.Error("bgm stop: %s", err)
			return
		}
		time.Sleep(500 * time.Millisecond)
	}
}

func Skip(logger *service.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, _ *http.Request) {
		_, err := sendCmd("skip")
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			logger.Error("bgm skip: %s", err)
			return
		}
		time.Sleep(500 * time.Millisecond)
	}
}

func SetPlayback(logger *service.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		playback := vars["playback"]

		_, err := sendCmd("set playback " + playback)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			logger.Error("bgm set playback: %s (%s)", err, playback)
			return
		}
		time.Sleep(500 * time.Millisecond)
	}
}

func SetPlaylist(logger *service.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		playlist := vars["playlist"]

		_, err := sendCmd("set playlist " + playlist)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			logger.Error("bgm set playlist: %s (%s)", err, playlist)
			return
		}
		time.Sleep(500 * time.Millisecond)
	}
}

func AllPlaylists(logger *service.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, _ *http.Request) {
		var playlists Playlists

		items, err := os.ReadDir(musicFolder)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			logger.Error("listing bgm playlists: %s", err)
			return
		}

		playlists = append(playlists, "none")

		for _, item := range items {
			if item.IsDir() && item.Name() != "boot" {
				playlists = append(playlists, item.Name())
			}
		}

		err = json.NewEncoder(w).Encode(playlists)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			logger.Error("encoding bgm playlists: %s", err)
			return
		}
	}
}
