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
	"strings"
	"testing"
)

func TestMatchRBFByLaunchNameAndPath(t *testing.T) {
	t.Parallel()

	rbfFiles := []RBFInfo{
		ParseRBF("/media/fat/_Console/NES_20240310.rbf"),
		ParseRBF("/media/fat/_Console/NES_Alt_20240102.rbf"),
	}

	tests := []struct {
		name  string
		want  string
		found bool
	}{
		{"_Console/NES", "NES_20240310.rbf", true},
		{"_console/nes_alt", "NES_Alt_20240102.rbf", true},
		{"/media/fat/_Console/NES_Alt_20240102.rbf", "NES_Alt_20240102.rbf", true},
		{"NES", "", false},
		{"_Console/NES_20240310", "", false},
		{"../menu", "", false},
		{"", "", false},
	}

	for _, tt := range tests {
		info, ok := matchRBF(rbfFiles, tt.name)
		if ok != tt.found {
			t.Fatalf("%q: found %v, want %v", tt.name, ok, tt.found)
		}
		if info.Filename != tt.want {
			t.Fatalf("%q: got %q, want %q", tt.name, info.Filename, tt.want)
		}
	}
}

func TestMatchSystemLaunchCoresKeepsVariantsAndNewestRelease(t *testing.T) {
	t.Parallel()

	system, err := GetSystem("NES")
	if err != nil {
		t.Fatal(err)
	}

	rbfFiles := []RBFInfo{
		ParseRBF("/media/fat/_Console/NES_20230101.rbf"),
		ParseRBF("/media/fat/_Console/NES_20240310.rbf"),
		ParseRBF("/media/fat/_Console/NES_Alt_20240102.rbf"),
		ParseRBF("/media/fat/_Other/NES_Dev_20240201.rbf"),
		ParseRBF("/media/fat/_Console/SNES_20240310.rbf"),
		ParseRBF("/media/fat/_Console/NESMusic_20240310.rbf"),
		ParseRBF("/media/fat/_Console/Genesis_20240310.rbf"),
	}

	got := matchSystemLaunchCores(rbfFiles, nil, system)

	want := []string{
		"_Console/NES", "NES_20240310.rbf",
		"_Console/NES_Alt", "NES_Alt_20240102.rbf",
		"_Other/NES_Dev", "NES_Dev_20240201.rbf",
	}
	if len(got) != len(want)/2 {
		t.Fatalf("got %d results, want %d: %+v", len(got), len(want)/2, got)
	}
	for i, core := range got {
		if core.Launch != want[i*2] || core.Filename != want[i*2+1] {
			t.Fatalf("result %d: got %s %s, want %s %s", i, core.Launch, core.Filename, want[i*2], want[i*2+1])
		}
		if core.RBF != core.Launch || core.SetName != "" || core.IsLauncher() {
			t.Fatalf("result %d: a core file should launch as itself with no setname: %+v", i, core)
		}
	}
}

func TestMatchSystemLaunchCoresWithoutMatches(t *testing.T) {
	t.Parallel()

	system, err := GetSystem("NES")
	if err != nil {
		t.Fatal(err)
	}

	got := matchSystemLaunchCores([]RBFInfo{ParseRBF("/media/fat/_Console/SNES_20240310.rbf")}, nil, system)
	if len(got) != 0 {
		t.Fatalf("expected no results, got %+v", got)
	}
}

const (
	raNESLauncher = "<mistergamedescription>\n\t<rbf>_RA_Cores/Cores/NES</rbf>\n" +
		"\t<setname same_dir=\"1\">RA_NES</setname>\n</mistergamedescription>"
	raSNESLauncher = "<mistergamedescription><rbf>_RA_Cores/Cores/SNES</rbf>" +
		"<setname same_dir=\"1\">RA_SNES</setname></mistergamedescription>"
	datedNESLauncher = "<mistergamedescription><rbf>_Console/NES_20240310</rbf>" +
		"<setname>NES_Dated</setname></mistergamedescription>"
	plainNESLauncher = "<mistergamedescription><rbf>_Console/NES</rbf></mistergamedescription>"
	nesGameLauncher  = "<mistergamedescription><rbf>_Console/NES</rbf>" +
		"<file delay=\"1\" type=\"f\" index=\"0\" path=\"x.nes\"/></mistergamedescription>"
)

