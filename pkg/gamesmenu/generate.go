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
	"path"
	"path/filepath"
	"sort"
	"strings"

	"github.com/wizzomafizzo/mrext/pkg/games"
	"github.com/wizzomafizzo/mrext/pkg/mister"
	"github.com/wizzomafizzo/mrext/pkg/utils"
)

type Progress struct {
	Entry                 *Entry
	Folder                string
	Index, Total, Created int
}

type ScanResult struct {
	Errors                                  []error
	Created, Skipped, Failed, ErrorsDropped int
}

func (r *ScanResult) fail(err error) {
	r.Failed++
	if len(r.Errors) < 10 {
		r.Errors = append(r.Errors, err)
	} else {
		r.ErrorsDropped++
	}
}

type GenerateResult struct {
	Removed []string
	Scan    ScanResult
}

// openSD confines every menu mutation, including symlinks, to the SD root.
func (m *Manager) openSD() (*os.Root, string, error) {
	relative, err := filepath.Rel(m.paths.SDRoot, m.paths.MenuFolder)
	if err != nil || !filepath.IsLocal(relative) || relative == "." {
		return nil, "", errors.New("games menu must be below the SD root")
	}
	root, err := os.OpenRoot(m.paths.SDRoot)
	if err != nil {
		return nil, "", fmt.Errorf("open SD root: %w", err)
	}
	return root, relative, nil
}

func (m *Manager) DeselectedFolders(selected []Entry) ([]string, error) {
	folders, err := m.menuFolders()
	if err != nil {
		return nil, err
	}
	keep := make(map[string]bool, len(selected))
	for _, entry := range selected {
		keep[strings.ToLower(entry.MenuFolder)] = true
	}
	var removed []string
	for lower, name := range folders {
		if !keep[lower] {
			removed = append(removed, name)
		}
	}
	sort.Strings(removed)
	return removed, nil
}

func (m *Manager) RemoveAll() error {
	root, menu, err := m.openSD()
	if err != nil {
		return err
	}
	defer func() { _ = root.Close() }()
	if err := root.RemoveAll(menu); err != nil {
		return fmt.Errorf("remove Games menu: %w", err)
	}
	return nil
}

func (m *Manager) Generate(selected []Entry, progress func(Progress)) (GenerateResult, error) {
	var result GenerateResult
	for _, entry := range selected {
		if !filepath.IsLocal(entry.MenuFolder) || filepath.Base(entry.MenuFolder) != entry.MenuFolder ||
			entry.MenuFolder == "." {
			return result, fmt.Errorf("invalid menu folder: %q", entry.MenuFolder)
		}
	}
	removed, err := m.DeselectedFolders(selected)
	if err != nil {
		return result, err
	}
	root, menu, err := m.openSD()
	if err != nil {
		return result, err
	}
	defer func() { _ = root.Close() }()
	// #nosec G301 -- MiSTer menu folders are shared filesystem data.
	if err := root.MkdirAll(menu, 0o755); err != nil {
		return result, fmt.Errorf("create Games menu: %w", err)
	}
	for _, name := range removed {
		if err := root.RemoveAll(filepath.Join(menu, name)); err != nil {
			return result, fmt.Errorf("remove deselected folder %s: %w", name, err)
		}
		result.Removed = append(result.Removed, name)
	}
	scan := menuScanner{
		manager: m, root: root, menu: menu, result: &result.Scan,
		directories: make(map[string]map[string]bool),
	}
	for index := range selected {
		entry := &selected[index]
		scan.entry, scan.rules = entry, extensionRules(entry.Systems)
		scan.prepareNeoGeo()
		scan.visited = make(map[string]bool)
		for _, folder := range entry.Paths {
			scan.folder = folder
			scan.notify = func() {
				if progress != nil {
					progress(Progress{
						Entry: entry, Folder: folder, Index: index, Total: len(selected), Created: result.Scan.Created,
					})
				}
			}
			scan.notify()
			scan.walk(folder)
		}
	}
	if progress != nil {
		progress(Progress{Index: len(selected), Total: len(selected), Created: result.Scan.Created})
	}
	return result, nil
}

//nolint:govet // Groups scan state and filesystem caches.
type menuScanner struct {
	manager     *Manager
	root        *os.Root
	menu        string
	result      *ScanResult
	entry       *Entry
	neoGeo      *games.System
	neoGeoNames map[string]string
	rules       []extensionRule
	folder      string
	visited     map[string]bool
	directories map[string]map[string]bool
	notify      func()
	processed   int
}

