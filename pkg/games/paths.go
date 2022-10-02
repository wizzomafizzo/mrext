package games

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/wizzomafizzo/mrext/pkg/config"
)

type listDirFn func(string) ([]fs.DirEntry, error)

func memoListDir() listDirFn {
	cache := make(map[string][]fs.DirEntry)

	return func(path string) ([]fs.DirEntry, error) {
		if files, ok := cache[path]; ok {
			return files, nil
		}

		files, err := os.ReadDir(path)
		if err != nil {
			return nil, err
		}

		cache[path] = files
		return files, nil
	}
}

func getCaseInsensitiveDir(fn listDirFn, path string) (string, error) {
	if f, err := os.Stat(path); err == nil {
		if f.IsDir() {
			return path, nil
		} else {
			return "", fmt.Errorf("not a directory: %s", path)
		}
	}

	parent := filepath.Dir(path)
	files, err := fn(parent)
	if err != nil {
		return "", err
	}

	name := filepath.Base(path)
	for _, file := range files {
		if strings.EqualFold(file.Name(), name) {
			return filepath.Join(parent, file.Name()), nil
		}
	}

	return "", fmt.Errorf("directory not found: %s", path)
}

// Given any path, return what systems it could be for.
func FolderToSystems(path string) []System {
	var systems []System
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

	for _, system := range Systems {
		for _, folder := range system.Folder {
			systemPath := strings.ToLower(filepath.Join(gamesFolder, folder))
			if strings.HasPrefix(path, systemPath) {
				systems = append(systems, system)
				break
			}
		}
	}

	return systems
}

type PathResult struct {
	System System
	Path   string
}

// Return all possible paths for each system.
func GetSystemPaths(systems []System) []PathResult {
	var matches []PathResult
	listFolder := memoListDir()

	for _, system := range systems {
		for _, gamesFolder := range config.GamesFolders {
			gf, err := getCaseInsensitiveDir(listFolder, gamesFolder)
			if err != nil {
				continue
			}

			for _, folder := range system.Folder {
				systemFolder := filepath.Join(gf, folder)
				path, err := getCaseInsensitiveDir(listFolder, systemFolder)
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

// Return the active path for each system.
func GetActiveSystemPaths(systems []System) []PathResult {
	var matches []PathResult
	listFolder := memoListDir()

	for _, system := range systems {
		for _, gamesFolder := range config.GamesFolders {
			gf, err := getCaseInsensitiveDir(listFolder, gamesFolder)
			if err != nil {
				continue
			}

			found := false

			for _, folder := range system.Folder {
				systemFolder := filepath.Join(gf, folder)
				path, err := getCaseInsensitiveDir(listFolder, systemFolder)
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
