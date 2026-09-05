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
	"context"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestStopServiceWithoutSocket(t *testing.T) {
	paths := newTestPaths(t)
	logger, out := newTestLogger(&paths)
	if err := StopService(&paths, logger); err != nil {
		t.Fatal(err)
	}
	if out.String() != "BGM service is not running\n" {
		t.Fatalf("output %q", out.String())
	}
}

func TestStopServiceRemovesStaleSocket(t *testing.T) {
	paths := newTestPaths(t)
	listener, err := (&net.ListenConfig{}).Listen(context.Background(), "unix", paths.SocketFile)
	if err != nil {
		t.Fatal(err)
	}
	unixListener, ok := listener.(*net.UnixListener)
	if !ok {
		t.Fatal("expected a unix listener")
	}
	unixListener.SetUnlinkOnClose(false)
	_ = listener.Close()
	if !SocketStale(&paths) {
		t.Fatal("closed listener must be detected as stale")
	}
	logger, out := newTestLogger(&paths)
	if err := StopService(&paths, logger); err != nil {
		t.Fatal(err)
	}
	if SocketExists(&paths) {
		t.Fatal("stale socket must be removed")
	}
	if out.String() != "BGM service is not running\n" {
		t.Fatalf("output %q", out.String())
	}
}

func TestSpawnServiceCopiesBinaryAndPassesRoot(t *testing.T) {
	paths := newTestPaths(t)
	record := filepath.Join(t.TempDir(), "args")
	script := "#!/bin/sh\nprintf '%s\\n' \"$0 $*\" > '" + record + "'\n"
	paths.AppPath = filepath.Join(t.TempDir(), "bgm.sh")
	if err := os.WriteFile(paths.AppPath, []byte(script), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := SpawnService(&paths, "/my/root"); err != nil {
		t.Fatal(err)
	}
	waitFor(t, func() bool { _, err := os.Stat(record); return err == nil })
	waitFor(t, func() bool { return strings.Contains(readFile(t, record), "exec") })
	want := ServiceBinaryPath(&paths) + " --root /my/root exec\n"
	if got := readFile(t, record); got != want {
		t.Fatalf("service argv %q want %q", got, want)
	}
}

// A restart can copy the new service while the old one is still exiting from
// the previous copy; the copy must not fail with "text file busy".
func TestCopyServiceBinaryReplacesRunningCopy(t *testing.T) {
	sleepPath, err := exec.LookPath("sleep")
	if err != nil {
		t.Skip("sleep not available")
	}
	paths := newTestPaths(t)
	target := ServiceBinaryPath(&paths)
	writeFile(t, target, readFile(t, sleepPath))
	if chmodErr := os.Chmod(target, 0o755); chmodErr != nil {
		t.Fatal(chmodErr)
	}
	running := exec.CommandContext(context.Background(), target, "30")
	if startErr := running.Start(); startErr != nil {
		t.Skipf("cannot execute from the temp folder: %v", startErr)
	}
	t.Cleanup(func() {
		_ = running.Process.Kill()
		_ = running.Wait()
	})
	script := "#!/bin/sh\nexit 0\n"
	paths.AppPath = filepath.Join(t.TempDir(), "bgm.sh")
	writeFile(t, paths.AppPath, script)
	got, err := copyServiceBinary(&paths)
	if err != nil {
		t.Fatal(err)
	}
	if got != target || readFile(t, target) != script {
		t.Fatalf("copy %q content %q", got, readFile(t, target))
	}
	info, err := os.Stat(target)
	if err != nil || info.Mode().Perm()&0o111 == 0 {
		t.Fatalf("copy must be executable: %v %v", info, err)
	}
	entries, err := os.ReadDir(paths.TempFolder)
	if err != nil {
		t.Fatal(err)
	}
	for _, entry := range entries {
		if strings.HasPrefix(entry.Name(), "."+AppFilename+"-") {
			t.Fatalf("temporary copy left behind: %s", entry.Name())
		}
	}
}

func TestWaitForSocketHelpers(t *testing.T) {
	paths := newTestPaths(t)
	if WaitForSocket(&paths, 60*time.Millisecond) {
		t.Fatal("no socket should time out")
	}
	writeFile(t, paths.SocketFile, "")
	if !WaitForSocket(&paths, time.Second) || WaitForSocketGone(&paths, 60*time.Millisecond) {
		t.Fatal("socket presence not detected")
	}
	go func() {
		time.Sleep(30 * time.Millisecond)
		_ = os.Remove(paths.SocketFile)
	}()
	if !WaitForSocketGone(&paths, time.Second) {
		t.Fatal("socket removal not detected")
	}
}
