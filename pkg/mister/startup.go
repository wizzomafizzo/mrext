package mister

import (
	"os"
	"strings"

	"github.com/wizzomafizzo/mrext/pkg/config"
)

// TODO: add entry to startup
// TODO: delete entry from startup
// TODO: enable/disable entry in startup
// TODO: check if service is running

type StartupEntry struct {
	Name    string
	Enabled bool
	Cmds    string
}

func GetStartupEntries() ([]StartupEntry, error) {
	var entries []StartupEntry

	if _, err := os.Stat(config.StartupFile); err != nil {
		return nil, err
	}

	contents, err := os.ReadFile(config.StartupFile)
	if err != nil {
		return nil, err
	}

	sections := strings.Split(string(contents), "\n\n")

	for i, section := range sections {
		if len(section) == 0 {
			continue
		}

		lines := strings.Split(section, "\n")

		if i == 0 && strings.HasPrefix(lines[0], "#!") {
			continue
		}

		name := ""
		cmds := ""
		enabled := false

		if lines[0][0] == '#' {
			name = strings.TrimSpace(lines[0][1:])
			cmds = strings.Join(lines[1:], "\n")
		} else {
			cmds = strings.Join(lines, "\n")
		}

		for _, line := range strings.Split(cmds, "\n") {
			if len(line) > 0 && line[0] != '#' {
				enabled = true
				break
			}
		}

		if cmds != "" {
			entries = append(entries, StartupEntry{
				Name:    name,
				Enabled: enabled,
				Cmds:    cmds,
			})
		}
	}

	return entries, nil
}

func StartupEntryExists(name string) (bool, error) {
	entries, err := GetStartupEntries()
	if err != nil {
		return false, err
	}

	for _, entry := range entries {
		if entry.Name == name {
			return true, nil
		}
	}

	return false, nil
}
