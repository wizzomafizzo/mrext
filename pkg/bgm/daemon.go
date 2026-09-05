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

package bgm

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"time"
)

// ServiceBinaryPath is where the running service copy lives so the installed
// bgm.sh can be replaced by an updater while music plays.
func ServiceBinaryPath(paths *Paths) string {
	return filepath.Join(paths.TempFolder, AppFilename)
}

// SpawnService mirrors `bgm.sh exec &`: start the service detached. The
// binary is first copied to the temp folder, as mrext's other daemons do.
func SpawnService(paths *Paths, root string) error {
	binary, err := copyServiceBinary(paths)
	if err != nil {
		return err
	}
	args := make([]string, 0, 3)
	if root != "" {
		args = append(args, "--root", root)
	}
	args = append(args, "exec")
	// #nosec G204 -- the executable is the copy of this program created above.
	cmd := exec.CommandContext(context.Background(), binary, args...)
	cmd.Env = os.Environ()
	cmd.SysProcAttr = detachAttributes()
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("start BGM service: %w", err)
	}
	_ = cmd.Process.Release()
	return nil
}

// copyServiceBinary writes the copy beside its final name and renames it into
// place: during a restart the previous service can still be exiting from the
// old copy, and Linux refuses to overwrite a running executable in place.
func copyServiceBinary(paths *Paths) (string, error) {
	// #nosec G304 -- the source is this program's own executable path.
	source, err := os.Open(paths.AppPath)
	if err != nil {
		return "", fmt.Errorf("open BGM binary: %w", err)
	}
	defer func() { _ = source.Close() }()
	target := ServiceBinaryPath(paths)
	destination, err := os.CreateTemp(paths.TempFolder, "."+AppFilename+"-*")
	if err != nil {
		return "", fmt.Errorf("create BGM service copy: %w", err)
	}
	temporaryPath := destination.Name()
	removeTemporary := true
	defer func() {
		_ = destination.Close()
		if removeTemporary {
			_ = os.Remove(temporaryPath)
		}
	}()
	if _, err := io.Copy(destination, source); err != nil {
		return "", fmt.Errorf("copy BGM binary: %w", err)
	}
	// #nosec G302 -- the service copy must be executable.
	if err := destination.Chmod(0o755); err != nil {
		return "", fmt.Errorf("set BGM service copy permissions: %w", err)
	}
	if err := destination.Close(); err != nil {
		return "", fmt.Errorf("close BGM service copy: %w", err)
	}
	// #nosec G703 -- destination is the fixed service copy path.
	if err := os.Rename(temporaryPath, target); err != nil {
		return "", fmt.Errorf("replace BGM service copy: %w", err)
	}
	removeTemporary = false
	return target, nil
}

// RemoveServiceBinary deletes the temp copy when this process runs from it.
func RemoveServiceBinary(paths *Paths) {
	exe, err := os.Executable()
	if err != nil {
		return
	}
	if filepath.Clean(exe) == filepath.Clean(ServiceBinaryPath(paths)) {
		_ = os.Remove(exe)
	}
}

// StopService mirrors `bgm.sh stop`: ask the service for its pid and send
// SIGTERM. A socket nobody listens on is removed instead of crashing.
func StopService(paths *Paths, logger *Logger) error {
	if !SocketExists(paths) {
		logger.Print("BGM service is not running")
		return nil
	}
	if SocketStale(paths) {
		_ = os.Remove(paths.SocketFile)
		logger.Print("BGM service is not running")
		return nil
	}
	reply, replied, err := Send(paths, "pid")
	if err != nil {
		return err
	}
	if !replied {
		return nil
	}
	pid, err := strconv.Atoi(strings.TrimSpace(reply))
	if err != nil {
		return fmt.Errorf("parse BGM service pid %q: %w", reply, err)
	}
	process, err := os.FindProcess(pid)
	if err != nil {
		return fmt.Errorf("find BGM service process: %w", err)
	}
	if err := process.Signal(syscall.SIGTERM); err != nil {
		return fmt.Errorf("stop BGM service: %w", err)
	}
	return nil
}

// WaitForSocket waits until the socket file exists.
func WaitForSocket(paths *Paths, timeout time.Duration) bool {
	return waitUntil(timeout, func() bool { return SocketExists(paths) })
}

// WaitForSocketGone waits until the socket file has been removed.
func WaitForSocketGone(paths *Paths, timeout time.Duration) bool {
	return waitUntil(timeout, func() bool { return !SocketExists(paths) })
}

func waitUntil(timeout time.Duration, condition func() bool) bool {
	deadline := time.Now().Add(timeout)
	for {
		if condition() {
			return true
		}
		if time.Now().After(deadline) {
			return false
		}
		time.Sleep(50 * time.Millisecond)
	}
}
