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

package systems

import (
	"testing"

	"github.com/wizzomafizzo/mrext/pkg/games"
)

func TestSystemRBFResultsMarksConfiguredDefault(t *testing.T) {
	t.Parallel()

	cores := []games.LaunchCore{
		{
			Name:     "Custom",
			Launch:   "_Console/Custom",
			RBF:      "_Console/Custom",
			Path:     "/media/fat/_Console/Custom_20240310.rbf",
			Filename: "Custom_20240310.rbf",
		},
		{
			Name:     "NES",
			Launch:   "_Console/NES",
			RBF:      "_Console/NES",
			Path:     "/media/fat/_Console/NES_20240310.rbf",
			Filename: "NES_20240310.rbf",
		},
	}

	got := systemRBFResults(cores, "_Console/Custom")
	if len(got) != 2 {
		t.Fatalf("got %d results, want 2", len(got))
	}
	if !got[0].Default {
		t.Fatalf("configured core is not default: %+v", got[0])
	}
	if got[1].Default {
		t.Fatalf("catalog core is unexpectedly default: %+v", got[1])
	}
}