func TestMatchSystemLaunchCoresIncludesLaunchersForTheCore(t *testing.T) {
	t.Parallel()

	system, err := GetSystem("NES")
	if err != nil {
		t.Fatal(err)
	}

	rbfFiles := []RBFInfo{ParseRBF("/media/fat/_Console/NES_20240310.rbf")}
	launchers := []LaunchCore{
		mustParseLauncher(t, "/media/fat/_RA_Cores/SNES.mgl", raSNESLauncher),
		mustParseLauncher(t, "/media/fat/_RA_Cores/NES.mgl", raNESLauncher),
		mustParseLauncher(t, "/media/fat/_Other/NES dated.mgl", datedNESLauncher),
		mustParseLauncher(t, "/media/fat/_Other/NESMusic.mgl",
			"<mistergamedescription><rbf>_Console/NESMusic</rbf></mistergamedescription>"),
	}

	got := matchSystemLaunchCores(rbfFiles, launchers, system)
	if len(got) != 3 {
		t.Fatalf("got %d results, want 3: %+v", len(got), got)
	}
	if got[0].Launch != "_Console/NES" || got[0].IsLauncher() {
		t.Fatalf("core file should come first: %+v", got[0])
	}

	ra := got[1]
	if ra.Launch != "_Other/NES dated.mgl" || ra.Name != "NES_Dated" || ra.SetNameSameDir {
		t.Fatalf("dated launcher: %+v", ra)
	}
	ra = got[2]
	if ra.Launch != "_RA_Cores/NES.mgl" || ra.RBF != "_RA_Cores/Cores/NES" ||
		ra.SetName != "RA_NES" || !ra.SetNameSameDir || ra.Name != "RA_NES" || !ra.IsLauncher() {
		t.Fatalf("RA launcher: %+v", ra)
	}
}

func TestParseMGLLauncherSkipsGameLaunchers(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		data string
		want bool
	}{
		{"core only", plainNESLauncher, true},
		{"with game", nesGameLauncher, false},
		{"no rbf", "<mistergamedescription><setname>X</setname></mistergamedescription>", false},
		{"not xml", "hello", false},
	}
	for _, tt := range tests {
		_, ok := parseMGLLauncher("/media/fat/_Other/"+tt.name+".mgl", []byte(tt.data))
		if ok != tt.want {
			t.Fatalf("%s: got %v, want %v", tt.name, ok, tt.want)
		}
	}

	core, ok := parseMGLLauncher("/media/fat/_Other/Plain.mgl", []byte(plainNESLauncher))
	if !ok || core.Name != "Plain" || core.SetName != "" || core.Launch != "_Other/Plain.mgl" {
		t.Fatalf("launcher without setname: %+v", core)
	}
}

func TestMatchMGLLauncherByRelativeAndAbsolutePath(t *testing.T) {
	t.Parallel()

	launchers := []LaunchCore{
		mustParseLauncher(t, "/media/fat/_RA_Cores/NES.mgl", raNESLauncher),
	}

	tests := []struct {
		name  string
		found bool
	}{
		{"_RA_Cores/NES.mgl", true},
		{"_ra_cores/nes.MGL", true},
		{"/media/fat/_RA_Cores/NES.mgl", true},
		{"NES.mgl", false},
		{"../_RA_Cores/NES.mgl", false},
		{"", false},
	}
	for _, tt := range tests {
		_, ok := matchMGLLauncher(launchers, tt.name)
		if ok != tt.found {
			t.Fatalf("%q: found %v, want %v", tt.name, ok, tt.found)
		}
	}
}

func TestLauncherSetNamesCoverEverySystemOnTheCore(t *testing.T) {
	t.Parallel()

	launchers := []LaunchCore{
		mustParseLauncher(t, "/media/fat/_RA_Cores/NES.mgl", raNESLauncher),
		mustParseLauncher(t, "/media/fat/_Other/Plain.mgl", plainNESLauncher),
		mustParseLauncher(t, "/media/fat/_RA_Cores/SNES.mgl", raSNESLauncher),
	}

	got := setNamesOfLaunchers(launchers)

	want := map[string][]string{"RA_NES": {"FDS", "NES", "NESMusic"}, "RA_SNES": {"SNES", "SNESMusic"}}
	seen := make(map[string][]string)
	for _, entry := range got {
		seen[entry.SetName] = append(seen[entry.SetName], entry.System.Id)
	}
	for setName, systems := range want {
		if strings.Join(seen[setName], ",") != strings.Join(systems, ",") {
			t.Fatalf("%s: got %v, want %v", setName, seen[setName], systems)
		}
	}
	if len(seen) != len(want) {
		t.Fatalf("unexpected setnames: %v", seen)
	}
}

func mustParseLauncher(t *testing.T, path, data string) LaunchCore {
	t.Helper()
	core, ok := parseMGLLauncher(path, []byte(data))
	if !ok {
		t.Fatalf("%s did not parse as a core launcher", path)
	}
	return core
}
