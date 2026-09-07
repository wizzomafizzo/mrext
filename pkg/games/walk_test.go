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
	"archive/zip"
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"slices"
	"testing"
)

func TestWalkFilesMatchesLegacyScanner(t *testing.T) {
	root, external := t.TempDir(), t.TempDir()
	for _, path := range []string{filepath.Join(root, "game.nes"), filepath.Join(external, "linked.nes")} {
		if err := os.WriteFile(path, nil, 0o600); err != nil {
			t.Fatal(err)
		}
	}
	if err := os.Symlink(external, filepath.Join(root, "linked")); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(root, filepath.Join(external, "cycle")); err != nil {
		t.Fatal(err)
	}
	// #nosec G304 -- archive is created only beneath t.TempDir.
	file, err := os.Create(filepath.Join(root, "pack.zip"))
	if err != nil {
		t.Fatal(err)
	}
	archive := zip.NewWriter(file)
	for _, name := range []string{"nested/zip.nes", "ignored.sfc"} {
		if _, createErr := archive.Create(name); createErr != nil {
			t.Fatal(createErr)
		}
	}
	if closeErr := archive.Close(); closeErr != nil {
		t.Fatal(closeErr)
	}
	if closeErr := file.Close(); closeErr != nil {
		t.Fatal(closeErr)
	}
	want, err := GetFiles("NES", root)
	if err != nil {
		t.Fatal(err)
	}
	cwd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	var got []string
	walkErr := WalkFiles("NES", root, func(path string) error { got = append(got, path); return nil })
	if walkErr != nil {
		t.Fatal(walkErr)
	}
	slices.Sort(got)
	slices.Sort(want)
	if !reflect.DeepEqual(got, want) || len(got) != 3 {
		t.Fatalf("streamed paths = %v, legacy = %v", got, want)
	}
	if after, cwdErr := os.Getwd(); cwdErr != nil || after != cwd {
		t.Fatalf("scanner changed working directory: %q, %v", after, cwdErr)
	}
	failure := errors.New("stop scan")
	calls := 0
	err = WalkFiles("NES", root, func(string) error { calls++; return failure })
	if !errors.Is(err, failure) || calls != 1 {
		t.Fatalf("callback failure not propagated immediately: %v, calls=%d", err, calls)
	}
}
