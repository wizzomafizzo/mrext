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
	"os"
	"path/filepath"
	"strings"

	"github.com/wizzomafizzo/mrext/pkg/config"
)

// StorageRootFor returns the removable or network mount point a path lives
// under, or "" for paths that are not on removable storage. The longest match
// wins so /media/fat/cifs is not mistaken for the SD card.
func StorageRootFor(path string) string {
	cleaned := filepath.Clean(path)
	best := ""
	for _, root := range config.StorageRoots {
		if cleaned != root && !strings.HasPrefix(cleaned, root+string(filepath.Separator)) {
			continue
		}
		if len(root) > len(best) {
			best = root
		}
	}
	return best
}

// TargetAvailable reports whether the storage holding path is attached.
//
// A missing file on attached storage was deleted by the user. A missing file
// on detached storage is only out of reach, and treating the two the same
// destroys shortcuts and favorites that took real effort to build: a USB drive
// left unplugged, or a NAS powered off, would otherwise wipe every entry
// pointing at it. Callers must consult this before acting on os.IsNotExist.
func TargetAvailable(path string) bool {
	root := StorageRootFor(path)
	if root == "" {
		return true
	}
	entries, err := os.ReadDir(root)
	if err != nil {
		// The mount point itself is gone or unreadable.
		return false
	}
	// An unmounted mount point is an empty directory.
	return len(entries) > 0
}
