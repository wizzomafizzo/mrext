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
	"bytes"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/wizzomafizzo/mrext/pkg/mister"
	"github.com/wizzomafizzo/mrext/pkg/utils"
)

type CleanupProgress struct{ Checked, Removed int }

type CleanupResult struct {
	Errors                                                   []error
	Checked, Removed, Unreadable, PrunedFolders, Unavailable int
	ErrorsDropped                                            int
}

// errStorageUnavailable marks a shortcut whose target lives on storage that is
// not attached. The game is out of reach, not gone, so the shortcut stays.
var errStorageUnavailable = errors.New("target storage is not available")

func (r *CleanupResult) unreadable(err error) {
	r.Unreadable++
	if len(r.Errors) < 10 {
		r.Errors = append(r.Errors, err)
	} else {
		r.ErrorsDropped++
	}
}

type zipContents struct {
	members map[string]bool
	err     error
}

// CleanUp removes only proven missing targets. Unknown paths, inaccessible
// files and invalid archives remain untouched. System folders encode toggle
// state and are never pruned. Symlinks in the menu are never traversed.
func (m *Manager) CleanUp(progress func(CleanupProgress)) (CleanupResult, error) {
	var result CleanupResult
	root, menu, err := m.openSD()
	if err != nil {
		return result, err
	}
	defer func() { _ = root.Close() }()
	info, err := root.Lstat(menu)
	if errors.Is(err, fs.ErrNotExist) {
		return result, nil
	}
	if err != nil {
		return result, fmt.Errorf("stat Games menu: %w", err)
	}
	if !info.IsDir() {
		return result, errors.New("games menu is not a regular directory")
	}
	archives := make(map[string]zipContents)
	var dirs []string
	err = fs.WalkDir(root.FS(), menu, func(name string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			result.unreadable(fmt.Errorf("read menu path %s: %w", name, walkErr))
			return nil
		}
		if entry.IsDir() {
			if name != menu && path.Dir(name) != menu {
				dirs = append(dirs, name)
			}
			return nil
		}
		if !entry.Type().IsRegular() || !strings.EqualFold(path.Ext(name), ".mgl") {
			return nil
		}
		result.Checked++
		missing, checkErr := checkShortcut(root, name, archives)
		switch {
		case errors.Is(checkErr, errStorageUnavailable):
			// A drive that is merely unplugged must not cost the user their
			// shortcuts. Count it so the summary can say what was skipped.
			result.Unavailable++
		case checkErr != nil:
			result.unreadable(fmt.Errorf("check shortcut %s: %w", name, checkErr))
		case missing:
			if removeErr := root.Remove(name); removeErr != nil {
				result.unreadable(fmt.Errorf("remove shortcut %s: %w", name, removeErr))
			} else {
				result.Removed++
			}
		}
		if progress != nil {
			progress(CleanupProgress{Checked: result.Checked, Removed: result.Removed})
		}
		return nil
	})
	if err != nil {
		return result, fmt.Errorf("walk Games menu: %w", err)
	}
	for index := len(dirs) - 1; index >= 0; index-- {
		children, readErr := fs.ReadDir(root.FS(), dirs[index])
		if readErr != nil {
			result.unreadable(fmt.Errorf("read menu folder %s: %w", dirs[index], readErr))
			continue
		}
		if len(children) != 0 {
			continue
		}
		if removeErr := root.Remove(dirs[index]); removeErr != nil {
			result.unreadable(fmt.Errorf("prune menu folder %s: %w", dirs[index], removeErr))
		} else {
			result.PrunedFolders++
		}
	}
	return result, nil
}

func checkShortcut(root *os.Root, name string, archives map[string]zipContents) (bool, error) {
	data, err := root.ReadFile(name)
	if err != nil {
		return false, fmt.Errorf("read MGL: %w", err)
	}
	// ReadMGL only returns the first file; cleanup must validate every target.
	// Strict=false preserves the Python writer's legacy raw ampersands.
	type mglFile struct {
		Path string `xml:"path,attr"`
	}
	var mgl struct {
		XMLName xml.Name  `xml:"mistergamedescription"`
		Files   []mglFile `xml:"file"`
	}
	decoder := xml.NewDecoder(bytes.NewReader(data))
	decoder.Strict = false
	if err := decoder.Decode(&mgl); err != nil {
		return false, fmt.Errorf("decode MGL: %w", err)
	}
	for {
		token, tokenErr := decoder.Token()
		if errors.Is(tokenErr, io.EOF) {
			break
		}
		if tokenErr != nil {
			return false, fmt.Errorf("decode MGL ending: %w", tokenErr)
		}
		switch value := token.(type) {
		case xml.Comment, xml.ProcInst:
		case xml.CharData:
			if strings.TrimSpace(string(value)) != "" {
				return false, errors.New("unexpected text after MGL")
			}
		default:
			return false, errors.New("unexpected content after MGL")
		}
	}
	missing := false
	for _, file := range mgl.Files {
		if file.Path == "" {
			return false, errors.New("MGL file has no target")
		}
		target := file.Path
		const prefix = "../../../../.."
		if strings.HasPrefix(target, prefix+"/") {
			target = strings.TrimPrefix(target, prefix)
		}
		if !filepath.IsAbs(target) {
			return false, fmt.Errorf("ambiguous MGL target: %q", file.Path)
		}
		absent, err := missingTarget(target, archives)
		if err != nil {
			return false, err
		}
		missing = missing || absent
	}
	return missing, nil
}

func missingTarget(target string, archives map[string]zipContents) (bool, error) {
	if index := strings.Index(strings.ToLower(target), ".zip/"); index >= 0 {
		archive, member := target[:index+4], target[index+5:]
		if filepath.Clean(archive) != archive || !validZIPMember(member) {
			return false, fmt.Errorf("ambiguous ZIP target: %q", target)
		}
		contents, ok := archives[archive]
		if !ok {
			info, err := os.Stat(archive)
			if errors.Is(err, fs.ErrNotExist) {
				if !mister.TargetAvailable(archive) {
					return false, fmt.Errorf("%w: %s", errStorageUnavailable, archive)
				}
				return true, nil
			}
			if err != nil {
				return false, fmt.Errorf("stat archive: %w", err)
			}
			if !info.Mode().IsRegular() {
				return false, fmt.Errorf("archive is not a regular file: %s", archive)
			}
			names, err := utils.ListZip(archive)
			contents = zipContents{members: make(map[string]bool, len(names)), err: err}
			for _, name := range names {
				contents.members[name] = true
			}
			archives[archive] = contents
		}
		if contents.err != nil {
			return false, fmt.Errorf("read archive: %w", contents.err)
		}
		return !contents.members[member], nil
	}
	// Only ZIP member keys may contain harmless noncanonical components.
	// Ordinary filesystem targets must stay unambiguous before deletion.
	if filepath.Clean(target) != target {
		return false, fmt.Errorf("ambiguous MGL target: %q", target)
	}
	_, err := os.Stat(target)
	if errors.Is(err, fs.ErrNotExist) {
		if !mister.TargetAvailable(target) {
			return false, fmt.Errorf("%w: %s", errStorageUnavailable, target)
		}
		return true, nil
	}
	if err != nil {
		return false, fmt.Errorf("stat target: %w", err)
	}
	return false, nil
}
