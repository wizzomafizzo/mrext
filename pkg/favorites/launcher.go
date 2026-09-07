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
	"sort"
	"strings"

	"github.com/ZaparooProject/zaparoo-core/mister/catalog"
	mglgen "github.com/ZaparooProject/zaparoo-core/mister/mgl"
	"github.com/wizzomafizzo/mrext/pkg/games"
	"github.com/wizzomafizzo/mrext/pkg/mister"
)

const BadCharacters = `<>:"/\|?*`

var alternateCores = map[string]map[string]string{
	"llapi": {
		"Atari7800": "Atari7800_LLAPI", "Gameboy": "Gameboy_LLAPI", "GBA2P": "GBA2P_LLAPI",
		"GBA": "GBA_LLAPI", "Genesis": "Genesis_LLAPI", "MegaCD": "MegaCD_LLAPI",
		"NeoGeo": "NeoGeo_LLAPI", "NES": "NES_LLAPI", "Sega32X": "S32X_LLAPI",
		"SuperGameboy": "SGB_LLAPI", "MasterSystem": "SMS_LLAPI", "SNES": "SNES_LLAPI",
		"TurboGrafx16CD": "TurboGrafx16_LLAPI", "TurboGrafx16": "TurboGrafx16_LLAPI",
	},
	"yc": {
		"Atari2600": "Atari7800YC", "Atari7800": "Atari7800YC", "C64": "C64YC",
		"ColecoVision": "ColecoVisionYC", "Gameboy": "GameboyYC", "Genesis": "GenesisYC",
		"MegaCD": "MegaCDYC", "NeoGeo": "NeoGeoYC", "NES": "NESYC", "PSX": "PSXYC",
		"Sega32X": "S32XYC", "SuperGameboy": "SGBYC", "MasterSystem": "SMSYC", "SNES": "SNESYC",
		"TurboGrafx16CD": "TurboGrafx16YC", "TurboGrafx16": "TurboGrafx16YC",
	},
}

func HasBadChars(value string) bool {
	return strings.ContainsAny(value, BadCharacters)
}

func ValidateDisplayName(name string) error {
	switch {
	case strings.TrimSpace(name) == "":
		return errors.New("display name cannot be empty")
	case HasBadChars(name):
		return fmt.Errorf("display name cannot contain any of these characters: %s", BadCharacters)
	default:
		return nil
	}
}

func (m *Manager) DefaultName(system *games.System, path string) string {
	name := strings.TrimSuffix(filepath.Base(path), filepath.Ext(path))
	if system != nil && system.Id == "NeoGeo" && strings.EqualFold(filepath.Ext(path), ".zip") {
		if title := m.NeoGeoTitle(path); title != "" {
			cleaned := strings.Map(func(value rune) rune {
				if strings.ContainsRune(BadCharacters, value) {
					return ' '
				}
				return value
			}, title)
			if safeName := strings.Trim(strings.Join(strings.Fields(cleaned), " "), "."); safeName != "" {
				return safeName
			}
		}
	}
	return name
}

func (m *Manager) CreateCoreFavorite(source, destination, name string) (string, error) {
	if err := m.validateDestination(destination, true); err != nil {
		return "", err
	}
	if err := ValidateDisplayName(name); err != nil {
		return "", err
	}
	extension := strings.ToLower(filepath.Ext(source))
	if extension != ".rbf" && extension != ".mra" && extension != ".mgl" {
		return "", fmt.Errorf("unsupported favorite file: %s", source)
	}
	if strings.EqualFold(filepath.Ext(name), extension) {
		name = strings.TrimSuffix(name, filepath.Ext(name))
	}
	path := filepath.Join(destination, name+extension)
	if _, err := os.Lstat(path); err == nil {
		return "", fmt.Errorf("favorite already exists: %s", filepath.Base(path))
	} else if !os.IsNotExist(err) {
		return "", fmt.Errorf("inspect favorite destination: %w", err)
	}
	if err := os.Symlink(source, path); err != nil {
		return "", fmt.Errorf("create core favorite: %w", err)
	}
	return path, nil
}

