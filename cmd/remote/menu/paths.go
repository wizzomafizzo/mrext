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

package menu

import (
	"errors"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/wizzomafizzo/mrext/pkg/config"
)

// errOutsideMenuRoot is returned for any request path that does not resolve
// inside the SD root. Handlers turn it into a 400.
var errOutsideMenuRoot = errors.New("path is outside the menu root")

// errProtectedPath is returned for paths MiSTer needs to boot. The menu API
// manages shortcuts; it has no reason to touch these, and a mistake here
// leaves a device that will not start.
var errProtectedPath = errors.New("path is protected")

// protectedNames are SD-root entries the menu API must never create over,
// rename or delete, matched against the first path component so naming a
// directory once covers everything inside it.
var protectedNames = []string{
	"menu.rbf",
	"u-boot.img",
	"u-boot.txt",
	"linux",
	"config",
}

// protectedPrefixes cover the dated and suffixed variants MiSTer ships, so
// MiSTer, MiSTer.ini and MiSTer_20240101.rbf are all matched. The delete
// handler already guarded this prefix; the others are new.
var protectedPrefixes = []string{
	"MiSTer",
}

// relativeMenuPath normalises a client-supplied path to one relative to the SD
// root. Clients send either an absolute path under the root or one already
// relative to it, and both have always been accepted.
//
// The empty path is kept empty rather than becoming ".", because the listing
// response uses it verbatim for parent and next links.
func relativeMenuPath(path string) string {
	path = removeRoot.ReplaceAllLiteralString(path, "")
	if path == "" {
		return ""
	}
	return filepath.Clean(path)
}

// resolveMenuPath turns a client-supplied path into an absolute one under the
// SD root, or reports why it cannot.
//
// The containment check is lexical and runs on the cleaned path, so "..", an
// absolute path elsewhere, and a path that only looks rooted are all rejected
// before any filesystem call. Cleaning before joining is what previously let
// leading ".." survive: filepath.Join cleans again, so a relative path with
// leading ".." escaped the root it was joined to.
func resolveMenuPath(path string) (string, error) {
	relative := relativeMenuPath(path)
	if filepath.IsAbs(relative) {
		return "", fmt.Errorf("%w: %s", errOutsideMenuRoot, path)
	}

	resolved := filepath.Join(config.SdFolder, relative)
	inside, err := filepath.Rel(config.SdFolder, resolved)
	if err != nil {
		return "", fmt.Errorf("resolve menu path: %w", err)
	}
	if inside == ".." || strings.HasPrefix(inside, ".."+string(filepath.Separator)) {
		return "", fmt.Errorf("%w: %s", errOutsideMenuRoot, path)
	}

	if inside != "." && isProtectedPath(inside) {
		return "", fmt.Errorf("%w: %s", errProtectedPath, path)
	}

	return resolved, nil
}

// isProtectedPath reports whether a root-relative path names a protected entry
// or anything inside one.
func isProtectedPath(relative string) bool {
	first := relative
	if separator := strings.IndexRune(relative, filepath.Separator); separator >= 0 {
		first = relative[:separator]
	}
	for _, name := range protectedNames {
		if strings.EqualFold(first, name) {
			return true
		}
	}
	for _, prefix := range protectedPrefixes {
		if len(first) >= len(prefix) && strings.EqualFold(first[:len(prefix)], prefix) {
			return true
		}
	}
	return false
}
