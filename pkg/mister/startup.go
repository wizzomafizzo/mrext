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
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/wizzomafizzo/mrext/pkg/config"
)

// TODO: delete entry from startup
// TODO: enable/disable entry in startup

type Startup struct {
	Entries []StartupEntry
}

type StartupEntry struct {
	Name    string
	Cmds    []string
	Enabled bool
}

func (s *Startup) Load() error {
	var entries []StartupEntry

	contents, err := os.ReadFile(config.StartupFile)
	if os.IsNotExist(err) {
		contents = []byte{}
	} else if err != nil {
		return fmt.Errorf("read startup script: %w", err)
	}

	lines := strings.Split(string(contents), "\n")
	sections := make([][]string, 0)

	section := make([]string, 0)
	for i, line := range lines {
		if i == 0 && strings.HasPrefix(line, "#!") {
			continue
		}

		if line == "" && len(section) != 0 {
			sections = append(sections, section)
			section = make([]string, 0)
			continue
		}
		if line != "" {
			section = append(section, line)
		}
	}

	for _, section := range sections {
		name := ""
		cmds := make([]string, 0)
		enabled := false

		if section[0] != "" && section[0][0] == '#' {
			name = strings.TrimSpace(section[0][1:])
			cmds = append(cmds, section[1:]...)
		} else {
			cmds = append(cmds, section...)
		}

		for _, line := range cmds {
			if line != "" && line[0] != '#' {
				enabled = true
				break
			}
		}

		if len(cmds) != 0 {
			entries = append(entries, StartupEntry{
				Name:    name,
				Enabled: enabled,
				Cmds:    cmds,
			})
		}
	}

	s.Entries = entries

	return nil
}

func (s *Startup) Save() error {
	if len(s.Entries) == 0 {
		return errors.New("no startup entries to save")
	}

	var contents strings.Builder
	_, _ = contents.WriteString("#!/bin/sh\n\n")
	for i := range s.Entries {
		entry := &s.Entries[i]
		if entry.Name != "" {
			_, _ = fmt.Fprintf(&contents, "# %s\n", entry.Name)
		}
		for _, cmd := range entry.Cmds {
			_, _ = fmt.Fprintln(&contents, cmd)
		}
		_ = contents.WriteByte('\n')
	}

	// #nosec G306 -- MiSTer startup script must remain readable by system tooling.
	if err := os.WriteFile(config.StartupFile, []byte(contents.String()), 0o644); err != nil {
		return fmt.Errorf("write startup script: %w", err)
	}
	return nil
}

func (s *Startup) Exists(name string) bool {
	for _, entry := range s.Entries {
		if entry.Name == name {
			return true
		}
	}

	return false
}

func (s *Startup) Add(name, cmd string) error {
	if s.Exists(name) {
		return fmt.Errorf("startup entry already exists: %s", name)
	}

	s.Entries = append(s.Entries, StartupEntry{
		Name:    name,
		Enabled: true,
		Cmds:    strings.Split(cmd, "\n"),
	})

	return nil
}

func (s *Startup) AddService(name string) error {
	path, err := os.Executable()
	if err != nil {
		return fmt.Errorf("resolve service executable: %w", err)
	}

	cmd := fmt.Sprintf("[[ -e %q ]] && %q -service $1", path, path)

	return s.Add(name, cmd)
}

func (s *Startup) Remove(name string) error {
	for i, entry := range s.Entries {
		if entry.Name == name {
			s.Entries = append(s.Entries[:i], s.Entries[i+1:]...)
			return nil
		}
	}

	return fmt.Errorf("startup entry not found: %s", name)
}
