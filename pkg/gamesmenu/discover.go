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

package gamesmenu

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/wizzomafizzo/mrext/pkg/config"
	"github.com/wizzomafizzo/mrext/pkg/games"
)

// Entry is one games folder offered on the main screen. The Python script
// keyed everything by games folder name, and so does this: Key is the catalog
// spelling of the folder, Display its names.txt replacement, and MenuFolder
// the _Games subfolder that receives its shortcuts.
//
//nolint:govet // Field order groups identity, filesystem locations and state.
type Entry struct {
	Key        string
	Display    string
	MenuFolder string
	Systems    []games.System
	Paths      []string
	Existing   bool
	Selected   bool
}

// Discover lists every catalog games folder that exists under a games root,
// in the roots' priority order. MenuFolder reuses the spelling of a matching
// _Games subfolder when one exists, compared case-insensitively, so existing
// installations keep their folders. Entries are selected when their folder
// exists, or when there is no _Games folder at all.
func (m *Manager) Discover() ([]Entry, error) {
	existing, err := m.menuFolders()
	if err != nil {
		return nil, err
	}
	menuExists := existing != nil
	roots := m.resolvedRoots()

	var entries []Entry
	for _, folder := range menuSystems() {
		paths := folderPaths(roots, folder.Key)
		if len(paths) == 0 {
			continue
		}
		entry := Entry{
			Key:        folder.Key,
			Display:    m.names.Replace(folder.Key),
			MenuFolder: m.names.FolderName(folder.Key),
			Systems:    folder.Systems,
			Paths:      paths,
		}
		if name, ok := existing[strings.ToLower(entry.MenuFolder)]; ok {
			entry.MenuFolder = name
			entry.Existing = true
		}
		entry.Selected = entry.Existing || !menuExists
		entries = append(entries, entry)
	}
	return entries, nil
}

type gamesRoot struct {
	entries map[string]string
	path    string
}

// resolvedRoots lists the games roots that exist, each with a case-insensitive
// index of its children, so every system folder costs one lookup instead of a
// directory read per system.
func (m *Manager) resolvedRoots() []gamesRoot {
	var roots []gamesRoot
	folders := games.GetGamesFolders(m.cfg)
	// Relocated managers never probe the device's built-in roots. Explicit
	// configured roots still work, including external fixture libraries.
	if filepath.Clean(m.paths.SDRoot) != config.SdFolder {
		folders = folders[:len(folders)-len(config.GamesFolders)]
	}
	for _, root := range folders {
		path, err := games.FindFile(root)
		if err != nil {
			continue
		}
		children, err := os.ReadDir(path)
		if err != nil {
			continue
		}
		index := make(map[string]string, len(children))
		for _, child := range children {
			lower := strings.ToLower(child.Name())
			if _, ok := index[lower]; !ok {
				index[lower] = child.Name()
			}
		}
		roots = append(roots, gamesRoot{path: path, entries: index})
	}
	return roots
}

func folderPaths(roots []gamesRoot, key string) []string {
	var paths []string
	seen := make(map[string]struct{})
	lower := strings.ToLower(key)
	for _, root := range roots {
		name, ok := root.entries[lower]
		if !ok {
			continue
		}
		path := filepath.Join(root.path, name)
		info, err := os.Stat(path)
		if err != nil || !info.IsDir() {
			continue
		}
		clean := filepath.Clean(path)
		if _, ok := seen[clean]; ok {
			continue
		}
		seen[clean] = struct{}{}
		paths = append(paths, clean)
	}
	return paths
}

// menuFolders indexes the directories directly under _Games by lower-cased
// name. It returns a nil map when _Games does not exist.
func (m *Manager) menuFolders() (map[string]string, error) {
	children, err := os.ReadDir(m.paths.MenuFolder)
	if errors.Is(err, fs.ErrNotExist) {
		return nil, nil //nolint:nilnil // A nil index distinguishes an absent menu from an empty one.
	}
	if err != nil {
		return nil, fmt.Errorf("read Games menu folder: %w", err)
	}
	folders := make(map[string]string, len(children))
	for _, child := range children {
		if !isDirectory(m.paths.MenuFolder, child) {
			continue
		}
		lower := strings.ToLower(child.Name())
		if _, ok := folders[lower]; !ok {
			folders[lower] = child.Name()
		}
	}
	return folders, nil
}

// isDirectory follows symlinks the way the Python script's os.path.isdir did.
func isDirectory(parent string, entry fs.DirEntry) bool {
	if entry.IsDir() {
		return true
	}
	if entry.Type()&fs.ModeSymlink == 0 {
		return false
	}
	info, err := os.Stat(filepath.Join(parent, entry.Name()))
	return err == nil && info.IsDir()
}
