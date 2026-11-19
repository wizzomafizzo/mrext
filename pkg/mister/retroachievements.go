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

package mister

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// Main of the RetroAchievements build reads retroachievements.cfg with fgets
// into a 512-byte buffer, so a longer line is parsed as several.
const raLineSize = 511

// RAHardcore is the hardcore part of retroachievements.cfg.
type RAHardcore struct {
	// Hardcore is the hardcore= key: the mode each core load starts in.
	Hardcore bool
	// Forced is the force_hardcore= key, which turns hardcore on for every
	// core and overrides Hardcore. mrext reports it and never writes it.
	Forced bool
}

// raLine is one line of retroachievements.cfg as Main splits it, with the
// byte offsets of its value inside the file.
type raLine struct {
	key        string
	value      string
	valueStart int
	valueEnd   int
}

// raLines splits a retroachievements.cfg the way Main's parser does: lines
// of at most raLineSize bytes, cut at the first CR, with "#" comments only in
// column 0, the key untrimmed and the value trimmed of leading spaces and
// tabs. Lines without "=" are skipped.
func raLines(data []byte) []raLine {
	var lines []raLine
	for start := 0; start < len(data); {
		end := start + raLineSize
		if end > len(data) {
			end = len(data)
		}
		if newline := bytes.IndexByte(data[start:end], '\n'); newline >= 0 {
			end = start + newline + 1
		}
		line := data[start:end]
		next := end

		// Main cuts at the first LF, then at the first CR.
		if i := bytes.IndexByte(line, '\n'); i >= 0 {
			line = line[:i]
		}
		if i := bytes.IndexByte(line, '\r'); i >= 0 {
			line = line[:i]
		}
		if len(line) > 0 && line[0] != '#' {
			if eq := bytes.IndexByte(line, '='); eq >= 0 {
				valueStart := eq + 1
				for valueStart < len(line) && (line[valueStart] == ' ' || line[valueStart] == '\t') {
					valueStart++
				}
				lines = append(lines, raLine{
					key:        string(line[:eq]),
					value:      string(line[valueStart:]),
					valueStart: start + valueStart,
					valueEnd:   start + len(line),
				})
			}
		}
		start = next
	}
	return lines
}

// raOn reads a value the way Main does, with C atoi, and reports whether it
// is non-zero: leading white space, an optional sign, then digits up to the
// first other byte. Any digit other than 0 makes it on.
func raOn(s string) bool {
	i := 0
	for i < len(s) && strings.IndexByte(" \t\n\v\f\r", s[i]) >= 0 {
		i++
	}
	if i < len(s) && (s[i] == '+' || s[i] == '-') {
		i++
	}
	for ; i < len(s) && s[i] >= '0' && s[i] <= '9'; i++ {
		if s[i] != '0' {
			return true
		}
	}
	return false
}

// parseRAHardcore reads the hardcore keys as Main does: keys are matched
// without regard to case, the last one wins, and any non-zero number is on.
// A missing key is off, Main's default.
func parseRAHardcore(data []byte) RAHardcore {
	var settings RAHardcore
	for _, line := range raLines(data) {
		switch {
		case strings.EqualFold(line.key, "hardcore"):
			settings.Hardcore = raOn(line.value)
		case strings.EqualFold(line.key, "force_hardcore"):
			settings.Forced = raOn(line.value)
		}
	}
	return settings
}

// ReadRAHardcore reads the hardcore keys from a retroachievements.cfg. A
// missing file returns an error matching os.ErrNotExist.
func ReadRAHardcore(path string) (RAHardcore, error) {
	// #nosec G304 -- path is the fixed RetroAchievements settings path or a test fixture.
	data, err := os.ReadFile(path)
	if err != nil {
		return RAHardcore{}, fmt.Errorf("read RetroAchievements settings: %w", err)
	}
	return parseRAHardcore(data), nil
}

