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
	"bytes"
	"os"
	"path/filepath"
	"strings"

	"github.com/wizzomafizzo/mrext/pkg/config"
	"github.com/wizzomafizzo/mrext/pkg/games"
)

func (m *Manager) NeoGeoTitle(path string) string {
	if !strings.EqualFold(filepath.Ext(path), ".zip") {
		return ""
	}
	base := m.neoGeoBase(path)
	if base == "" {
		return ""
	}
	folder := filepath.Dir(path)
	for {
		xmlPath := filepath.Join(folder, config.NeoGeoRomsetsFile)
		if info, err := os.Stat(xmlPath); err == nil && !info.IsDir() {
			names := m.loadNeoGeoNames(xmlPath)
			return names[strings.ToLower(strings.TrimSuffix(filepath.Base(path), filepath.Ext(path)))]
		}
		if samePath(folder, base) {
			return ""
		}
		parent := filepath.Dir(folder)
		if parent == folder {
			return ""
		}
		folder = parent
	}
}

func (m *Manager) neoGeoBase(path string) string {
	resolvedPath := resolveExistingPath(path)
	neoGeo, err := games.GetSystem("NeoGeo")
	if err != nil {
		return ""
	}
	best := ""
	folderNames := append(append([]string(nil), neoGeo.Folder...), "NeoGeo", "NEOGEO")
	for _, root := range games.GetGamesFolders(m.cfg) {
		for _, folder := range folderNames {
			candidate := filepath.Join(root, folder)
			if matched, matchErr := games.FindFile(candidate); matchErr == nil {
				candidate = matched
			}
			resolvedCandidate := resolveExistingPath(candidate)
			relative, relErr := filepath.Rel(resolvedCandidate, resolvedPath)
			if relErr != nil || relative == ".." || strings.HasPrefix(relative, ".."+string(filepath.Separator)) {
				continue
			}
			if len(resolvedCandidate) > len(best) {
				best = resolvedCandidate
			}
		}
	}
	return best
}

func resolveExistingPath(path string) string {
	resolved, err := filepath.EvalSymlinks(path)
	if err == nil {
		return resolved
	}
	return filepath.Clean(path)
}

func samePath(first, second string) bool {
	return filepath.Clean(resolveExistingPath(first)) == filepath.Clean(resolveExistingPath(second))
}

func (m *Manager) loadNeoGeoNames(path string) map[string]string {
	path = resolveExistingPath(path)
	if names, found := m.neoGeoNamesCache[path]; found {
		return names
	}
	names := make(map[string]string)
	defer func() { m.neoGeoNamesCache[path] = names }()
	// #nosec G304 -- path is a discovered romsets.xml file under a selected NeoGeo folder.
	data, err := os.ReadFile(path)
	if err != nil {
		return names
	}
	names, _ = games.ReadNeoGeoNames(bytes.NewReader(data))
	return names
}
