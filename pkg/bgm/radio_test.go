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
	"errors"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"sync/atomic"
	"testing"
	"time"
)

func radioFixture(t *testing.T, address string) (*Player, string, *syncBuffer) {
	t.Helper()
	paths := newTestPaths(t)
	logger, out := newTestLogger(&paths)
	player := newTestPlayer(t, &paths, ptr(DefaultConfig()))
	player.logger = logger
	file := filepath.Join(paths.MusicFolder, "radio.pls")
	writeFile(t, file, "[Playlist]\nFile1="+address+"\n")
	return player, file, out
}

func TestRadioHTTPSVerifiedAndPipedToPlayer(t *testing.T) {
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Icy-MetaData") != "0" {
			t.Error("metadata not disabled")
		}
		w.Header().Set("Content-Type", "audio/mpeg")
		_, _ = w.Write([]byte("stream bytes"))
	}))
	t.Cleanup(server.Close)
	logPath := installStubPlayers(t)
	t.Setenv("BGM_STUB_EXIT", "1")
	input := filepath.Join(t.TempDir(), "input")
	t.Setenv("BGM_STUB_INPUT", input)
	player, file, out := radioFixture(t, server.URL)
	player.Play(file)
	if !strings.Contains(out.String(), "certificate") {
		t.Fatalf("missing TLS error: %s", out.String())
	}
	if len(stubInvocations(t, logPath)) != 0 {
		t.Fatal("untrusted stream reached decoder")
	}
	// Trust only the local fixture certificate; production verification stays on.
	player.radioClient.Transport = server.Client().Transport
	player.Play(file)
	if got := readFile(t, input); got != "stream bytes" {
		t.Fatalf("input %q", got)
	}
	calls := stubInvocations(t, logPath)
	if len(calls) != 1 || calls[0] != "mpg123 --no-control -" {
		t.Fatalf("calls %v", calls)
	}
}

func TestRadioRejectsHTTPSDowngrade(t *testing.T) {
	var reached atomic.Bool
	plain := httptest.NewServer(http.HandlerFunc(func(http.ResponseWriter, *http.Request) { reached.Store(true) }))
	t.Cleanup(plain.Close)
	secure := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, plain.URL, http.StatusFound)
	}))
	t.Cleanup(secure.Close)
	player, file, out := radioFixture(t, secure.URL)
	player.radioClient.Transport = secure.Client().Transport
	player.Play(file)
	if reached.Load() || !strings.Contains(out.String(), "insecure HTTP") {
		t.Fatalf("downgrade: %s", out.String())
	}
}

func TestRadioRejectsUnsupportedResponses(t *testing.T) {
	for _, scenario := range []struct {
		name, mediaType, message string
		status                   int
	}{
		{name: "AAC", mediaType: "audio/aacp", status: 200, message: "unsupported stream type"},
		{name: "HLS", mediaType: "application/vnd.apple.mpegurl", status: 200, message: "unsupported stream type"},
		{name: "HTTP error", status: 503, message: "HTTP 503"},
	} {
		t.Run(scenario.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
				w.Header().Set("Content-Type", scenario.mediaType)
				w.WriteHeader(scenario.status)
			}))
			t.Cleanup(server.Close)
			calls := installStubPlayers(t)
			player, file, out := radioFixture(t, server.URL)
			player.Play(file)
			if !strings.Contains(out.String(), scenario.message) {
				t.Fatalf("error: %s", out.String())
			}
			if len(stubInvocations(t, calls)) != 0 {
				t.Fatal("unsupported response reached decoder")
			}
		})
	}
}

func TestRadioStopCancelsHeadersAndBody(t *testing.T) {
	for _, body := range []bool{false, true} {
		t.Run(map[bool]string{false: "headers", true: "body"}[body], func(t *testing.T) {
			requested := make(chan struct{})
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if body {
					w.Header().Set("Content-Type", "audio/mpeg")
					w.WriteHeader(http.StatusOK)
					w.(http.Flusher).Flush()
				}
				close(requested)
				<-r.Context().Done()
			}))
			t.Cleanup(server.Close)
			installStubPlayers(t)
			t.Setenv("BGM_STUB_INPUT", filepath.Join(t.TempDir(), "input"))
			player, file, _ := radioFixture(t, server.URL)
			done := make(chan struct{})
			go func() { player.Play(file); close(done) }()
			select {
			case <-requested:
			case <-time.After(5 * time.Second):
				t.Fatal("no request")
			}
			if body {
				waitFor(t, func() bool { player.mu.Lock(); defer player.mu.Unlock(); return player.proc != nil })
			}
			player.Stop()
			select {
			case <-done:
			case <-time.After(2 * time.Second):
				t.Fatal("Stop blocked on stream")
			}
		})
	}
}

func TestRadioDecoderFailureClosesStalledInput(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/octet-stream")
		w.WriteHeader(http.StatusOK)
		w.(http.Flusher).Flush()
		<-r.Context().Done()
	}))
	t.Cleanup(server.Close)
	installStubPlayers(t)
	t.Setenv("BGM_STUB_FAIL", "1")
	player, file, out := radioFixture(t, server.URL)
	done := make(chan struct{})
	go func() { player.Play(file); close(done) }()
	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("decoder exit hung on stdin")
	}
	if !strings.Contains(out.String(), "unsupported audio encoding") {
		t.Fatalf("missing decoder error: %s", out.String())
	}
}

func TestRadioErrorsPersistWithoutDebugAndDoNotFlood(t *testing.T) {
	player, file, out := radioFixture(t, "http://unused.invalid")
	failure := errors.New("unsupported codec")
	player.reportRadioError(file, failure)
	player.reportRadioError(file, failure)
	if strings.Count(out.String(), "unsupported codec") != 1 {
		t.Fatal("repeated console error")
	}
	if strings.Count(readFile(t, player.paths.LogFile), "unsupported codec") != 1 {
		t.Fatal("missing or repeated log error")
	}
	player.reportRadioError(file, nil)
	player.reportRadioError(file, failure)
	if strings.Count(out.String(), "unsupported codec") != 2 {
		t.Fatal("recovered station did not report a new failure")
	}
}

func TestRadioRetryDelayIsInterruptible(t *testing.T) {
	var requests atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		requests.Add(1)
		w.WriteHeader(http.StatusServiceUnavailable)
	}))
	t.Cleanup(server.Close)
	player, _, out := radioFixture(t, server.URL)
	player.radioRetryDelay = time.Hour
	player.StartPlaylist(PlaybackLoop)
	waitFor(t, func() bool { return strings.Contains(out.String(), "HTTP 503") })
	time.Sleep(30 * time.Millisecond)
	if requests.Load() != 1 {
		t.Fatal("failed station retried without pacing")
	}
	done := make(chan struct{})
	go func() { player.StopPlaylist(); close(done) }()
	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("Stop waited for retry timer")
	}
}
