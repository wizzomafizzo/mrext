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

package games

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/wizzomafizzo/mrext/pkg/utils"
)

// WalkFiles emits matching game paths without retaining a whole-library list
// or changing the process working directory. Directory order and ZIP member
// matching follow GetFiles; symlink paths remain usable through their aliases.
func WalkFiles(systemID, root string, visit func(string) error) error {
	system, err := GetSystem(systemID)
	if err != nil {
		return err
	}
	resolved, err := filepath.EvalSymlinks(root)
	if err != nil {
		return fmt.Errorf("resolve game root: %w", err)
	}
	info, err := os.Stat(resolved)
	if err != nil {
		return fmt.Errorf("stat game root: %w", err)
	}
	if !info.IsDir() {
		return fmt.Errorf("game root is not a directory: %s", root)
	}
	visited := make(map[string]bool)
	var walk func(string, string) error
	walk = func(realRoot, displayRoot string) error {
		err := filepath.WalkDir(realRoot, func(path string, entry fs.DirEntry, walkErr error) error {
			if walkErr != nil {
				return fmt.Errorf("scan game path: %w", walkErr)
			}
			relative, relErr := filepath.Rel(realRoot, path)
			if relErr != nil {
				return fmt.Errorf("resolve game path: %w", relErr)
			}
			display := filepath.Join(displayRoot, relative)
			if entry.IsDir() {
				if visited[path] {
					return filepath.SkipDir
				}
				visited[path] = true
				return nil
			}
			if entry.Type()&os.ModeSymlink != 0 {
				// A symlink that does not resolve points at something that is
				// not there: a drive left unplugged, or a game since deleted.
				// That is not a scan failure. Treating it as one meant a single
				// dangling link anywhere in a library aborted the whole index,
				// so nothing at all got indexed and the user had no way to see
				// which link was at fault.
				target, resolveErr := filepath.EvalSymlinks(path)
				if resolveErr != nil {
					return nil //nolint:nilerr // A broken link is an absent file, not an error.
				}
				targetInfo, statErr := os.Stat(target)
				if statErr != nil {
					return nil //nolint:nilerr // As above: unreachable target, not a scan failure.
				}
				if targetInfo.IsDir() {
					return walk(target, display)
				}
			}
			if strings.HasSuffix(strings.ToLower(path), ".zip") {
				members, zipErr := utils.ListZip(path)
				if zipErr != nil {
					return nil //nolint:nilerr // GetFiles also ignores invalid archives.
				}
				for _, member := range members {
					if MatchSystemFile(system, member) {
						if err := visit(filepath.Join(display, member)); err != nil {
							return err
						}
					}
				}
				return nil
			}
			if MatchSystemFile(system, path) {
				return visit(display)
			}
			return nil
		})
		if err != nil {
			return fmt.Errorf("walk games: %w", err)
		}
		return nil
	}
	return walk(resolved, root)
}
