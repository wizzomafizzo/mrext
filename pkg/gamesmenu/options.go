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
	"path/filepath"
	"regexp"
	"strings"

	"github.com/wizzomafizzo/mrext/pkg/config"
)

// reservedMenuFolders are stock MiSTer folders that must never be used as the
// output folder, because Generate removes deselected subfolders.
var reservedMenuFolders = map[string]bool{
	"_alternative": true,
	"_arcade":      true,
	"_computer":    true,
	"_console":     true,
	"_jotego":      true,
	"_llapi":       true,
	"_other":       true,
	"_unofficial":  true,
	"_unstable":    true,
	"_utility":     true,
}

// MenuFolderNameFor returns the configured output folder name, or _Games.
func MenuFolderNameFor(cfg *config.UserConfig) string {
	if cfg == nil {
		return MenuFolderName
	}
	if name := strings.TrimSpace(cfg.GamesMenu.MenuFolder); name != "" {
		return name
	}
	return MenuFolderName
}

// ValidateMenuFolder checks a menu_folder value. It must be a single folder
// name at the SD root, start with "_" so the MiSTer menu lists it, and not be
// a stock core folder or another app's "_@" folder.
func ValidateMenuFolder(name string) error {
	if len(name) < 2 || !strings.HasPrefix(name, "_") {
		return fmt.Errorf("gamesmenu.menu_folder must start with _ and have a name: %q", name)
	}
	if strings.ContainsAny(name, `/\`+"\x00") || filepath.Base(name) != name || !filepath.IsLocal(name) {
		return fmt.Errorf("gamesmenu.menu_folder must be a single folder name: %q", name)
	}
	if strings.HasPrefix(name, "_@") || reservedMenuFolders[strings.ToLower(name)] {
		return fmt.Errorf("gamesmenu.menu_folder cannot be a MiSTer or app folder: %q", name)
	}
	return nil
}

// checkCustomMenuFolder refuses a custom output folder that holds cores, so a
// mistyped name cannot make Generate or Remove delete a core folder.
func checkCustomMenuFolder(menuFolder string) error {
	if filepath.Base(menuFolder) == MenuFolderName {
		return nil
	}
	children, err := os.ReadDir(menuFolder)
	if errors.Is(err, fs.ErrNotExist) {
		return nil
	}
	if err != nil {
		return fmt.Errorf("read Games menu folder: %w", err)
	}
	for _, child := range children {
		ext := strings.ToLower(filepath.Ext(child.Name()))
		if !child.IsDir() && (ext == ".rbf" || ext == ".mra") {
			return fmt.Errorf("refusing to use %s as the Games menu: it contains cores", filepath.Base(menuFolder))
		}
	}
	return nil
}

// excludeRules match file names against exclude patterns. Only * and ? are
// wildcards, so brackets and parentheses in ROM names match literally.
type excludeRules []*regexp.Regexp

func compileExclude(patterns []string) excludeRules {
	var rules excludeRules
	for _, pattern := range patterns {
		pattern = strings.TrimSpace(pattern)
		if pattern == "" {
			continue
		}
		var expr strings.Builder
		expr.WriteString("(?is)^")
		for _, r := range pattern {
			switch r {
			case '*':
				expr.WriteString(".*")
			case '?':
				expr.WriteString(".")
			default:
				expr.WriteString(regexp.QuoteMeta(string(r)))
			}
		}
		expr.WriteString("$")
		rules = append(rules, regexp.MustCompile(expr.String()))
	}
	return rules
}

func (r excludeRules) match(name string) bool {
	for _, rule := range r {
		if rule.MatchString(name) {
			return true
		}
	}
	return false
}
