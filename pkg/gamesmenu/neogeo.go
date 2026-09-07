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
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/wizzomafizzo/mrext/pkg/config"
	"github.com/wizzomafizzo/mrext/pkg/games"
)

// Definitions are shared across the entry's roots so an SD romsets.xml can
// describe external-drive sets. Earlier roots retain naming priority.
func (s *menuScanner) prepareNeoGeo() {
	s.neoGeo, s.neoGeoNames = nil, nil
	for index := range s.entry.Systems {
		if s.entry.Systems[index].Id == "NeoGeo" {
			s.neoGeo = &s.entry.Systems[index]
			break
		}
	}
	if s.neoGeo == nil {
		return
	}
	s.neoGeoNames = make(map[string]string)
	for _, folder := range s.entry.Paths {
		filename, err := games.FindFile(filepath.Join(folder, config.NeoGeoRomsetsFile))
		if err != nil {
			continue // A mapping is optional; ordinary .neo files still work.
		}
		names, err := readNeoGeoNames(filename)
		if err != nil {
			s.result.fail(fmt.Errorf("read %s: %w", filename, err))
			continue
		}
		for name, title := range names {
			if _, exists := s.neoGeoNames[name]; !exists {
				s.neoGeoNames[name] = title
			}
		}
	}
}

func readNeoGeoNames(filename string) (map[string]string, error) {
	info, err := os.Stat(filename)
	if err != nil {
		return nil, fmt.Errorf("stat ROM-set mapping: %w", err)
	}
	if !info.Mode().IsRegular() {
		return nil, fmt.Errorf("ROM-set mapping is not a regular file: %s", filename)
	}
	// #nosec G304 -- mapping is discovered under a selected games root.
	file, err := os.Open(filename)
	if err != nil {
		return nil, fmt.Errorf("open ROM-set mapping: %w", err)
	}
	defer func() { _ = file.Close() }()
	names, err := games.ReadNeoGeoNames(file)
	if err != nil {
		return nil, fmt.Errorf("read ROM-set mapping: %w", err)
	}
	return names, nil
}

// A mapped set is one launch target. Never descend into its component ROMs or
// expand its ZIP as an ordinary collection, even when shortcut creation fails.
func (s *menuScanner) writeNeoGeoSet(full, relative string, isDirectory bool) bool {
	if s.neoGeo == nil {
		return false
	}
	name := strings.ToLower(filepath.Base(full))
	if !isDirectory {
		if filepath.Ext(name) != ".zip" {
			return false
		}
		name = strings.TrimSuffix(name, ".zip")
	}
	title, ok := s.neoGeoNames[name]
	if !ok {
		return false
	}
	override, err := games.NeoGeoMGLOverride(s.neoGeo, "../../../../.."+filepath.ToSlash(full))
	if err != nil {
		s.result.fail(fmt.Errorf("generate NeoGeo set %s: %w", full, err))
		return true
	}
	s.writeShortcut(full, relativeFolders(relative), sanitizeName(title)+".mgl", s.neoGeo, override)
	return true
}
