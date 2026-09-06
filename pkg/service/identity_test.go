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

//go:build linux

//nolint:gosec // Tests use only temporary filesystem fixtures.
package service

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestDaemonIdentityRejectsReusedPID(t *testing.T) {
	root := t.TempDir()
	proc := filepath.Join(root, "42")
	if err := os.Mkdir(proc, 0o700); err != nil {
		t.Fatal(err)
	}
	executable := filepath.Join(root, "remote.sh")
	link := filepath.Join(proc, "exe")
	if err := os.Symlink(executable, link); err != nil {
		t.Fatal(err)
	}
	cmdline := filepath.Join(proc, "cmdline")
	if err := os.WriteFile(cmdline, []byte("remote.sh\x00-service\x00exec\x00&\x00"), 0o600); err != nil {
		t.Fatal(err)
	}
	if !daemonIdentity(root, 42, executable) {
		t.Fatal("valid daemon was rejected")
	}
	if daemonIdentity(root, 42, filepath.Join(root, "playlog.sh")) {
		t.Fatal("different executable was treated as Remote")
	}
	if err := os.WriteFile(cmdline, []byte("remote.sh\x00-service\x00start\x00"), 0o600); err != nil {
		t.Fatal(err)
	}
	if daemonIdentity(root, 42, executable) {
		t.Fatal("launcher process was treated as daemon")
	}
	if err := os.Remove(cmdline); err != nil {
		t.Fatal(err)
	}
	if daemonIdentity(root, 42, executable) || daemonIdentity(root, 99, executable) {
		t.Fatal("unverifiable process was accepted")
	}
}

func TestServicePIDValidation(t *testing.T) {
	path := filepath.Join(t.TempDir(), "remote.pid")
	if pid, err := readServicePID(path); err != nil || pid != 0 {
		t.Fatalf("missing PID: %d %v", pid, err)
	}
	for _, value := range []string{"", "0", "1", "-42", "not a PID"} {
		if err := os.WriteFile(path, []byte(value), 0o600); err != nil {
			t.Fatal(err)
		}
		if _, err := readServicePID(path); err == nil {
			t.Fatalf("unsafe PID accepted: %q", value)
		}
	}
	if err := os.WriteFile(path, []byte("42\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	if pid, err := readServicePID(path); err != nil || pid != 42 {
		t.Fatalf("valid legacy PID: %d %v", pid, err)
	}
}

func TestStopReportsNotRunningWithoutPIDFile(t *testing.T) {
	svc := &Service{Name: "mrext-stop-" + filepath.Base(t.TempDir())}
	err := svc.Stop()
	if err == nil || !strings.Contains(err.Error(), "service not running") {
		t.Fatalf("expected not-running error, got %v", err)
	}
}