// raDigits returns the offsets of the digits raOn reads in data[start:end],
// from first up to last, or start, start when there are none.
func raDigits(data []byte, start, end int) (first, last int) {
	i := start
	for i < end && strings.IndexByte(" \t\n\v\f\r", data[i]) >= 0 {
		i++
	}
	if i < end && (data[i] == '+' || data[i] == '-') {
		i++
	}
	first = i
	for i < end && data[i] >= '0' && data[i] <= '9' {
		i++
	}
	if first == i {
		return start, start
	}
	return first, i
}

// withRAHardcore returns data with the hardcore key set. In the line Main
// would use, only the digits Main reads are rewritten, in place and at the
// same length, so a line longer than Main's buffer splits where it did.
// With no such line, or one without digits, a line is appended, using the
// file's own line ending; Main uses the last. Every other byte is kept.
func withRAHardcore(data []byte, enabled bool) []byte {
	var last *raLine
	for _, line := range raLines(data) {
		if strings.EqualFold(line.key, "hardcore") {
			current := line
			last = &current
		}
	}
	if last != nil {
		if first, end := raDigits(data, last.valueStart, last.valueEnd); first < end {
			out := bytes.Clone(data)
			for i := first; i < end; i++ {
				out[i] = '0'
			}
			if enabled {
				out[first] = '1'
			}
			return out
		}
	}

	value := "0"
	if enabled {
		value = "1"
	}

	newline := "\n"
	if bytes.Contains(data, []byte("\r\n")) {
		newline = "\r\n"
	}
	out := make([]byte, 0, len(data)+len(newline)*2+len("hardcore=0"))
	out = append(out, data...)
	if len(out) > 0 && out[len(out)-1] != '\n' {
		out = append(out, newline...)
	}
	out = append(out, "hardcore="+value+newline...)
	return out
}

// SetRAHardcore sets the hardcore key of an existing retroachievements.cfg,
// keeping the rest of the file as it is. It never creates the file: without
// a login the setting is of no use. Nothing is written when Main would
// already read the asked value, so changed reports whether the file was
// touched.
func SetRAHardcore(path string, enabled bool) (changed bool, err error) {
	// #nosec G304 -- path is the fixed RetroAchievements settings path or a test fixture.
	data, err := os.ReadFile(path)
	if err != nil {
		return false, fmt.Errorf("read RetroAchievements settings: %w", err)
	}
	// A missing key is already off, so turning it off adds nothing.
	if parseRAHardcore(data).Hardcore == enabled {
		return false, nil
	}
	if err := writeRAFile(path, withRAHardcore(data, enabled)); err != nil {
		return false, err
	}
	return true, nil
}

// writeRAFile stages the new settings beside the old ones and renames them
// over, so a failed write never leaves the login half written.
func writeRAFile(path string, content []byte) error {
	temporary, err := os.CreateTemp(filepath.Dir(path), "."+filepath.Base(path)+"-*")
	if err != nil {
		return fmt.Errorf("create staged RetroAchievements settings: %w", err)
	}
	staged := temporary.Name()
	published := false
	defer func() {
		_ = temporary.Close()
		if !published {
			_ = os.Remove(staged)
		}
	}()

	if _, err := temporary.Write(content); err != nil {
		return fmt.Errorf("write staged RetroAchievements settings: %w", err)
	}
	if err := temporary.Sync(); err != nil {
		return fmt.Errorf("sync staged RetroAchievements settings: %w", err)
	}
	if err := temporary.Close(); err != nil {
		return fmt.Errorf("close staged RetroAchievements settings: %w", err)
	}
	// Best effort: the SD card is exFAT, where the mode comes from the mount
	// and chmod reports EPERM.
	// #nosec G302 -- Main reads this file as root; it keeps the usual mode.
	_ = os.Chmod(staged, 0o644)

	// #nosec G703 -- path is the fixed RetroAchievements settings path or a test fixture.
	if err := os.Rename(staged, path); err != nil {
		return fmt.Errorf("replace RetroAchievements settings: %w", err)
	}
	published = true
	return nil
}
