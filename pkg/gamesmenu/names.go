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
	"bufio"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"strings"
)

// illegalNameChars are removed from names.txt replacements before they are
// used as folder names, exactly as the Python script did.
const illegalNameChars = `<>:"/\|?*`

type nameEntry struct {
	key         string
	replacement string
}

// Names is MiSTer's names.txt display-name mapping with the Python script's
// lookup rules: the key before the first colon is compared case-insensitively
// after trimming, the first matching line wins, and the replacement has "/"
// turned into " & " and other filename-illegal characters turned into spaces.
type Names struct {
	entries []nameEntry
}

// LoadNames reads names.txt from path. A missing file yields an empty mapping
// that returns every name unchanged.
func LoadNames(path string) (*Names, error) {
	// #nosec G304 -- path is the configured names.txt location.
	file, err := os.Open(path)
	if errors.Is(err, fs.ErrNotExist) {
		return &Names{}, nil
	}
	if err != nil {
		return nil, fmt.Errorf("open names file: %w", err)
	}
	defer func() { _ = file.Close() }()

	names := &Names{}
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		key, replacement, ok := strings.Cut(scanner.Text(), ":")
		if !ok {
			continue
		}
		names.entries = append(names.entries, nameEntry{
			key:         strings.ToLower(strings.TrimSpace(key)),
			replacement: sanitizeName(strings.TrimSpace(replacement)),
		})
	}
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("read names file: %w", err)
	}
	return names, nil
}

func sanitizeName(replacement string) string {
	replacement = strings.ReplaceAll(replacement, "/", " & ")
	return strings.Map(func(r rune) rune {
		if strings.ContainsRune(illegalNameChars, r) {
			return ' '
		}
		return r
	}, replacement)
}

// Replace returns the names.txt display name for name, or name itself.
func (n *Names) Replace(name string) string {
	if n == nil {
		return name
	}
	lower := strings.ToLower(name)
	for _, entry := range n.entries {
		if entry.key == lower {
			return entry.replacement
		}
	}
	return name
}

// FolderName is the stock-menu folder name for name: an underscore so the
// MiSTer menu descends into it, then the display name.
func (n *Names) FolderName(name string) string {
	return "_" + n.Replace(name)
}

// MenuPath joins the menu folder name of every component with slashes.
func (n *Names) MenuPath(components ...string) string {
	folders := make([]string, 0, len(components))
	for _, component := range components {
		folders = append(folders, n.FolderName(component))
	}
	return strings.Join(folders, "/")
}
