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
	"os"
	"path/filepath"
	"strings"

	"github.com/wizzomafizzo/mrext/pkg/config"
	"github.com/wizzomafizzo/mrext/pkg/utils"
)

// GetSystem looks up an exact system definition by ID.
func GetSystem(id string) (*System, error) {
	if system, ok := Systems[id]; ok {
		return &system, nil
	}
	return nil, fmt.Errorf("unknown system: %s", id)
}

func GetGroup(groupID string) (System, error) {
	var merged System
	group, ok := CoreGroups[groupID]
	if !ok {
		return merged, fmt.Errorf("no system group found for %s", groupID)
	}

	if len(group) < 1 {
		return merged, fmt.Errorf("no systems in %s", groupID)
	}
	if len(group) == 1 {
		return group[0], nil
	}

	merged = group[0]
	merged.Slots = make([]Slot, 0)
	merged.extensions = make([]string, 0)
	for i := range group {
		merged.Slots = append(merged.Slots, group[i].Slots...)
		merged.extensions = append(merged.extensions, group[i].extensions...)
	}

	return merged, nil
}

// LookupSystem case-insensitively looks up system ID definition including aliases.
func LookupSystem(id string) (*System, error) {
	if system, err := GetGroup(id); err == nil {
		return &system, nil
	}

	for k := range Systems {
		system := Systems[k]
		if strings.EqualFold(k, id) {
			return &system, nil
		}

		for _, alias := range system.Alias {
			if strings.EqualFold(alias, id) {
				return &system, nil
			}
		}
	}

	return nil, fmt.Errorf("unknown system: %s", id)
}

// MatchSystemFile returns true if a given file's extension is valid for a system.
func MatchSystemFile(system *System, path string) bool {
	// ignore dot files
	if strings.HasPrefix(filepath.Base(path), ".") {
		return false
	}

	lowerPath := strings.ToLower(path)
	if len(system.extensions) > 0 {
		for _, ext := range system.extensions {
			if strings.HasSuffix(lowerPath, ext) {
				return true
			}
		}
		return false
	}

	// Preserve compatibility for callers constructing System values manually.
	for _, slot := range system.Slots {
		for _, ext := range slot.Exts {
			if strings.HasSuffix(lowerPath, ext) {
				return true
			}
		}
	}

	return false
}

func AllSystems() []System {
	keys := utils.AlphaMapKeys(Systems)
	systems := make([]System, 0, len(keys))

	for _, k := range keys {
		systems = append(systems, Systems[k])
	}

	return systems
}

// GetFiles searches for all valid games in a given path and returns a list of
// files. It deep searches .zip files and handles symlinks at all levels.
//
// This used to walk the tree itself, calling os.Chdir so relative symlink
// targets resolved. os.Chdir is process-global, and this runs inside Remote's
// daemon alongside HTTP handlers and the tracker goroutine, so a random-game
// launch could move the working directory out from under another request.
// WalkFiles was written to do the same scan without that, and
// TestWalkFilesMatchesGetFiles pinned the two to the same results.
func GetFiles(systemID, path string) ([]string, error) {
	var results []string
	if err := WalkFiles(systemID, path, func(file string) error {
		results = append(results, file)
		return nil
	}); err != nil {
		return nil, err
	}
	return results, nil
}

func FilterUniqueFilenames(files []string) []string {
	var filtered []string
	filenames := make(map[string]struct{})
	for i := range files {
		fn := filepath.Base(files[i])
		if _, ok := filenames[fn]; ok {
			continue
		}
		filenames[fn] = struct{}{}
		filtered = append(filtered, files[i])
	}
	return filtered
}

type RBFInfo struct {
	Path      string // full path to RBF file
	Filename  string // base filename of RBF file
	ShortName string // base filename without date or extension
	MGLName   string // relative path launch-able from MGL file
}

func ParseRBF(path string) RBFInfo {
	info := RBFInfo{
		Path:     path,
		Filename: filepath.Base(path),
	}

	if strings.Contains(info.Filename, "_") {
		info.ShortName = info.Filename[0:strings.LastIndex(info.Filename, "_")]
	} else {
		info.ShortName = strings.TrimSuffix(info.Filename, filepath.Ext(info.Filename))
	}

	if strings.HasPrefix(path, config.SdFolder) {
		relDir := strings.TrimPrefix(filepath.Dir(path), config.SdFolder+"/")
		info.MGLName = filepath.Join(relDir, info.ShortName)
	} else {
		info.MGLName = path
	}

	return info
}

// Find all rbf files in the top 2 menu levels of the SD card.
func shallowScanRBF() ([]RBFInfo, error) {
	results := make([]RBFInfo, 0)

	isRBF := func(file os.DirEntry) bool {
		return filepath.Ext(strings.ToLower(file.Name())) == ".rbf"
	}

	infoSymlink := func(path string) (RBFInfo, error) {
		info, err := os.Lstat(path)
		if err != nil {
			return RBFInfo{}, fmt.Errorf("inspect RBF path: %w", err)
		}

		if info.Mode()&os.ModeSymlink != 0 {
			newPath, err := os.Readlink(path)
			if err != nil {
				return RBFInfo{}, fmt.Errorf("read RBF symlink: %w", err)
			}
			return ParseRBF(newPath), nil
		}
		return ParseRBF(path), nil
	}

	files, err := os.ReadDir(config.SdFolder)
	if err != nil {
		return results, fmt.Errorf("read MiSTer root: %w", err)
	}

	for _, file := range files {
		if file.IsDir() && strings.HasPrefix(file.Name(), "_") {
			subFiles, err := os.ReadDir(filepath.Join(config.SdFolder, file.Name()))
			if err != nil {
				continue
			}

			for _, subFile := range subFiles {
				if isRBF(subFile) {
					path := filepath.Join(config.SdFolder, file.Name(), subFile.Name())
					info, err := infoSymlink(path)
					if err != nil {
						continue
					}
					results = append(results, info)
				}
			}
		} else if isRBF(file) {
			path := filepath.Join(config.SdFolder, file.Name())
			info, err := infoSymlink(path)
			if err != nil {
				continue
			}
			results = append(results, info)
		}
	}

	return results, nil
}

// SystemsWithRbf returns a map of all system IDs which have an existing rbf file.
func SystemsWithRBF() map[string]RBFInfo {
	// TODO: include alt rbfs somehow?
	results := make(map[string]RBFInfo)

	rbfFiles, err := shallowScanRBF()
	if err != nil {
		return results
	}

	for _, rbfFile := range rbfFiles {
		for id := range Systems {
			system := Systems[id]
			shortName := system.Rbf

			if strings.Contains(shortName, "/") {
				shortName = shortName[strings.LastIndex(shortName, "/")+1:]
			}

			if strings.EqualFold(rbfFile.ShortName, shortName) {
				results[system.Id] = rbfFile
			}
		}
	}

	return results
}
