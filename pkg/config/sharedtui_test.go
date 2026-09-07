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

//nolint:gosec // Tests only operate on temporary fixture paths.
package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"gopkg.in/ini.v1"
)

func sharedTUIFixture(t *testing.T, contents string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "tui.ini")
	if contents != "" {
		if err := os.WriteFile(path, []byte(contents), 0o600); err != nil {
			t.Fatal(err)
		}
	}
	t.Setenv(SharedTUIConfigEnv, path)
	return path
}

func TestSharedTUIResolutionOrder(t *testing.T) {
	appINI := filepath.Join(t.TempDir(), "favorites.ini")

	cases := []struct {
		name   string
		shared string
		app    string
		want   TUIConfig
	}{
		{
			name: "neither file: built-in defaults",
			want: TUIConfig{Theme: "default", Mouse: true, CRTMode: true, OnScreenKeyboard: true},
		},
		{
			name:   "shared only",
			shared: "[tui]\ntheme = nord\nmouse = false\n",
			want:   TUIConfig{Theme: "nord", Mouse: false, CRTMode: true, OnScreenKeyboard: true},
		},
		{
			name: "app only",
			app:  "[tui]\ntheme = dracula\n",
			want: TUIConfig{Theme: "dracula", Mouse: true, CRTMode: true, OnScreenKeyboard: true},
		},
		{
			// The app INI wins per key, and only for the keys it declares:
			// crt_mode is not in the app file, so the shared value survives.
			name:   "both: app wins per key",
			shared: "[tui]\ntheme = nord\nmouse = false\ncrt_mode = false\n",
			app:    "[tui]\ntheme = gruvbox\n",
			want:   TUIConfig{Theme: "gruvbox", Mouse: false, CRTMode: false, OnScreenKeyboard: true},
		},
	}

	for _, item := range cases {
		t.Run(item.name, func(t *testing.T) {
			sharedTUIFixture(t, item.shared)
			if item.app == "" {
				_ = os.Remove(appINI)
			} else if err := os.WriteFile(appINI, []byte(item.app), 0o600); err != nil {
				t.Fatal(err)
			}

			defaults := &UserConfig{
				TUI: TUIConfig{Theme: "default", Mouse: true, CRTMode: true, OnScreenKeyboard: true},
			}
			cfg, err := LoadUserConfigAt(appINI, "/media/fat/Scripts/favorites.sh", defaults)
			if err != nil {
				t.Fatal(err)
			}
			if cfg.TUI != item.want {
				t.Fatalf("TUI = %+v, want %+v", cfg.TUI, item.want)
			}
		})
	}
}

func TestSaveSharedTUIKeepsUnrelatedContent(t *testing.T) {
	path := sharedTUIFixture(t, "; hand written\n[tui]\ntheme = nord\n\n[other]\nkeep = yes\n")

	settings := TUIConfig{Theme: "gruvbox", Mouse: true, CRTMode: false, OnScreenKeyboard: true}
	if err := SaveSharedTUI(settings); err != nil {
		t.Fatal(err)
	}

	// go-ini aligns the "=" column, so match on the pair rather than spacing.
	saved := readFileString(t, path)
	reloaded, err := ini.Load(path)
	if err != nil {
		t.Fatal(err)
	}
	tui := reloaded.Section("tui")
	if tui.Key("theme").String() != "gruvbox" || tui.Key("crt_mode").String() != "false" {
		t.Errorf("interface settings not saved:\n%s", saved)
	}
	if reloaded.Section("other").Key("keep").String() != "yes" {
		t.Errorf("unrelated section was lost:\n%s", saved)
	}
	if !strings.Contains(saved, "hand written") {
		t.Errorf("comment was lost:\n%s", saved)
	}
}

func TestRemoveLocalTUIKeysLeavesEverythingElse(t *testing.T) {
	path := filepath.Join(t.TempDir(), "favorites.ini")
	original := "[favorites]\ndefault_folder = _@Favorites\n\n[tui]\ntheme = nord\nmouse = false\n"
	if err := os.WriteFile(path, []byte(original), 0o600); err != nil {
		t.Fatal(err)
	}

	// An app INI that spells out these keys always beats the shared file, so
	// sharing has to clear them or it would never take effect.
	if err := RemoveLocalTUIKeys(path); err != nil {
		t.Fatal(err)
	}

	saved := readFileString(t, path)
	if strings.Contains(saved, "theme") || strings.Contains(saved, "mouse") {
		t.Errorf("interface keys survived:\n%s", saved)
	}
	if !strings.Contains(saved, "default_folder = _@Favorites") {
		t.Errorf("unrelated settings were lost:\n%s", saved)
	}

	// Harmless to repeat, and harmless when the file does not exist.
	if err := RemoveLocalTUIKeys(path); err != nil {
		t.Fatal(err)
	}
	if err := RemoveLocalTUIKeys(filepath.Join(t.TempDir(), "missing.ini")); err != nil {
		t.Fatal(err)
	}
}

func readFileString(t *testing.T, path string) string {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	return string(data)
}
