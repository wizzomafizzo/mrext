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
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
)

const startupMarker = "Startup BGM"

// TryAddToStartup mirrors try_add_to_startup(): create user-startup.sh when
// missing, then append the BGM hook once.
func TryAddToStartup(paths *Paths, logger *Logger) error {
	if _, err := os.Stat(paths.StartupScript); err != nil {
		// #nosec G306 -- MiSTer's startup script must stay editable and executable.
		if writeErr := os.WriteFile(paths.StartupScript, []byte("#!/bin/sh\n"), 0o755); writeErr != nil {
			return fmt.Errorf("create startup script: %w", writeErr)
		}
	}
	// #nosec G304 -- the startup script path is fixed by MiSTer.
	data, err := os.ReadFile(filepath.Clean(paths.StartupScript))
	if err != nil {
		return fmt.Errorf("read startup script: %w", err)
	}
	if strings.Contains(string(data), startupMarker) {
		return nil
	}
	// #nosec G302,G304 -- appending to MiSTer's shared startup script.
	file, err := os.OpenFile(filepath.Clean(paths.StartupScript), os.O_APPEND|os.O_WRONLY, 0o755)
	if err != nil {
		return fmt.Errorf("open startup script: %w", err)
	}
	defer func() { _ = file.Close() }()
	command := paths.AppPath
	if strings.ContainsAny(command, " \t") {
		command = strconv.Quote(command)
	}
	entry := fmt.Sprintf("\n# %s\n[[ -e %s ]] && %s $1\n", startupMarker, command, command)
	if _, err := file.WriteString(entry); err != nil {
		return fmt.Errorf("append BGM startup entry: %w", err)
	}
	logger.Print("Added service to startup script.")
	return nil
}

// Playlists lists the playlist folders shown in the menu: every directory in
// the music folder except boot, sorted.
func Playlists(paths *Paths) ([]string, error) {
	entries, err := os.ReadDir(paths.MusicFolder)
	if err != nil {
		return nil, fmt.Errorf("list playlists: %w", err)
	}
	playlists := make([]string, 0, len(entries))
	for _, entry := range entries {
		if entry.Name() == BootFolder || !entry.IsDir() {
			continue
		}
		playlists = append(playlists, entry.Name())
	}
	sort.Strings(playlists)
	return playlists, nil
}
