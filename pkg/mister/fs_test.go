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
	"path/filepath"
	"testing"

	"github.com/wizzomafizzo/mrext/pkg/config"
)

func TestResolvePath(t *testing.T) {
	t.Parallel()

	absolute := filepath.Join(string(filepath.Separator), "tmp", "game.rom")
	if got := ResolvePath(absolute); got != absolute {
		t.Fatalf("ResolvePath(%q) = %q, want unchanged absolute path", absolute, got)
	}

	relative := filepath.Join("games", "NES", "game.nes")
	want := filepath.Join(config.SdFolder, relative)
	if got := ResolvePath(relative); got != want {
		t.Fatalf("ResolvePath(%q) = %q, want %q", relative, got, want)
	}
}
