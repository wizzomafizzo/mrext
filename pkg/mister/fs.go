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

package mister

import (
	"bytes"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"syscall"

	"github.com/wizzomafizzo/mrext/pkg/config"
	"github.com/wizzomafizzo/mrext/pkg/games"
	"github.com/wizzomafizzo/mrext/pkg/utils"
)

func GetActiveCoreName() (string, error) {
	data, err := os.ReadFile(config.CoreNameFile)
	if err != nil {
		return "", fmt.Errorf("read active core name: %w", err)
	}

	return string(data), nil
}

func ActiveGameEnabled() bool {
	_, err := os.Stat(config.ActiveGameFile)
	return err == nil
}

func SetActiveGame(path string) error {
	// #nosec G306 -- ACTIVEGAME is a legacy cross-process compatibility file.
	if err := os.WriteFile(config.ActiveGameFile, []byte(path), 0o644); err != nil {
		return fmt.Errorf("write active game: %w", err)
	}
	return nil
}

func GetActiveGame() (string, error) {
	data, err := os.ReadFile(config.ActiveGameFile)
	if err != nil {
		return "", fmt.Errorf("read active game: %w", err)
	}

	return string(data), nil
}

// Convert a launchable path to an absolute path.
func ResolvePath(path string) string {
	if path == "" {
		return path
	}
	if filepath.IsAbs(path) {
		return filepath.Clean(path)
	}

	abs, err := filepath.Abs(filepath.Join(config.SdFolder, path))
	if err != nil {
		return path
	}

	return abs
}

// Search for directories in root that start with "_".
func GetMenuFolders(root string) []string {
	var folders []string

	// TODO: confirm menu can't traverse symlinks
	var scan func(path string)
	scan = func(folder string) {
		files, err := os.ReadDir(folder)
		if err != nil {
			return
		}
		for _, file := range files {
			if file.IsDir() && file.Name()[0] == '_' {
				path := filepath.Join(folder, file.Name())
				folders = append(folders, path)
				scan(path)
			}
		}
	}

	scan(root)
	return folders
}

type RecentEntry struct {
	Directory string
	Name      string
	Label     string
}

func ReadRecent(path string) ([]RecentEntry, error) {
	var recents []RecentEntry

	if _, err := os.Stat(path); err != nil {
		return nil, fmt.Errorf("stat recent file: %w", err)
	}

	// #nosec G304 -- path is an explicit MiSTer recent-file input.
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("open recent file: %w", err)
	}
	defer func() { _ = file.Close() }()

	for {
		entry := make([]byte, 1024+256+256)
		n, err := io.ReadFull(file, entry)
		if errors.Is(err, io.EOF) && n == 0 {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("read recent entry: %w", err)
		}

		empty := true
		for _, b := range entry {
			if b != 0 {
				empty = false
			}
		}
		if empty {
			break
		}

		recents = append(recents, RecentEntry{
			Directory: strings.Trim(string(entry[:1024]), "\x00"),
			Name:      strings.Trim(string(entry[1024:1280]), "\x00"),
			Label:     strings.Trim(string(entry[1280:1536]), "\x00"),
		})
	}

	return recents, nil
}

type MGLFile struct {
	XMLName xml.Name `xml:"file"`
	Type    string   `xml:"type,attr"`
	Path    string   `xml:"path,attr"`
	Delay   int      `xml:"delay,attr"`
	Index   int      `xml:"index,attr"`
}

type MGL struct {
	XMLName xml.Name `xml:"mistergamedescription"`
	Rbf     string   `xml:"rbf"` //nolint:revive // Legacy field name is part of Remote JSON.
	SetName string   `xml:"setname"`
	File    MGLFile  `xml:"file"`
}

