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

package version

import (
	"strings"
	"testing"
)

func TestStringPrefersStampedValues(t *testing.T) {
	previousVersion, previousCommit := Version, Commit
	t.Cleanup(func() { Version, Commit = previousVersion, previousCommit })

	Version, Commit = "v1.4.2", "8839ae1"
	if got := String(); got != "v1.4.2 (8839ae1)" {
		t.Errorf("String() = %q", got)
	}
	if got := Short(); got != "v1.4.2" {
		t.Errorf("Short() = %q", got)
	}

	// Without a commit the version still stands on its own.
	Commit = ""
	if got := String(); !strings.HasPrefix(got, "v1.4.2") {
		t.Errorf("String() without a commit = %q", got)
	}
}

func TestStringFallsBackToBuildInfo(t *testing.T) {
	previousVersion, previousCommit := Version, Commit
	t.Cleanup(func() { Version, Commit = previousVersion, previousCommit })

	// A plain "go build" leaves these empty; Go's own VCS stamps fill in, and
	// failing that the output is still something a user can report.
	Version, Commit = "", ""
	if got := String(); got == "" {
		t.Error("String() must never be empty")
	}
	if got := Short(); got == "" {
		t.Error("Short() must never be empty")
	}
}
