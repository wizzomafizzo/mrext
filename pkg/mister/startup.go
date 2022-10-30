package mister

import (
	"fmt"
	"os"
	"strings"

	"github.com/wizzomafizzo/mrext/pkg/config"
)

// TODO: delete entry from startup
// TODO: enable/disable entry in startup
// TODO: check if service is running

type Startup struct {
	Entries []StartupEntry
}

type StartupEntry struct {
	Name    string
	Enabled bool
	Cmds    []string
}

func (s *Startup) Load() error {
	var entries []StartupEntry

	if _, err := os.Stat(config.StartupFile); err != nil {
		return err
	}

	contents, err := os.ReadFile(config.StartupFile)
	if err != nil {
		return err
	}

	lines := strings.Split(string(contents), "\n")
	sections := make([][]string, 0)

	section := make([]string, 0)
	for i, line := range lines {
		if i == 0 && strings.HasPrefix(line, "#!") {
			continue
		}

		if len(line) == 0 && len(section) != 0 {
			sections = append(sections, section)
			section = make([]string, 0)
		} else if len(line) > 0 {
			section = append(section, line)
		}
	}

	for _, section := range sections {
		name := ""
		cmds := make([]string, 0)
		enabled := false

		if len(section[0]) > 0 && section[0][0] == '#' {
			name = strings.TrimSpace(section[0][1:])
			cmds = append(cmds, section[1:]...)
		} else {
			cmds = append(cmds, section...)
		}

		for _, line := range cmds {
			if len(line) > 0 && line[0] != '#' {
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

	if len(entries) > 0 {
		s.Entries = entries
	} else {
		return fmt.Errorf("no startup entries found")
	}

	return nil
}

func (s *Startup) Save() error {
	if len(s.Entries) == 0 {
		return fmt.Errorf("no startup entries to save")
	}

	contents := "#!/bin/sh\n\n"

	for _, entry := range s.Entries {
		if len(entry.Name) != 0 {
			contents += "# " + entry.Name + "\n"
		}

		for _, cmd := range entry.Cmds {
			contents += cmd + "\n"
		}

		contents += "\n"
	}

	return os.WriteFile(config.StartupFile, []byte(contents), 0644)
}

func (s *Startup) Exists(name string) bool {
	for _, entry := range s.Entries {
		if entry.Name == name {
			return true
		}
	}

	return false
}

func (s *Startup) Enable(name string) error {
	for i, entry := range s.Entries {
		if entry.Name == name && !entry.Enabled {
			s.Entries[i].Enabled = true
			for j, cmd := range entry.Cmds {
				if len(cmd) > 0 && cmd[0] == '#' {
					s.Entries[i].Cmds[j] = cmd[1:]
				}
			}

			return nil
		}
	}

	return fmt.Errorf("startup entry not found: %s", name)
}

func (s *Startup) Add(name string, cmd string) error {
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
		return err
	}

	cmd := fmt.Sprintf("[[ -e %s ]] && %s -service $1", path, path)

	return s.Add(name, cmd)
}