func ReadMGL(path string) (MGL, error) {
	var mgl MGL

	if _, err := os.Stat(path); err != nil {
		return mgl, fmt.Errorf("stat MGL file: %w", err)
	}

	// #nosec G304 -- path is an explicit MGL file input.
	file, err := os.ReadFile(path)
	if err != nil {
		return mgl, fmt.Errorf("read MGL file: %w", err)
	}

	decoder := xml.NewDecoder(bytes.NewReader(file))
	decoder.Strict = false

	err = decoder.Decode(&mgl)
	if err != nil {
		return mgl, fmt.Errorf("decode MGL file: %w", err)
	}

	return mgl, nil
}

type MenuConfig struct {
	BackgroundMode int
}

const (
	BackgroundModeNone      = 0
	BackgroundModeWallpaper = 2
	BackgroundModeHBars1    = 4
	BackgroundModeHBars2    = 6
	BackgroundModeVBars1    = 8
	BackgroundModeVBars2    = 10
	BackgroundModeSpectrum  = 12
	BackgroundModeBlack     = 14
)

func ReadMenuConfig() (MenuConfig, error) {
	var cfg MenuConfig

	if _, err := os.Stat(config.MenuConfigFile); err != nil {
		return cfg, fmt.Errorf("stat menu configuration: %w", err)
	}

	file, err := os.ReadFile(config.MenuConfigFile)
	if err != nil {
		return cfg, fmt.Errorf("read menu configuration: %w", err)
	}
	if len(file) == 0 {
		return cfg, errors.New("menu configuration is empty")
	}

	cfg.BackgroundMode = int(file[0])

	return cfg, nil
}

func SetMenuBackgroundMode(mode int) error {
	if !utils.Contains([]int{
		BackgroundModeNone,
		BackgroundModeWallpaper,
		BackgroundModeHBars1,
		BackgroundModeHBars2,
		BackgroundModeVBars1,
		BackgroundModeVBars2,
		BackgroundModeSpectrum,
		BackgroundModeBlack,
	}, mode) {
		return errors.New("invalid background mode")
	}

	cfg, err := ReadMenuConfig()
	if err != nil {
		return err
	}

	if cfg.BackgroundMode == mode {
		return nil
	}

	file, err := os.ReadFile(config.MenuConfigFile)
	if err != nil {
		return fmt.Errorf("read menu configuration: %w", err)
	}
	if len(file) == 0 {
		return errors.New("menu configuration is empty")
	}

	// #nosec G115 -- mode is restricted to known byte-sized constants above.
	file[0] = byte(mode)

	// #nosec G306,G703 -- fixed MiSTer menu configuration must remain world-readable.
	if err := os.WriteFile(config.MenuConfigFile, file, 0o644); err != nil {
		return fmt.Errorf("write menu configuration: %w", err)
	}
	return nil
}

func GetMounts(cfg *config.UserConfig) ([]string, error) {
	file, err := os.ReadFile("/proc/mounts")
	if err != nil {
		return nil, fmt.Errorf("read mounts: %w", err)
	}

	var mounts []string
	gamesFolders := games.GetGamesFolders(cfg)

	for _, line := range strings.Split(string(file), "\n") {
		if line == "" {
			continue
		}

		parts := strings.Split(line, " ")

		if len(parts) < 2 {
			continue
		}

		if utils.Contains(gamesFolders, parts[1]) {
			mounts = append(mounts, parts[1])
		}
	}

	return mounts, nil
}

type DiskUsage struct {
	Total uint64
	Free  uint64
	Used  uint64
}

func GetDiskUsage(path string) (DiskUsage, error) {
	var usage DiskUsage

	stat := syscall.Statfs_t{}
	err := syscall.Statfs(path, &stat)
	if err != nil {
		return usage, fmt.Errorf("stat filesystem: %w", err)
	}
	if stat.Bsize < 0 {
		return usage, fmt.Errorf("invalid filesystem block size: %d", stat.Bsize)
	}

	blockSize := uint64(stat.Bsize)
	usage.Total = stat.Blocks * blockSize
	usage.Free = stat.Bfree * blockSize
	usage.Used = usage.Total - usage.Free

	return usage, nil
}
