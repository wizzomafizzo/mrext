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

//nolint:gosec // Tests operate only on paths rooted in temporary directories.
package bgm

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"
)

// newTestPaths builds a rooted layout with a short socket path, because unix
// socket paths are limited to 108 bytes and t.TempDir names are long.
func newTestPaths(t *testing.T) Paths {
	t.Helper()
	root := t.TempDir()
	paths := RootedPaths(root)
	folders := []string{
		paths.MusicFolder, paths.TempFolder, filepath.Dir(paths.CmdInterface), filepath.Dir(paths.StartupScript),
	}
	for _, folder := range folders {
		if err := os.MkdirAll(folder, 0o755); err != nil {
			t.Fatal(err)
		}
	}
	socketDir, err := os.MkdirTemp("", "bgm")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.RemoveAll(socketDir) })
	paths.SocketFile = filepath.Join(socketDir, SocketFilename)
	if len(paths.SocketFile) >= 100 {
		t.Fatalf("socket path too long for unix sockets: %s", paths.SocketFile)
	}
	return paths
}

func writeFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

func readFile(t *testing.T, path string) string {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	return string(data)
}

func writeINI(t *testing.T, paths *Paths, content string) {
	t.Helper()
	writeFile(t, paths.IniFile, content)
}

// syncBuffer captures log output written from several goroutines.
type syncBuffer struct {
	buf strings.Builder
	mu  sync.Mutex
}

func (b *syncBuffer) Write(data []byte) (int, error) {
	b.mu.Lock()
	defer b.mu.Unlock()
	count, err := b.buf.Write(data)
	if err != nil {
		return count, fmt.Errorf("buffer write: %w", err)
	}
	return count, nil
}

func (b *syncBuffer) String() string {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.buf.String()
}

func newTestLogger(paths *Paths) (*Logger, *syncBuffer) {
	out := &syncBuffer{}
	return NewLoggerTo(paths, out), out
}

const stubScript = `#!/bin/sh
printf '%s\n' "$(basename "$0") $*" >> "$BGM_STUB_LOG"
if [ -n "$BGM_STUB_FINISH" ]; then echo "[0:03] Decoding of track finished."; fi
if [ -n "$BGM_STUB_FAIL" ]; then echo "unsupported audio encoding" >&2; exit 1; fi
if [ -n "$BGM_STUB_INPUT" ]; then cat > "$BGM_STUB_INPUT"; fi
if [ -n "$BGM_STUB_EXIT" ]; then exit 0; fi
exec sleep "${BGM_STUB_SLEEP:-30}"
`

// installStubPlayers puts fake mpg123/ogg123/aplay/aplaymidi/vgmplay scripts
// first on PATH and returns the log file they append their argv to.
func installStubPlayers(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	for _, name := range []string{"mpg123", "ogg123", "aplay", "aplaymidi", "vgmplay"} {
		if err := os.WriteFile(filepath.Join(dir, name), []byte(stubScript), 0o755); err != nil {
			t.Fatal(err)
		}
	}
	logPath := filepath.Join(dir, "invocations.log")
	t.Setenv("BGM_STUB_LOG", logPath)
	t.Setenv("BGM_STUB_FINISH", "")
	t.Setenv("BGM_STUB_EXIT", "")
	t.Setenv("BGM_STUB_SLEEP", "")
	t.Setenv("BGM_STUB_INPUT", "")
	t.Setenv("BGM_STUB_FAIL", "")
	t.Setenv("PATH", dir+string(os.PathListSeparator)+os.Getenv("PATH"))
	return logPath
}

func stubInvocations(t *testing.T, logPath string) []string {
	t.Helper()
	data, err := os.ReadFile(logPath)
	if os.IsNotExist(err) {
		return nil
	}
	if err != nil {
		t.Fatal(err)
	}
	return strings.Split(strings.TrimSpace(string(data)), "\n")
}

func waitFor(t *testing.T, condition func() bool) {
	t.Helper()
	deadline := time.Now().Add(5 * time.Second)
	for !condition() {
		if time.Now().After(deadline) {
			t.Fatal("condition not met in time")
		}
		time.Sleep(5 * time.Millisecond)
	}
}

func ptr[T any](value T) *T { return &value }
