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
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/wizzomafizzo/mrext/pkg/config"
	"github.com/wizzomafizzo/mrext/pkg/games"
)

type raCore struct {
	name    string
	setName string
	sameDir bool
}

// Adapted from Zaparoo Core pkg/platforms/mister/launchers.go (7cae7f1f),
// retroAchievementsSetName and RA launcher registrations. Keep these variant
// overrides aligned until Core exposes them through a shared library. System
// folders, extensions, and slot parameters still come from the shared catalog.
var raCores = map[string]raCore{
	"Atari2600":      {"Atari7800", "RA_Atari7800", true},
	"Atari7800":      {"Atari7800", "RA_Atari7800", true},
	"FDS":            {"NES", "RA_FDS", false},
	"Gameboy":        {"Gameboy", "RA_Gameboy", true},
	"GameboyColor":   {"Gameboy", "RA_GBC", false},
	"GameGear":       {"SMS", "RA_GameGear", false},
	"SuperGameboy":   {"Gameboy", "RA_SGB", false},
	"GBA":            {"GBA", "RA_GBA", true},
	"Genesis":        {"MegaDrive", "RA_MegaDrive", true},
	"MegaCD":         {"MegaCD", "RA_MegaCD", true},
	"MasterSystem":   {"SMS", "RA_SMS", true},
	"NeoGeo":         {"NeoGeo", "RA_NeoGeo", true},
	"NeoGeoCD":       {"NeoGeo", "RA_NeoGeoCD", false},
	"NES":            {"NES", "RA_NES", true},
	"Nintendo64":     {"N64", "RA_N64", true},
	"PSX":            {"PSX", "RA_PSX", true},
	"Sega32X":        {"S32X", "RA_S32X", true},
	"SNES":           {"SNES", "RA_SNES", true},
	"Saturn":         {"Saturn", "RA_Saturn", true},
	"TurboGrafx16":   {"TurboGrafx16", "RA_TurboGrafx16", true},
	"TurboGrafx16CD": {"TurboGrafx16", "RA_TurboGrafx16CD", false},
}

func (m *Manager) configureRACore(system *games.System) error {
	variant, ok := raCores[system.Id]
	if !ok {
		return nil
	}
	folder := filepath.Join(m.paths.SDRoot, config.RACoresFolder)
	entries, err := os.ReadDir(folder)
	if os.IsNotExist(err) {
		return nil
	}
	if err != nil {
		return fmt.Errorf("read RA cores: %w", err)
	}
	for _, entry := range entries {
		if !strings.EqualFold(filepath.Ext(entry.Name()), ".rbf") {
			continue
		}
		core := games.ParseRBF(filepath.Join(folder, entry.Name()))
		if !strings.EqualFold(core.ShortName, variant.name) {
			continue
		}
		info, statErr := os.Stat(core.Path)
		if statErr != nil {
			return fmt.Errorf("inspect RA core: %w", statErr)
		}
		if !info.Mode().IsRegular() {
			continue
		}
		// The RA Atari2600 launcher uses the Atari7800 core's loading slot,
		// retaining Atari2600's accepted media extensions and identity.
		if system.Id == "Atari2600" {
			atari7800, lookupErr := games.GetSystem("Atari7800")
			if lookupErr != nil {
				return fmt.Errorf("get RA Atari7800 slot: %w", lookupErr)
			}
			params, slotErr := games.PathToMglDef(atari7800, "game.a78")
			if slotErr != nil {
				return fmt.Errorf("get RA Atari7800 slot parameters: %w", slotErr)
			}
			if params == nil {
				return errors.New("RA Atari7800 catalog slot is missing")
			}
			slots := make([]games.Slot, len(system.Slots))
			for i, slot := range system.Slots {
				slots[i] = slot
				slots[i].Mgl = params
			}
			system.Slots = slots
		}
		system.Rbf = filepath.Join(config.RACoresFolder, core.ShortName)
		system.SetName = variant.setName
		system.SetNameSameDir = variant.sameDir
		return nil
	}
	return nil
}
