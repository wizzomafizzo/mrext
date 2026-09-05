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

package favorites

import (
	"archive/zip"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/wizzomafizzo/mrext/pkg/games"
)

var allowedRootEntries = map[string]struct{}{
	"_arcade": {}, "_console": {}, "_computer": {}, "_dos games": {},
	"_games": {}, "_other": {}, "_utility": {}, "_llapi": {},
	"_ycarcade": {}, "cifs": {}, "games": {},
}

type BrowseEntry struct {
	System      *games.System
	Path        string
	Name        string
	Label       string
	IsDirectory bool
	IsArchive   bool
}

func (m *Manager) ExternalFolder() string {
	folder := filepath.Clean(m.cfg.Favorites.ExternalFolder)
	gamesFolder := filepath.Join(folder, "games")
	if info, err := os.Stat(gamesFolder); err == nil && info.IsDir() {
		return gamesFolder
	}
	return folder
}

func (m *Manager) ListDirectory(folder string) ([]BrowseEntry, error) {
	archivePath, archiveFolder, inArchive := SplitArchivePath(folder)
	if inArchive {
		return m.listArchive(archivePath, archiveFolder)
	}
	entries, err := os.ReadDir(folder)
	if err != nil {
		return nil, fmt.Errorf("read browser folder: %w", err)
	}
	results := make([]BrowseEntry, 0, len(entries))
	for _, entry := range entries {
		if m.cfg.Favorites.HideRootFiles && filepath.Clean(folder) == filepath.Clean(m.paths.SDRoot) {
			if _, allowed := allowedRootEntries[strings.ToLower(entry.Name())]; !allowed {
				continue
			}
		}
		path := filepath.Join(folder, entry.Name())
		if entry.IsDir() || isDirectorySymlink(path, entry) {
			results = append(results, BrowseEntry{
				Path: path, Name: entry.Name(), Label: entry.Name() + "/", IsDirectory: true,
			})
			continue
		}

		extension := strings.ToLower(filepath.Ext(entry.Name()))
		if extension == ".zip" {
			if system := m.matchMediaSystem(path); system != nil && system.Id == "NeoGeo" {
				results = append(results, BrowseEntry{
					System: system, Path: path, Name: entry.Name(),
					Label: m.browserLabel(system, path), IsArchive: true,
				})
			} else if isZip(path) {
				results = append(results, BrowseEntry{
					Path: path, Name: entry.Name(), Label: entry.Name() + "/", IsDirectory: true, IsArchive: true,
				})
			}
			continue
		}

		if extension == ".rbf" || extension == ".mra" || extension == ".mgl" {
			results = append(results, BrowseEntry{Path: path, Name: entry.Name(), Label: entry.Name()})
			continue
		}
		if system := m.matchMediaSystem(path); system != nil {
			results = append(results, BrowseEntry{
				System: system, Path: path, Name: entry.Name(), Label: m.browserLabel(system, path),
			})
		}
	}
	sortBrowseEntries(results)
	return results, nil
}

func isDirectorySymlink(path string, entry os.DirEntry) bool {
	if entry.Type()&os.ModeSymlink == 0 {
		return false
	}
	info, err := os.Stat(path)
	return err == nil && info.IsDir()
}

func sortBrowseEntries(entries []BrowseEntry) {
	sort.SliceStable(entries, func(i, j int) bool {
		if entries[i].IsDirectory != entries[j].IsDirectory {
			return entries[i].IsDirectory
		}
		return strings.ToLower(entries[i].Label) < strings.ToLower(entries[j].Label)
	})
}

func SplitArchivePath(path string) (archivePath, innerPath string, ok bool) {
	lowerPath := strings.ToLower(filepath.ToSlash(path))
	index := strings.Index(lowerPath, ".zip/")
	if index < 0 {
		return "", "", false
	}
	archiveEnd := index + len(".zip")
	normalized := filepath.ToSlash(path)
	return filepath.FromSlash(normalized[:archiveEnd]), strings.Trim(normalized[archiveEnd+1:], "/"), true
}

