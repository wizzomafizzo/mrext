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
	"sort"
	"strings"

	"github.com/wizzomafizzo/mrext/pkg/games"
)

// folderSystems lists the catalog systems that can launch files found in one
// games folder. Several systems share folders (GAMEBOY, SMS, NES, TGFX16 and
// so on), so the systems are ordered for deterministic extension matching:
// systems that list the folder earlier in their own folder list come first,
// then by ID.
type folderSystems struct {
	Key     string
	Systems []games.System
}

type candidate struct {
	system      games.System
	folderIndex int
}

// menuSystems returns every catalog games folder that at least one system can
// generate MGL shortcuts for, sorted case-insensitively by folder name.
func menuSystems() []folderSystems {
	byKey := make(map[string]*folderSystems)
	candidates := make(map[string][]candidate)
	systems := games.AllSystems()
	for index := range systems {
		system := &systems[index]
		if !hasMglSlot(system) {
			continue
		}
		for index, folder := range system.Folder {
			lower := strings.ToLower(folder)
			if _, ok := byKey[lower]; !ok {
				byKey[lower] = &folderSystems{Key: folder}
			}
			candidates[lower] = append(candidates[lower], candidate{system: *system, folderIndex: index})
		}
	}

	result := make([]folderSystems, 0, len(byKey))
	for lower, entry := range byKey {
		matches := candidates[lower]
		sort.Slice(matches, func(i, j int) bool {
			if matches[i].folderIndex != matches[j].folderIndex {
				return matches[i].folderIndex < matches[j].folderIndex
			}
			return matches[i].system.Id < matches[j].system.Id
		})
		entry.Systems = make([]games.System, 0, len(matches))
		for index := range matches {
			entry.Systems = append(entry.Systems, matches[index].system)
		}
		result = append(result, *entry)
	}
	sort.Slice(result, func(i, j int) bool {
		left, right := strings.ToLower(result[i].Key), strings.ToLower(result[j].Key)
		if left != right {
			return left < right
		}
		return result[i].Key < result[j].Key
	})
	return result
}

func hasMglSlot(system *games.System) bool {
	for _, slot := range system.Slots {
		if slot.Mgl != nil && len(slot.Exts) > 0 {
			return true
		}
	}
	return false
}

// extensionRule maps one lower-cased file suffix to the system that launches
// it. Rules are ordered longest suffix first so ".p8.png" beats ".png".
type extensionRule struct {
	system *games.System
	suffix string
}

func extensionRules(candidates []games.System) []extensionRule {
	var rules []extensionRule
	for index := range candidates {
		system := &candidates[index]
		for _, slot := range system.Slots {
			if slot.Mgl == nil {
				continue
			}
			for _, ext := range slot.Exts {
				rules = append(rules, extensionRule{system: system, suffix: strings.ToLower(ext)})
			}
		}
	}
	sort.SliceStable(rules, func(i, j int) bool { return len(rules[i].suffix) > len(rules[j].suffix) })
	return rules
}

// matchSystem returns the system that launches name, or nil when no MGL slot
// accepts its extension. Matching is a case-insensitive suffix test, the same
// rule the catalog uses to pick an MGL slot.
func matchSystem(rules []extensionRule, name string) *games.System {
	lower := strings.ToLower(name)
	for _, rule := range rules {
		if strings.HasSuffix(lower, rule.suffix) {
			return rule.system
		}
	}
	return nil
}
