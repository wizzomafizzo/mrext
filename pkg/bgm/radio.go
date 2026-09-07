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
	"errors"
	"fmt"
	"mime"
	"net"
	"net/http"
	"net/url"
	"path/filepath"
	"strings"
	"time"
)

const radioRetryDelay = 5 * time.Second

func newRadioClient() *http.Client {
	return &http.Client{
		Transport: &http.Transport{
			Proxy:                 http.ProxyFromEnvironment,
			DialContext:           (&net.Dialer{Timeout: 10 * time.Second, KeepAlive: 30 * time.Second}).DialContext,
			TLSHandshakeTimeout:   10 * time.Second,
			ResponseHeaderTimeout: 15 * time.Second,
			IdleConnTimeout:       30 * time.Second,
		},
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			if len(via) >= 10 {
				return errors.New("too many radio redirects")
			}
			if req.URL.Scheme != "http" && req.URL.Scheme != "https" {
				return errors.New("radio redirect must use HTTP or HTTPS")
			}
			if via[len(via)-1].URL.Scheme == "https" && req.URL.Scheme != "https" {
				return errors.New("refusing radio redirect from HTTPS to insecure HTTP")
			}
			return nil
		},
	}
}

func (p *Player) playRadio(state *playState, filename string) {
	started := time.Now()
	err := p.streamRadio(state, filename)
	if state.ctx.Err() != nil {
		return
	}
	p.reportRadioError(filename, err)
	// Pace even clean but short-lived streams: a closed connection or rejected
	// format must not cause an unbounded player-spawn/reconnect loop.
	if remaining := p.radioRetryDelay - time.Since(started); remaining > 0 {
		timer := time.NewTimer(remaining)
		defer timer.Stop()
		select {
		case <-timer.C:
		case <-state.ctx.Done():
		}
	}
}

// Suppress identical station failures to keep default-mode retry logs bounded.
func (p *Player) reportRadioError(filename string, err error) {
	p.mu.Lock()
	if err == nil {
		delete(p.radioErrors, filename)
		p.mu.Unlock()
		return
	}
	message := fmt.Sprintf("Radio %s: %v", filepath.Base(filename), err)
	changed := p.radioErrors[filename] != message
	p.radioErrors[filename] = message
	p.mu.Unlock()
	if changed {
		p.logger.Error(message)
	}
}

func (p *Player) streamRadio(state *playState, filename string) error {
	address := PLSURL(filename, p.logger)
	if address == "" {
		return errors.New("playlist has no HTTP/HTTPS stream URL")
	}
	req, err := http.NewRequestWithContext(state.ctx, http.MethodGet, address, http.NoBody)
	if err != nil {
		return errors.New("invalid radio stream URL")
	}
	req.Header.Set("Icy-MetaData", "0")
	response, err := p.radioClient.Do(req)
	if err != nil {
		// URL errors include the full address, which can contain stream credentials.
		var requestErr *url.Error
		if errors.As(err, &requestErr) {
			err = requestErr.Err
		}
		return fmt.Errorf("fetch stream (HTTPS requires a valid clock and trusted certificates): %w", err)
	}
	defer func() { _ = response.Body.Close() }()
	if response.StatusCode != http.StatusOK {
		return fmt.Errorf("station returned HTTP %d", response.StatusCode)
	}
	mediaType := response.Header.Get("Content-Type")
	if mediaType != "" {
		mediaType, _, err = mime.ParseMediaType(mediaType)
		if err != nil {
			return errors.New("station returned an invalid Content-Type")
		}
	}
	switch mediaType {
	case "", "audio/mpeg", "audio/mp3", "audio/x-mpeg", "audio/x-mp3", "application/octet-stream":
	default:
		return fmt.Errorf(
			"unsupported stream type %q; use direct MP3/MPEG audio, not AAC, HLS, or a web page", mediaType,
		)
	}
	if interval := response.Header.Get("Icy-Metaint"); interval != "" && interval != "0" {
		return errors.New("station sent ICY metadata despite requesting audio only; use a metadata-free MP3 stream")
	}
	proc, err := p.startProcessInput(state, response.Body, "mpg123", "--no-control", "-")
	if err != nil {
		return err
	}
	scanner := bufio.NewScanner(proc.reader)
	lastLine := ""
	finished := false
	for scanner.Scan() {
		lastLine = strings.TrimSpace(scanner.Text())
		p.logger.Log(lastLine)
		if strings.Contains(lastLine, "finished.") {
			finished = true
			break
		}
	}
	// exec.Cmd copies stdin asynchronously. Close the network reader before
	// Wait, including when mpg123 rejects audio while the station keeps sending.
	_ = response.Body.Close()
	waitErr := p.finishProcess(proc)
	if scanErr := scanner.Err(); scanErr != nil {
		return fmt.Errorf("read radio player output: %w", scanErr)
	}
	if waitErr != nil && !finished {
		return fmt.Errorf("mpg123 could not play the stream (MP3/MPEG audio required): %w; %s", waitErr, lastLine)
	}
	return nil
}