func isZip(path string) bool {
	reader, err := zip.OpenReader(path)
	if err != nil {
		return false
	}
	_ = reader.Close()
	return true
}

func (m *Manager) listArchive(archivePath, folder string) ([]BrowseEntry, error) {
	cacheKey := filepath.Join(archivePath, filepath.FromSlash(folder))
	if entries, found := m.archiveEntriesCache[cacheKey]; found {
		return append([]BrowseEntry(nil), entries...), nil
	}
	reader, err := zip.OpenReader(archivePath)
	if err != nil {
		return nil, fmt.Errorf("open ZIP archive: %w", err)
	}
	defer func() { _ = reader.Close() }()

	prefix := strings.Trim(folder, "/")
	if prefix != "" {
		prefix += "/"
	}
	seen := make(map[string]struct{})
	results := make([]BrowseEntry, 0)
	for _, file := range reader.File {
		name := strings.TrimPrefix(file.Name, "/")
		if !strings.HasPrefix(name, prefix) || name == prefix {
			continue
		}
		remainder := strings.TrimPrefix(name, prefix)
		parts := strings.SplitN(remainder, "/", 2)
		child := parts[0]
		if child == "" {
			continue
		}
		if _, found := seen[child]; found {
			continue
		}
		seen[child] = struct{}{}
		virtualPath := filepath.Join(archivePath, filepath.FromSlash(prefix), child)
		if len(parts) == 2 || file.FileInfo().IsDir() {
			results = append(results, BrowseEntry{
				Path: virtualPath, Name: child, Label: child + "/", IsDirectory: true, IsArchive: true,
			})
			continue
		}
		system := m.matchMediaSystem(virtualPath)
		if system == nil {
			continue
		}
		results = append(results, BrowseEntry{
			System: system, Path: virtualPath, Name: child, Label: m.browserLabel(system, virtualPath),
			IsArchive: strings.EqualFold(filepath.Ext(child), ".zip"),
		})
	}
	sortBrowseEntries(results)
	m.archiveEntriesCache[cacheKey] = append([]BrowseEntry(nil), results...)
	return results, nil
}

func (m *Manager) ParentFolder(path string) (string, error) {
	archivePath, innerPath, inArchive := SplitArchivePath(path)
	if !inArchive {
		parent := filepath.Dir(path)
		if filepath.Clean(path) == filepath.Clean(filepath.Dir(m.paths.SDRoot)) {
			return path, errors.New("already at browser root")
		}
		return parent, nil
	}
	if innerPath == "" {
		return filepath.Dir(archivePath), nil
	}
	parent := filepath.Dir(innerPath)
	if parent == "." {
		return archivePath + string(filepath.Separator), nil
	}
	return filepath.Join(archivePath, parent), nil
}

func (m *Manager) matchMediaSystem(path string) *games.System {
	systems := games.FolderToSystems(m.cfg, path)
	matches := make([]games.System, 0, len(systems))
	for index := range systems {
		if games.MatchSystemFile(&systems[index], path) {
			matches = append(matches, systems[index])
		}
	}
	for index := range matches {
		if matches[index].SetName != "" {
			return &matches[index]
		}
	}
	if len(matches) > 0 {
		return &matches[0]
	}
	if strings.EqualFold(filepath.Ext(path), ".zip") {
		for index := range systems {
			if systems[index].Id == "NeoGeo" {
				return &systems[index]
			}
		}
	}
	return nil
}

func (m *Manager) browserLabel(system *games.System, path string) string {
	if system != nil && system.Id == "NeoGeo" && strings.EqualFold(filepath.Ext(path), ".zip") {
		if title := m.NeoGeoTitle(path); title != "" {
			return title
		}
	}
	return filepath.Base(path)
}