func (s *menuScanner) walk(dir string) {
	resolved, err := filepath.EvalSymlinks(dir)
	if err != nil {
		s.result.fail(fmt.Errorf("resolve games folder %s: %w", dir, err))
		return
	}
	if s.visited[resolved] {
		return
	}
	s.visited[resolved] = true
	children, err := os.ReadDir(dir)
	if err != nil {
		s.result.fail(fmt.Errorf("read games folder %s: %w", dir, err))
		return
	}
	var directories []string
	for _, child := range children {
		if strings.HasPrefix(child.Name(), ".") {
			continue
		}
		full := filepath.Join(dir, child.Name())
		info, err := os.Stat(full)
		if err != nil {
			s.result.fail(fmt.Errorf("stat game %s: %w", full, err))
			continue
		}
		if !info.IsDir() && !info.Mode().IsRegular() {
			continue
		}
		relative, err := filepath.Rel(s.folder, dir)
		if err != nil {
			s.result.fail(fmt.Errorf("resolve game folder: %w", err))
			continue
		}
		if s.writeNeoGeoSet(full, relative, info.IsDir()) {
			continue
		}
		if info.IsDir() {
			directories = append(directories, full)
			continue
		}
		if strings.EqualFold(filepath.Ext(full), ".zip") {
			s.archive(full, relative)
		} else {
			s.write(full, relativeFolders(relative), child.Name())
		}
	}
	// Python's os.walk yields current-directory files before its subfolders.
	// Parent ZIP members must keep priority over colliding loose files below.
	for _, directory := range directories {
		s.walk(directory)
	}
}

func relativeFolders(relative string) []string {
	if relative == "." || relative == "" {
		return nil
	}
	return strings.Split(filepath.ToSlash(relative), "/")
}

// ZIP member names are archive keys, not filesystem paths. Allow harmless dot
// and empty components, but reject traversal before any path normalization.
func validZIPMember(member string) bool {
	if member == "" || strings.HasPrefix(member, "/") ||
		strings.Contains(member, "\\") || strings.ContainsRune(member, 0) {
		return false
	}
	for _, component := range strings.Split(member, "/") {
		if component == ".." {
			return false
		}
	}
	clean := path.Clean(member)
	return clean != "." && fs.ValidPath(clean)
}

func hiddenMember(member string) bool {
	for _, component := range strings.Split(member, "/") {
		if component != "." && strings.HasPrefix(component, ".") {
			return true
		}
	}
	return false
}

func (s *menuScanner) archive(filename, relative string) {
	members, err := utils.ListZip(filename)
	if err != nil {
		return
	} // Python silently ignored invalid ZIPs.
	for _, member := range members {
		if strings.HasSuffix(member, "/") {
			continue
		}
		if !validZIPMember(member) {
			s.result.fail(fmt.Errorf("unsafe ZIP member in %s: %q", filename, member))
			continue
		}
		if hiddenMember(member) {
			continue
		}
		folders := relativeFolders(relative)
		if index := strings.LastIndex(member, "/"); index >= 0 {
			// Match Python's dirname: trim trailing separators, but preserve
			// interior empty/dot components until names.txt prefixes them.
			directory := strings.TrimRight(member[:index], "/")
			if directory != "" {
				folders = append(folders, strings.Split(directory, "/")...)
			}
		}
		s.write(filename+"/"+member, folders, path.Base(member))
	}
}

func (s *menuScanner) write(full string, folders []string, name string) {
	system := matchSystem(s.rules, name)
	if system == nil {
		return
	}
	s.writeShortcut(full, folders, strings.TrimSuffix(name, filepath.Ext(name))+".mgl", system, "")
}

func (s *menuScanner) writeShortcut(
	full string, folders []string, filename string, system *games.System, override string,
) {
	s.processed++
	if s.processed%250 == 0 {
		s.notify()
	}
	components := make([]string, 0, 2+len(folders))
	components = append(components, s.menu, s.entry.MenuFolder)
	for _, component := range folders {
		components = append(components, s.manager.names.FolderName(component))
	}
	dir := filepath.Join(components...)
	existing, err := s.outputDirectory(dir)
	if err != nil {
		s.result.fail(err)
		return
	}
	key := strings.ToLower(filename)
	if existing[key] {
		s.result.Skipped++
		return
	}
	data, err := mister.GenerateMgl(s.manager.cfg, system, full, override)
	if err != nil {
		s.result.fail(fmt.Errorf("generate shortcut for %s: %w", full, err))
		return
	}
	target := filepath.Join(dir, filename)
	// O_EXCL also protects against another writer or a dangling symlink.
	// #nosec G302 -- shortcuts are shared MiSTer menu data.
	file, err := s.root.OpenFile(target, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o644)
	if errors.Is(err, fs.ErrExist) {
		existing[key] = true
		s.result.Skipped++
		return
	}
	if err != nil {
		s.result.fail(fmt.Errorf("create shortcut %s: %w", target, err))
		return
	}
	_, writeErr := file.WriteString(data)
	closeErr := file.Close()
	if err := errors.Join(writeErr, closeErr); err != nil {
		_ = s.root.Remove(target)
		s.result.fail(fmt.Errorf("write shortcut %s: %w", target, err))
		return
	}
	existing[key] = true
	s.result.Created++
}

func (s *menuScanner) outputDirectory(dir string) (map[string]bool, error) {
	if existing, ok := s.directories[dir]; ok {
		return existing, nil
	}
	// #nosec G301 -- MiSTer menu folders are shared filesystem data.
	if err := s.root.MkdirAll(dir, 0o755); err != nil {
		return nil, fmt.Errorf("create shortcut folder %s: %w", dir, err)
	}
	children, err := fs.ReadDir(s.root.FS(), dir)
	if err != nil {
		return nil, fmt.Errorf("read shortcut folder %s: %w", dir, err)
	}
	existing := make(map[string]bool, len(children))
	for _, child := range children {
		existing[strings.ToLower(child.Name())] = true
	}
	s.directories[dir] = existing
	return existing, nil
}
