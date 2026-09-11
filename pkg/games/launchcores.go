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
	"bytes"
	"encoding/xml"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"github.com/wizzomafizzo/mrext/pkg/config"
)

// LaunchCore is one way to run a system's games: a core file the launcher
// can see, or an MGL launcher that names a core and, usually, a setname so
// the core keeps its own config folder. RetroAchievements builds are
// installed the second way, as _RA_Cores/NES.mgl pointing at
// _RA_Cores/Cores/NES with the setname RA_NES.
type LaunchCore struct {
	Name           string // what to call it: the setname when there is one, else the core's short name
	Launch         string // value a client passes back to launch with this core
	RBF            string // core in the short form an MGL takes
	SetName        string // setname the launcher sets, empty for a plain core file
	Path           string // file this entry came from, .rbf or .mgl
	Filename       string
	SetNameSameDir bool // same_dir attribute of that setname
}

// IsLauncher reports whether the entry came from an MGL launcher rather than
// a core file.
func (c *LaunchCore) IsLauncher() bool {
	return strings.EqualFold(filepath.Ext(c.Path), ".mgl")
}

// ErrUnknownRBF is returned by FindLaunchCore when nothing the launcher can
// see matches.
var ErrUnknownRBF = errors.New("unknown rbf")

func rbfLaunchCore(info RBFInfo) LaunchCore {
	return LaunchCore{
		Name:     info.ShortName,
		Launch:   info.MGLName,
		RBF:      info.MGLName,
		Path:     info.Path,
		Filename: info.Filename,
	}
}

// mglLauncher is the part of an MGL file that describes a core launch.
type mglLauncher struct {
	XMLName xml.Name      `xml:"mistergamedescription"`
	RBF     string        `xml:"rbf"`
	SetName mglSetName    `xml:"setname"`
	Files   []mglFileMark `xml:"file"`
}

type mglSetName struct {
	Value   string `xml:",chardata"`
	SameDir string `xml:"same_dir,attr"`
}

// mglFileMark only records that a <file> element is present.
type mglFileMark struct{}

// parseMGLLauncher reads an MGL that launches a core on its own. An MGL
// with a <file> element launches a game, so it is not a core launcher.
func parseMGLLauncher(path string, data []byte) (LaunchCore, bool) {
	var mgl mglLauncher
	decoder := xml.NewDecoder(bytes.NewReader(data))
	decoder.Strict = false
	if err := decoder.Decode(&mgl); err != nil {
		return LaunchCore{}, false
	}

	rbf := strings.TrimSpace(mgl.RBF)
	if rbf == "" || len(mgl.Files) > 0 {
		return LaunchCore{}, false
	}

	setName := strings.TrimSpace(mgl.SetName.Value)
	sameDir := strings.TrimSpace(mgl.SetName.SameDir)
	core := LaunchCore{
		Name:           setName,
		Launch:         path,
		RBF:            rbf,
		SetName:        setName,
		SetNameSameDir: sameDir == "1" || strings.EqualFold(sameDir, "true"),
		Path:           path,
		Filename:       filepath.Base(path),
	}
	if strings.HasPrefix(path, config.SdFolder+"/") {
		core.Launch = strings.TrimPrefix(path, config.SdFolder+"/")
	}
	if core.Name == "" {
		core.Name = strings.TrimSuffix(core.Filename, filepath.Ext(core.Filename))
	}
	return core, true
}

// shallowScanMGL finds the core launchers in the top 2 menu levels of the
// SD card, the same levels shallowScanRBF covers.
func shallowScanMGL() ([]LaunchCore, error) {
	paths, err := shallowScanPaths(".mgl")
	if err != nil {
		return nil, err
	}

	results := make([]LaunchCore, 0, len(paths))
	for _, path := range paths {
		// #nosec G304 -- path comes from the menu scan, not from a request.
		data, err := os.ReadFile(path)
		if err != nil {
			continue
		}
		if core, ok := parseMGLLauncher(path, data); ok {
			results = append(results, core)
		}
	}
	return results, nil
}

func matchMGLLauncher(launchers []LaunchCore, name string) (LaunchCore, bool) {
	name = strings.TrimSpace(name)
	if name == "" {
		return LaunchCore{}, false
	}

	for _, core := range launchers {
		if strings.EqualFold(core.Launch, name) || core.Path == name {
			return core, true
		}
	}
	return LaunchCore{}, false
}

