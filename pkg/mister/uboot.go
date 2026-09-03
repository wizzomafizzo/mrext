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
	"os"
	"strings"

	"github.com/wizzomafizzo/mrext/pkg/config"
)

const UBootMACParam = "ethaddr"

func ReadUBootParams() (map[string]string, error) {
	params := make(map[string]string)

	data, err := os.ReadFile(config.UBootConfigFile)
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

func WriteUBootParams(params map[string]string) error {
	pairs := make([]string, 0, len(params))

	for key, value := range params {
		pairs = append(pairs, fmt.Sprintf("%s=%s", key, value))
	}

	content := strings.Join(pairs, "\n") + "\n"

	if _, err := os.Stat(config.UBootConfigFile); err == nil {
		err = os.Rename(config.UBootConfigFile, config.UBootConfigFile+".backup")
		if err != nil {
			return fmt.Errorf("back up U-Boot configuration: %w", err)
		}
	}

	// #nosec G306 -- U-Boot configuration must remain readable by MiSTer services.
	err := os.WriteFile(config.UBootConfigFile, []byte(content), 0o644)
	if err != nil {
		return fmt.Errorf("write U-Boot configuration: %w", err)
	}

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
	params, err := ReadUBootParams()
	if err != nil {
		return err
	}

	params[UBootMACParam] = newMacAddress

	return WriteUBootParams(params)
}
