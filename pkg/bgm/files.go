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
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
)

// MIDIPort is the ALSA sequencer port passed to aplaymidi.
const MIDIPort = "128:0"

func hasSuffixFold(name, suffix string) bool {
	return strings.HasSuffix(strings.ToLower(name), suffix)
}

// IsMP3 reports an .mp3 file.
func IsMP3(name string) bool { return hasSuffixFold(name, ".mp3") }

// IsPLS reports a .pls internet radio playlist.
func IsPLS(name string) bool { return hasSuffixFold(name, ".pls") }

// IsOGG reports an .ogg file.
func IsOGG(name string) bool { return hasSuffixFold(name, ".ogg") }

// IsWAV reports a .wav file.
func IsWAV(name string) bool { return hasSuffixFold(name, ".wav") }

// IsMID reports a .mid file.
func IsMID(name string) bool { return hasSuffixFold(name, ".mid") }

// IsVGM reports a .vgm, .vgz or .vgm.gz file.
func IsVGM(name string) bool {
	return hasSuffixFold(name, ".vgm") || hasSuffixFold(name, ".vgz") || hasSuffixFold(name, ".vgm.gz")
}

// IsValidFile reports whether BGM can play the named file.
func IsValidFile(name string) bool {
	return IsMP3(name) || IsOGG(name) || IsWAV(name) || IsMID(name) || IsVGM(name) || IsPLS(name)
}

var loopPattern = regexp.MustCompile(`^X(\d{2})_`)

// LoopAmount returns the X##_ repeat count from a filename, or 1. X00_ plays
// nothing, as in the Python script.
func LoopAmount(path string) int {
	match := loopPattern.FindStringSubmatch(filepath.Base(path))
	if match == nil {
		return 1
	}
	loops, err := strconv.Atoi(match[1])
	if err != nil {
		return 1
	}
	return loops
}

var plsURLPattern = regexp.MustCompile(`https?:.+`)

// PLSURL extracts the first stream URL from a .pls file, or "" when none is
// found.
func PLSURL(path string, logger *Logger) string {
	// #nosec G304 -- the path comes from the user's music folder listing.
	data, err := os.ReadFile(filepath.Clean(path))
	if err != nil {
		logger.Log("Playlist URL not found")
		return ""
	}
	if url := plsURLPattern.FindString(string(data)); url != "" {
		return url
	}
	logger.Log("Playlist URL not found")
	return ""
}