// FindLaunchCore looks up a core to launch a game with, by the value a
// client got from SystemLaunchCores: the short name of a core file
// (_Console/NES_Alt), the SD-relative path of an MGL launcher
// (_RA_Cores/NES.mgl), or the absolute path of either. Only files the
// launcher can see match.
func FindLaunchCore(name string) (LaunchCore, error) {
	name = strings.TrimSpace(name)
	if strings.EqualFold(filepath.Ext(name), ".mgl") {
		launchers, err := shallowScanMGL()
		if err != nil {
			return LaunchCore{}, fmt.Errorf("scan mgl files: %w", err)
		}
		if core, ok := matchMGLLauncher(launchers, name); ok {
			return core, nil
		}
		return LaunchCore{}, fmt.Errorf("%w: %s", ErrUnknownRBF, name)
	}

	rbfFiles, err := shallowScanRBF()
	if err != nil {
		return LaunchCore{}, fmt.Errorf("scan rbf files: %w", err)
	}
	if info, ok := matchRBF(rbfFiles, name); ok {
		return rbfLaunchCore(info), nil
	}
	return LaunchCore{}, fmt.Errorf("%w: %s", ErrUnknownRBF, name)
}

func matchRBF(rbfFiles []RBFInfo, name string) (RBFInfo, bool) {
	name = strings.TrimSpace(name)
	if name == "" {
		return RBFInfo{}, false
	}

	for _, info := range rbfFiles {
		if strings.EqualFold(info.MGLName, name) || info.Path == name {
			return info, true
		}
	}
	return RBFInfo{}, false
}

// SystemLaunchCores returns the cores a game of a system can be launched
// with: the core file the catalog names and any variant that adds a suffix
// to that name, such as NES_Alt beside NES, followed by every MGL launcher
// whose core is one of those. Core files that share a launch name, like two
// dated builds, are reported once, as the newest file.
func SystemLaunchCores(system *System) ([]LaunchCore, error) {
	rbfFiles, err := shallowScanRBF()
	if err != nil {
		return nil, fmt.Errorf("scan rbf files: %w", err)
	}
	launchers, err := shallowScanMGL()
	if err != nil {
		return nil, fmt.Errorf("scan mgl files: %w", err)
	}
	return matchSystemLaunchCores(rbfFiles, launchers, system), nil
}

var rbfDateSuffix = regexp.MustCompile(`_\d{8}$`)

// isCoreVariant reports whether a core's short name is the system's core or
// a suffixed build of it.
func isCoreVariant(shortName, base string) bool {
	short := strings.ToLower(shortName)
	return short == base || strings.HasPrefix(short, base+"_")
}

func matchSystemLaunchCores(rbfFiles []RBFInfo, launchers []LaunchCore, system *System) []LaunchCore {
	base := strings.ToLower(filepath.Base(system.Rbf))
	if base == "" || base == "." {
		return nil
	}

	byName := make(map[string]RBFInfo)
	for _, info := range rbfFiles {
		if !isCoreVariant(info.ShortName, base) {
			continue
		}

		key := strings.ToLower(info.MGLName)
		if prev, ok := byName[key]; ok && prev.Filename >= info.Filename {
			continue
		}
		byName[key] = info
	}

	results := make([]LaunchCore, 0, len(byName))
	for _, info := range byName {
		results = append(results, rbfLaunchCore(info))
	}
	sort.Slice(results, func(i, j int) bool {
		return strings.ToLower(results[i].Launch) < strings.ToLower(results[j].Launch)
	})

	fromLaunchers := make([]LaunchCore, 0)
	for _, core := range launchers {
		shortName := rbfDateSuffix.ReplaceAllString(filepath.Base(core.RBF), "")
		if isCoreVariant(shortName, base) {
			fromLaunchers = append(fromLaunchers, core)
		}
	}
	sort.Slice(fromLaunchers, func(i, j int) bool {
		return strings.ToLower(fromLaunchers[i].Launch) < strings.ToLower(fromLaunchers[j].Launch)
	})

	return append(results, fromLaunchers...)
}
