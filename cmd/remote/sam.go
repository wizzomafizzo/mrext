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

package main

import (
	"fmt"
	"net/http"
	"os"

	"github.com/wizzomafizzo/mrext/pkg/config"
	"github.com/wizzomafizzo/mrext/pkg/service"
)

func withSAMActivity(next http.HandlerFunc, logger *service.Logger) http.HandlerFunc {
	return afterSuccessfulLaunch(next, func() error { return signalSAMActivity(config.SAMActivityFile) }, logger.Error)
}

func afterSuccessfulLaunch(
	next http.HandlerFunc, notify func() error, logError func(string, ...any),
) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		result := &launchResponse{ResponseWriter: w}
		next(result, r)
		if result.status == 0 || result.status >= 200 && result.status < 300 {
			if err := notify(); err != nil {
				logError("notify SAM after launch: %s", err)
			}
		}
	}
}

type launchResponse struct {
	http.ResponseWriter
	status int
}

func (w *launchResponse) WriteHeader(status int) {
	if w.status == 0 {
		w.status = status
		w.ResponseWriter.WriteHeader(status)
	}
}

func (w *launchResponse) Write(data []byte) (int, error) {
	if w.status == 0 {
		w.WriteHeader(http.StatusOK)
	}
	n, err := w.ResponseWriter.Write(data)
	if err != nil {
		return n, fmt.Errorf("write launch response: %w", err)
	}
	return n, nil
}

func signalSAMActivity(path string) error {
	info, err := os.Lstat(path)
	if os.IsNotExist(err) {
		return nil
	}
	if err != nil {
		return fmt.Errorf("inspect SAM activity file: %w", err)
	}
	if !info.Mode().IsRegular() {
		return fmt.Errorf("SAM activity path is not a regular file: %s", path)
	}
	// SAM owns this file. Never create it or its parent when SAM is absent.
	// #nosec G304 -- path is fixed by config in production and injected only by tests.
	file, err := os.OpenFile(path, os.O_WRONLY|os.O_TRUNC, 0)
	if os.IsNotExist(err) {
		return nil
	}
	if err != nil {
		return fmt.Errorf("open SAM activity file: %w", err)
	}
	// MCP recognizes this exact token as external activity and keeps the
	// current game running. Its generic keyboard action can return to Menu.
	_, writeErr := file.WriteString("zaparoo\n")
	closeErr := file.Close()
	if writeErr != nil {
		return fmt.Errorf("write SAM activity: %w", writeErr)
	}
	if closeErr != nil {
		return fmt.Errorf("close SAM activity file: %w", closeErr)
	}
	return nil
}
