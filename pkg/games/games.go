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
	"errors"
	"fmt"
	"io/fs"
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

type resultsStack [][]string

func (r *resultsStack) new() {
	*r = append(*r, []string{})
}

func (r *resultsStack) pop() {
	if len(*r) == 0 {
		return
	}
	*r = (*r)[:len(*r)-1]
}

func (r *resultsStack) get() (*[]string, error) {
	if len(*r) == 0 {
		return nil, errors.New("nothing on stack")
	}
	return &(*r)[len(*r)-1], nil
}

// GetFiles searches for all valid games in a given path and return a list of
// files. This function deep searches .zip files and handles symlinks at all
// levels.
func GetFiles(systemID, path string) ([]string, error) {
	var allResults []string
	var stack resultsStack
	visited := make(map[string]struct{})

	system, err := GetSystem(systemID)
	if err != nil {
		return nil, err
	}

	cwd, err := os.Getwd()
	if err != nil {
		return nil, fmt.Errorf("get working directory: %w", err)
	}
	defer func() { _ = os.Chdir(cwd) }()

	var scanner func(path string, file fs.DirEntry, err error) error
	scanner = func(path string, file fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return fmt.Errorf("scan game path: %w", walkErr)
		}
		// avoid recursive symlinks
		if file.IsDir() {
			if _, ok := visited[path]; ok {
				return filepath.SkipDir
			}
			visited[path] = struct{}{}
		}

		// handle symlinked directories
		if file.Type()&os.ModeSymlink != 0 {
			err = os.Chdir(filepath.Dir(path))
			if err != nil {
				return fmt.Errorf("enter symlink parent directory: %w", err)
			}

			realPath, resolveErr := filepath.EvalSymlinks(path)
			if resolveErr != nil {
				return fmt.Errorf("resolve game symlink: %w", resolveErr)
			}

			file, statErr := os.Stat(realPath)
			if statErr != nil {
				return fmt.Errorf("stat game symlink target: %w", statErr)
			}

			if file.IsDir() {
				err = os.Chdir(path)
				if err != nil {
					return fmt.Errorf("enter symlinked game directory: %w", err)
				}

				stack.new()
				defer stack.pop()

				err = filepath.WalkDir(realPath, scanner)
				if err != nil {
					return fmt.Errorf("scan symlinked game directory: %w", err)
				}

				results, stackErr := stack.get()
				if stackErr != nil {
					return stackErr
				}

				for i := range *results {
					allResults = append(allResults, strings.Replace((*results)[i], realPath, path, 1))
				}

				return nil
			}
		}

		results, stackErr := stack.get()
		if stackErr != nil {
			return stackErr
		}

		if strings.HasSuffix(strings.ToLower(path), ".zip") {
			// zip files
			zipFiles, zipErr := utils.ListZip(path)
			if zipErr != nil {
				// skip invalid zip files
				return nil
			}

			for i := range zipFiles {
				if MatchSystemFile(system, zipFiles[i]) {
					abs := filepath.Join(path, zipFiles[i])
					*results = append(*results, abs)
				}
			}
		} else if MatchSystemFile(system, path) {
			// regular files
			*results = append(*results, path)
		}

		return nil
	}

	stack.new()
	defer stack.pop()

	root, err := os.Lstat(path)
	if err != nil {
		return nil, fmt.Errorf("inspect game root: %w", err)
	}

	err = os.Chdir(filepath.Dir(path))
	if err != nil {
		return nil, fmt.Errorf("enter game root parent: %w", err)
	}

	// handle symlinks on root game folder because WalkDir fails silently on them
	var realPath string
	if root.Mode()&os.ModeSymlink == 0 {
		realPath = path
	} else {
		realPath, err = filepath.EvalSymlinks(path)
		if err != nil {
			return nil, fmt.Errorf("resolve game root: %w", err)
		}
	}

	realRoot, err := os.Stat(realPath)
	if err != nil {
		return nil, fmt.Errorf("stat game root: %w", err)
	}

	if !realRoot.IsDir() {
		return nil, errors.New("root is not a directory")
	}

	err = filepath.WalkDir(realPath, scanner)
	if err != nil {
		return nil, fmt.Errorf("scan game root: %w", err)
	}

	results, err := stack.get()
	if err != nil {
		return nil, err
	}

	allResults = append(allResults, *results...)

	// change root back to symlink
	if realPath != path {
		for i := range allResults {
			allResults[i] = strings.Replace(allResults[i], realPath, path, 1)
		}
	}

	err = os.Chdir(cwd)
	if err != nil {
		return nil, fmt.Errorf("restore working directory: %w", err)
	}

	return allResults, nil
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
