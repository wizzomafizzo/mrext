package games

import (
	"fmt"
	"github.com/wizzomafizzo/mrext/pkg/config"
	"os"
	"path/filepath"
	"strings"
)

func FindFile(path string) (string, error) {
	if _, err := os.Stat(path); err == nil {
		return path, nil
	}

	parent := filepath.Dir(path)
	name := filepath.Base(path)

	files, err := os.ReadDir(parent)
	if err != nil {
		return "", err
	}

	for _, file := range files {
		target := file.Name()

		if len(target) != len(name) {
			continue
		} else if strings.EqualFold(target, name) {
			return filepath.Join(parent, target), nil
		}
	}

	return "", fmt.Errorf("file match not found: %s", path)
}

// FolderToSystems returns what systems a path could be for.
func FolderToSystems(path string) []System {
	path = strings.ToLower(path)
	validGamesFolder := false
	gamesFolder := ""

	for _, folder := range config.GamesFolders {
		if strings.HasPrefix(path, strings.ToLower(folder)) {
			validGamesFolder = true
			gamesFolder = folder
			break
		}
	}

	if !validGamesFolder {
		return nil
	}

	var validSystems []System
	for _, system := range Systems {
		for _, folder := range system.Folder {
			systemPath := strings.ToLower(filepath.Join(gamesFolder, folder))
			if strings.HasPrefix(path, systemPath) {
				validSystems = append(validSystems, system)
				break
			}
		}
	}

	var matchedExtensions []System
	for _, system := range validSystems {
		if MatchSystemFile(system, path) {
			matchedExtensions = append(matchedExtensions, system)
		}
	}

	return matchedExtensions
}

func BestSystemMatch(path string) (System, error) {
	systems := FolderToSystems(path)

	if len(systems) == 0 {
		return System{}, fmt.Errorf("no systems found for %s", path)
	}

	if len(systems) == 1 {
		return systems[0], nil
	}

	// prefer the system with a setname
	for _, system := range systems {
		if system.SetName != "" {
			return system, nil
		}
	}

	// otherwise just return the first one
	return systems[0], nil
}

type PathResult struct {
	System System
	Path   string
}

// GetSystemPaths returns all possible paths for each system.
func GetSystemPaths(systems []System) []PathResult {
	var matches []PathResult

	for _, system := range systems {
		for _, gamesFolder := range config.GamesFolders {
			gf, err := FindFile(gamesFolder)
			if err != nil {
				continue
			}

			for _, folder := range system.Folder {
				systemFolder := filepath.Join(gf, folder)
				path, err := FindFile(systemFolder)
				if err != nil {
					continue
				}

				matches = append(matches, PathResult{system, path})
			}
		}
	}

	return matches
}

func GetAllSystemPaths() []PathResult {
	return GetSystemPaths(AllSystems())
}

// GetActiveSystemPaths returns the active path for each system.
func GetActiveSystemPaths(systems []System) []PathResult {
	var matches []PathResult

	for _, system := range systems {
		for _, gamesFolder := range config.GamesFolders {
			gf, err := FindFile(gamesFolder)
			if err != nil {
				continue
			}

			found := false

			for _, folder := range system.Folder {
				systemFolder := filepath.Join(gf, folder)
				path, err := FindFile(systemFolder)
				if err != nil {
					continue
				}

				matches = append(matches, PathResult{system, path})
				found = true
				break
			}

			if found {
				break
			}
		}

		if len(matches) == len(systems) {
			break
		}
	}

	return matches
}
