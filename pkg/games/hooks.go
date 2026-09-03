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

package games

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/wizzomafizzo/mrext/pkg/config"
	"github.com/wizzomafizzo/mrext/pkg/utils"
)

func copySetNameBIOS(cfg *config.UserConfig, origSystem, newSystem *System, name string) error {
	var biosPath string

	folders := GetActiveSystemPaths(cfg, []System{*origSystem})
	for i := range folders {
		checkPath := filepath.Join(folders[i].Path, name)
		if _, err := os.Stat(checkPath); err == nil {
			biosPath = checkPath
			break
		}
	}

	if biosPath == "" || newSystem.SetName == "" {
		return nil
	}

	newFolder, err := filepath.Abs(filepath.Join(filepath.Dir(biosPath), "..", newSystem.SetName))
	if err != nil {
		return fmt.Errorf("resolve BIOS destination: %w", err)
	}

	if _, err := os.Stat(filepath.Join(newFolder, name)); err == nil {
		return nil
	}

	if err := os.MkdirAll(newFolder, 0o750); err != nil {
		return fmt.Errorf("create BIOS destination: %w", err)
	}

	if err := utils.CopyFile(biosPath, filepath.Join(newFolder, name)); err != nil {
		return fmt.Errorf("copy BIOS: %w", err)
	}
	return nil
}

func hookFDS(cfg *config.UserConfig, system *System, _ string) (string, error) {
	nesSystem, err := GetSystem("NES")
	if err != nil {
		return "", err
	}

	return "", copySetNameBIOS(cfg, nesSystem, system, "boot0.rom")
}

func hookWSC(cfg *config.UserConfig, system *System, _ string) (string, error) {
	wsSystem, err := GetSystem("WonderSwan")
	if err != nil {
		return "", err
	}

	err = copySetNameBIOS(cfg, wsSystem, system, "boot.rom")
	if err != nil {
		return "", err
	}

	return "", copySetNameBIOS(cfg, wsSystem, system, "boot1.rom")
}

func hookAo486(_ *config.UserConfig, system *System, path string) (string, error) {
	mglDef, err := PathToMglDef(system, path)
	if err != nil {
		return "", err
	}

	var mgl string
	if !strings.HasSuffix(strings.ToLower(path), ".vhd") {
		return "", nil
	}

	dir := filepath.Dir(path)
	filename := filepath.Base(path)

	// exception for Top 300 pack which uses 2 disks
	if strings.HasSuffix(path, "IDE 0-1 Top 300 DOS Games.vhd") {
		mgl += fmt.Sprintf(
			"\t<file delay=\"%d\" type=%q index=\"%d\" path=%q/>\n",
			mglDef.Delay,
			mglDef.Method,
			mglDef.Index,
			"../../../../.."+filepath.Join(dir, "IDE 0-0 BOOT-DOS98.vhd"),
		)

		mgl += fmt.Sprintf(
			"\t<file delay=\"%d\" type=%q index=\"%d\" path=%q/>\n",
			mglDef.Delay,
			mglDef.Method,
			mglDef.Index+1,
			"../../../../.."+path,
		)
		mgl += "\t<reset delay=\"1\"/>\n"
		return mgl, nil
	}

	files, err := os.ReadDir(dir)
	if err != nil {
		return "", fmt.Errorf("read ao486 game directory: %w", err)
	}

	// if there's an iso in the same folder, mount it too
	var extraDisk strings.Builder
	for _, file := range files {
		lowerName := strings.ToLower(file.Name())
		isDisc := strings.HasSuffix(lowerName, ".iso") || strings.HasSuffix(lowerName, ".chd")
		if isDisc && file.Name() != filename {
			_, _ = fmt.Fprintf(
				&extraDisk,
				"\t<file delay=\"%d\" type=%q index=\"%d\" path=%q/>\n",
				mglDef.Delay,
				mglDef.Method,
				4,
				"../../../../.."+filepath.Join(dir, file.Name()),
			)
			break
		}
	}
	mgl += extraDisk.String()

	mgl += fmt.Sprintf(
		"\t<file delay=\"%d\" type=%q index=\"%d\" path=%q/>\n",
		mglDef.Delay,
		mglDef.Method,
		mglDef.Index,
		"../../../../.."+path,
	)
	mgl += "\t<reset delay=\"1\"/>\n"
	return mgl, nil
}

func hookAmiga(_ *config.UserConfig, _ *System, path string) (string, error) {
	lowerDir := strings.ToLower(filepath.Dir(path))
	if !strings.HasSuffix(lowerDir, "listings/games.txt") &&
		!strings.HasSuffix(lowerDir, "listings/demos.txt") {
		return "", nil
	}

	gameName := filepath.Base(path)
	sharedPath, err := filepath.Abs(filepath.Join(filepath.Dir(path), "..", "..", "shared"))
	if err != nil {
		return "", fmt.Errorf("resolve Amiga shared path: %w", err)
	}

	bootFile := filepath.Join(sharedPath, "ags_boot")
	if err := os.WriteFile(bootFile, []byte(gameName+"\n"), 0o600); err != nil {
		return "", fmt.Errorf("write Amiga boot selection: %w", err)
	}

	return "\t<setname>Amiga</setname>\n", nil
}

func hookNeoGeo(_ *config.UserConfig, _ *System, path string) (string, error) {
	// neogeo core allows launching zips and folders
	if strings.HasSuffix(strings.ToLower(path), ".zip") || filepath.Ext(path) == "" {
		return fmt.Sprintf(
			"\t<file delay=\"%d\" type=%q index=\"%d\" path=%q/>\n",
			1,
			"f",
			1,
			"../../../../.."+path,
		), nil
	}

	return "", nil
}

var systemHooks = map[string]func(*config.UserConfig, *System, string) (string, error){
	"FDS":             hookFDS,
	"WonderSwanColor": hookWSC,
	"ao486":           hookAo486,
	"Amiga":           hookAmiga,
	"NeoGeo":          hookNeoGeo,
}

func RunSystemHook(cfg *config.UserConfig, system *System, path string) (string, error) {
	if hook, ok := systemHooks[system.Id]; ok {
		return hook(cfg, system, path)
	}

	return "", nil
}
