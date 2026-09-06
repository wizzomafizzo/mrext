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
	"os"

	"github.com/wizzomafizzo/mrext/pkg/config"
)

// Manager discovers games folders and maintains the _Games shortcut tree.
type Manager struct {
	cfg   *config.UserConfig
	names *Names
	paths Paths
}

// NewManager targets the real MiSTer filesystem.
func NewManager(cfg *config.UserConfig) (*Manager, error) {
	return NewManagerWithPaths(cfg, DefaultPaths())
}

// NewManagerWithPaths targets a relocated root and loads its names.txt once.
func NewManagerWithPaths(cfg *config.UserConfig, paths Paths) (*Manager, error) {
	names, err := LoadNames(paths.NamesFile)
	if err != nil {
		return nil, err
	}
	return &Manager{cfg: cfg, names: names, paths: paths}, nil
}

func (m *Manager) Paths() Paths {
	return m.paths
}

func (m *Manager) Names() *Names {
	return m.names
}

// MenuExists reports whether the _Games folder is present.
func (m *Manager) MenuExists() bool {
	info, err := os.Stat(m.paths.MenuFolder)
	return err == nil && info.IsDir()
}
