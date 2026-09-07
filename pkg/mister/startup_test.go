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
package mister

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func startupFixture(t *testing.T, contents string) (startup *Startup, path string) {
	t.Helper()
	path = filepath.Join(t.TempDir(), "user-startup.sh")
	if contents != "" {
		if err := os.WriteFile(path, []byte(contents), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	startup = &Startup{Path: path}
	if err := startup.Load(); err != nil {
		t.Fatal(err)
	}
	return startup, path
}

func TestSaveKeepsTheExistingShebang(t *testing.T) {
	// MiSTer's own docs use #!/bin/bash, and the entries written here use
	// bash-only [[ ]] tests. Rewriting the line to #!/bin/sh whenever any
	// mrext app touched the file could stop a user's script working.
	startup, path := startupFixture(t, "#!/bin/bash\n\n# existing\n/media/fat/Scripts/other.sh\n")

	if err := startup.AddService("mrext/remote"); err != nil {
		t.Fatal(err)
	}
	if err := startup.Save(); err != nil {
		t.Fatal(err)
	}

	saved := readStartup(t, path)
	if !strings.HasPrefix(saved, "#!/bin/bash\n") {
		t.Errorf("shebang was rewritten:\n%s", saved)
	}
	if !strings.Contains(saved, "/media/fat/Scripts/other.sh") {
		t.Errorf("an unrelated entry was lost:\n%s", saved)
	}
}

func TestSaveWritesTheDefaultShebangForANewFile(t *testing.T) {
	startup, path := startupFixture(t, "")

	if err := startup.AddService("mrext/remote"); err != nil {
		t.Fatal(err)
	}
	if err := startup.Save(); err != nil {
		t.Fatal(err)
	}

	if saved := readStartup(t, path); !strings.HasPrefix(saved, defaultShebang+"\n") {
		t.Errorf("new file shebang = %q, want %q", saved, defaultShebang)
	}
}

func TestRemovingTheLastEntrySaves(t *testing.T) {
	// Uninstalling the only mrext app leaves no entries. Refusing to save left
	// the entry in the file while reporting a failure, so the service came
	// back on the next boot.
	startup, path := startupFixture(t, "#!/bin/sh\n\n# mrext/remote\n/media/fat/Scripts/remote.sh -service $1\n")

	if !startup.Exists("mrext/remote") {
		t.Fatal("fixture entry was not parsed")
	}
	if err := startup.Remove("mrext/remote"); err != nil {
		t.Fatal(err)
	}
	if err := startup.Save(); err != nil {
		t.Fatalf("removing the last entry failed: %v", err)
	}

	saved := readStartup(t, path)
	if strings.Contains(saved, "remote.sh") {
		t.Errorf("entry survived removal:\n%s", saved)
	}

	reloaded := &Startup{Path: path}
	if err := reloaded.Load(); err != nil {
		t.Fatal(err)
	}
	if len(reloaded.Entries) != 0 {
		t.Errorf("reloaded %d entries, want none", len(reloaded.Entries))
	}
}

func readStartup(t *testing.T, path string) string {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	return string(data)
}