func (m *Manager) CreateGameFavorite(
	system *games.System,
	mediaPath, destination, name string,
) (string, error) {
	if system == nil {
		return "", errors.New("no system selected")
	}
	if err := m.validateDestination(destination, true); err != nil {
		return "", err
	}
	if err := ValidateDisplayName(name); err != nil {
		return "", err
	}
	launcherPath := filepath.Join(destination, name+".mgl")
	if _, err := os.Lstat(launcherPath); err == nil {
		return "", fmt.Errorf("favorite already exists: %s", filepath.Base(launcherPath))
	} else if !os.IsNotExist(err) {
		return "", fmt.Errorf("inspect game favorite destination: %w", err)
	}

	resolvedSystem := *system
	resolvedSystem.Rbf = m.resolveCore(&resolvedSystem)
	if strings.EqualFold(strings.TrimSpace(m.cfg.FavoritesCores.All), "ra") {
		if err := m.configureRACore(&resolvedSystem); err != nil {
			return "", err
		}
	}
	launcher, err := m.generateMGL(&resolvedSystem, mediaPath)
	if err != nil {
		return "", err
	}
	// #nosec G306,G703 -- MiSTer menu launchers must remain world-readable.
	if err := os.WriteFile(launcherPath, []byte(launcher), 0o644); err != nil {
		return "", fmt.Errorf("write game favorite: %w", err)
	}
	return launcherPath, nil
}

func (m *Manager) generateMGL(system *games.System, mediaPath string) (string, error) {
	if system.Id != "NeoGeo" {
		launcher, err := mister.GenerateMgl(m.cfg, system, mediaPath, "")
		if err != nil {
			return "", fmt.Errorf("generate favorite MGL: %w", err)
		}
		return launcher, nil
	}

	base := m.neoGeoBase(mediaPath)
	if base == "" {
		launcher, err := mister.GenerateMgl(m.cfg, system, mediaPath, "")
		if err != nil {
			return "", fmt.Errorf("generate NeoGeo favorite MGL: %w", err)
		}
		return launcher, nil
	}
	relativePath, err := filepath.Rel(base, resolveExistingPath(mediaPath))
	if err != nil {
		return "", fmt.Errorf("resolve NeoGeo favorite path: %w", err)
	}
	core := games.CatalogCore(system)
	if _, slotErr := catalog.PathToMGLDef(&core, mediaPath); slotErr != nil &&
		!strings.EqualFold(filepath.Ext(mediaPath), ".zip") {
		return "", fmt.Errorf("resolve NeoGeo MGL parameters: %w", slotErr)
	}
	override, err := games.NeoGeoMGLOverride(system, filepath.ToSlash(relativePath))
	if err != nil {
		return "", fmt.Errorf("generate NeoGeo favorite mount: %w", err)
	}
	launcher, err := mglgen.Generate(&core, core.RBF, mediaPath, override)
	if err != nil {
		return "", fmt.Errorf("generate relative NeoGeo favorite MGL: %w", err)
	}
	return launcher, nil
}

func (m *Manager) resolveCore(system *games.System) string {
	defaultRBF := m.cfg.Favorites.CorePrefix + system.Rbf
	mode := strings.ToLower(strings.TrimSpace(m.cfg.FavoritesCores.All))
	if mode == "" {
		return defaultRBF
	}
	coreName := alternateCores[mode][system.Id]
	if coreName == "" {
		return defaultRBF
	}
	for _, core := range m.shallowRBFs() {
		if strings.HasPrefix(filepath.Base(core.Path), coreName) {
			return core.MGLName
		}
	}
	return defaultRBF
}

func (m *Manager) shallowRBFs() []games.RBFInfo {
	entries, err := os.ReadDir(m.paths.SDRoot)
	if err != nil {
		return nil
	}
	paths := make([]string, 0)
	for _, entry := range entries {
		path := filepath.Join(m.paths.SDRoot, entry.Name())
		if !entry.IsDir() && !isDirectorySymlink(path, entry) {
			if strings.EqualFold(filepath.Ext(entry.Name()), ".rbf") {
				paths = append(paths, path)
			}
			continue
		}
		if _, allowed := allowedRootEntries[strings.ToLower(entry.Name())]; !allowed {
			continue
		}
		subEntries, readErr := os.ReadDir(path)
		if readErr != nil {
			continue
		}
		for _, subEntry := range subEntries {
			if !subEntry.IsDir() && strings.EqualFold(filepath.Ext(subEntry.Name()), ".rbf") {
				paths = append(paths, filepath.Join(path, subEntry.Name()))
			}
		}
	}
	sort.Strings(paths)
	cores := make([]games.RBFInfo, 0, len(paths))
	for _, path := range paths {
		core := games.ParseRBF(path)
		if relativeDir, relativeErr := filepath.Rel(m.paths.SDRoot, filepath.Dir(path)); relativeErr == nil {
			core.MGLName = filepath.Join(relativeDir, core.ShortName)
		}
		cores = append(cores, core)
	}
	return cores
}
