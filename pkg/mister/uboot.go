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
	"fmt"
	"net"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/wizzomafizzo/mrext/pkg/config"
)

const UBootMACParam = "ethaddr"

// ubootPath is a test seam over config.UBootConfigFile so these functions can
// be exercised without touching the real boot configuration.
var ubootPath = config.UBootConfigFile

func ReadUBootParams() (map[string]string, error) {
	params := make(map[string]string)

	// #nosec G304 -- ubootPath is a fixed configuration path, overridden only by tests.
	data, err := os.ReadFile(ubootPath)
	if os.IsNotExist(err) {
		return params, nil
	}
	if err != nil {
		return params, fmt.Errorf("read U-Boot configuration: %w", err)
	}

	for _, line := range strings.Split(string(data), "\n") {
		line = strings.TrimRight(line, "\r")
		line = strings.TrimSpace(line)

		if line == "" || !strings.Contains(line, "=") {
			continue
		}

		parts := strings.SplitN(line, "=", 2)

		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])
		params[key] = value
	}

	return params, nil
}

// WriteUBootParams replaces the U-Boot configuration with the given
// parameters. Keys are written in sorted order so repeated saves produce the
// same file; map iteration order used to reshuffle the whole thing each time.
//
// Prefer UpdateUBootParam, which keeps comments and the original ordering.
func WriteUBootParams(params map[string]string) error {
	keys := make([]string, 0, len(params))
	for key := range params {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	pairs := make([]string, 0, len(keys))
	for _, key := range keys {
		pairs = append(pairs, fmt.Sprintf("%s=%s", key, params[key]))
	}

	return writeUBoot(strings.Join(pairs, "\n") + "\n")
}

// UpdateUBootParam sets one parameter, leaving every other line of the file
// exactly as it was. Rewriting from a parsed map dropped comments and every
// line without an "=", which on the file that decides how the board boots is
// not an acceptable side effect of changing a MAC address.
func UpdateUBootParam(key, value string) error {
	// #nosec G304 -- ubootPath is a fixed configuration path, overridden only by tests.
	data, err := os.ReadFile(ubootPath)
	if os.IsNotExist(err) {
		return writeUBoot(fmt.Sprintf("%s=%s\n", key, value))
	}
	if err != nil {
		return fmt.Errorf("read U-Boot configuration: %w", err)
	}

	lines := strings.Split(string(data), "\n")
	trailingNewline := len(lines) > 0 && lines[len(lines)-1] == ""
	if trailingNewline {
		lines = lines[:len(lines)-1]
	}

	replaced := false
	for i, line := range lines {
		trimmed := strings.TrimSpace(strings.TrimRight(line, "\r"))
		if trimmed == "" || !strings.Contains(trimmed, "=") {
			continue
		}
		if strings.TrimSpace(strings.SplitN(trimmed, "=", 2)[0]) != key {
			continue
		}
		lines[i] = fmt.Sprintf("%s=%s", key, value)
		replaced = true
		break
	}
	if !replaced {
		lines = append(lines, fmt.Sprintf("%s=%s", key, value))
	}

	return writeUBoot(strings.Join(lines, "\n") + "\n")
}

// writeUBoot stages the replacement beside the original and renames over it.
// The previous version renamed the original away first, so a write that then
// failed left the board with no boot configuration at all.
func writeUBoot(content string) error {
	path := ubootPath
	temporary, err := os.CreateTemp(filepath.Dir(path), "."+filepath.Base(path)+"-*")
	if err != nil {
		return fmt.Errorf("create staged U-Boot configuration: %w", err)
	}
	staged := temporary.Name()
	published := false
	defer func() {
		_ = temporary.Close()
		if !published {
			_ = os.Remove(staged)
		}
	}()

	if _, err := temporary.WriteString(content); err != nil {
		return fmt.Errorf("write staged U-Boot configuration: %w", err)
	}
	if err := temporary.Sync(); err != nil {
		return fmt.Errorf("sync staged U-Boot configuration: %w", err)
	}
	if err := temporary.Close(); err != nil {
		return fmt.Errorf("close staged U-Boot configuration: %w", err)
	}
	// Best effort: the SD card is exFAT, where the mode comes from the mount
	// and chmod reports EPERM.
	// #nosec G302 -- U-Boot configuration must stay readable by MiSTer services.
	_ = os.Chmod(staged, 0o644)

	// Keep one backup of the configuration as it was before mrext first
	// touched it. Overwriting it on every save loses the original.
	if _, err := os.Stat(path); err == nil {
		if _, backupErr := os.Stat(path + ".backup"); os.IsNotExist(backupErr) {
			// #nosec G304,G306,G703 -- path is the fixed U-Boot configuration path.
			if data, readErr := os.ReadFile(path); readErr == nil {
				_ = os.WriteFile(path+".backup", data, 0o644)
			}
		}
	}

	if err := os.Rename(staged, path); err != nil {
		return fmt.Errorf("replace U-Boot configuration: %w", err)
	}
	published = true
	return nil
}

// GetConfiguredMacAddress returns the ethernet MAC address configured in the u-boot.txt file, if available.
func GetConfiguredMacAddress() (string, error) {
	params, err := ReadUBootParams()
	if err != nil {
		return "", err
	}

	if ethAddr, ok := params[UBootMACParam]; ok {
		return ethAddr, nil
	}

	return "", nil
}

// UpdateConfiguredMacAddress updates the ethernet MAC address configured in the u-boot.txt file. Setting a new one if
// it doesn't exist, or updating the existing one. Any existing u-boot.txt arguments are preserved.
func UpdateConfiguredMacAddress(newMacAddress string) error {
	// The value arrives from an unauthenticated HTTP request and is written
	// straight into the file U-Boot reads. Without this, a value containing a
	// newline injected arbitrary boot parameters.
	address := strings.TrimSpace(newMacAddress)
	if _, err := net.ParseMAC(address); err != nil {
		return fmt.Errorf("invalid MAC address %q: %w", newMacAddress, err)
	}

	return UpdateUBootParam(UBootMACParam, address)
}
