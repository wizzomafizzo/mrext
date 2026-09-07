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
	"strconv"
	"strings"
	"testing"
	"time"
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

func TestKillIgnoresAProcessThatIsNotItsDaemon(t *testing.T) {
	// Escalating to SIGKILL must be as careful as Stop is about PID reuse:
	// the recorded PID may since belong to something else entirely.
	svc := &Service{
		Name:   "mrext-kill-" + filepath.Base(t.TempDir()),
		Logger: NewLogger("mrext-kill-test"),
	}

	if err := svc.kill(); err != nil {
		t.Fatalf("kill without a PID file should be a no-op, got %v", err)
	}

	// This test's own PID is live but is not the service daemon.
	if err := os.WriteFile(svc.pidFilePath(), []byte(strconv.Itoa(os.Getpid())), 0o600); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Remove(svc.pidFilePath()) })

	if err := svc.kill(); err != nil {
		t.Fatalf("kill on a foreign PID should be a no-op, got %v", err)
	}
	// Still here, so nothing signalled us.
}

func TestRestartWaitIsBounded(t *testing.T) {
	// The wait used to be an unbounded "for s.Running() { sleep }".
	if stopTimeout <= 0 {
		t.Fatal("stopTimeout must bound the wait")
	}
	previous := stopTimeout
	stopTimeout = 10 * time.Millisecond
	t.Cleanup(func() { stopTimeout = previous })

	// Not running, so Restart goes straight to Start. Start fails here because
	// the test binary is not a service; the point is that it returns at all.
	svc := &Service{
		Name:   "mrext-restart-" + filepath.Base(t.TempDir()),
		Logger: NewLogger("mrext-restart-test"),
	}
	done := make(chan error, 1)
	go func() { done <- svc.Restart() }()
	select {
	case <-done:
	case <-time.After(5 * time.Second):
		t.Fatal("Restart did not return")
	}
}
