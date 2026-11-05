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
	"slices"
	"testing"
)

func TestLoadRemoteAllowedOrigins(t *testing.T) {
	sharedTUIFixture(t, "")
	path := filepath.Join(t.TempDir(), "remote.ini")
	contents := "[remote]\nallowed_origins = https://dashboard.example, http://localhost:5173\n"
	if err := os.WriteFile(path, []byte(contents), 0o600); err != nil {
		t.Fatal(err)
	}

	cfg, err := LoadUserConfigAt(path, "/media/fat/Scripts/remote.sh", &UserConfig{})
	if err != nil {
		t.Fatal(err)
	}
	want := []string{"https://dashboard.example", "http://localhost:5173"}
	if !slices.Equal(cfg.Remote.AllowedOrigins, want) {
		t.Fatalf("AllowedOrigins = %q, want %q", cfg.Remote.AllowedOrigins, want)
	}
}
