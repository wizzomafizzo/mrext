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

// Match a *top level* folder to its systems. Returns a list of pairs of
// systemId and path.
func matchSystemFolder(path string) ([][2]string, error) {
	// TODO: i think this is redundant with FolderToSystems
	var matches [][2]string

	folder, err := os.Stat(path)
	if err != nil {
		return nil, err
	}

	name := folder.Name()

	if !folder.IsDir() {
		return nil, fmt.Errorf("not a directory: %s", path)
	}

	for k, v := range Systems {
		if strings.EqualFold(name, v.Folder) {
			matches = append(matches, [2]string{k, path})
		}
	}

	if len(matches) == 0 {
		return nil, fmt.Errorf("unknown system: %s", name)
	} else {
		return matches, nil
	}
}

// Return a list of all possible parent system folders in a given
// path with their associated system ids.
func findSystemFolders(path string) [][2]string {
	var found [][2]string

	root, err := os.Stat(path)
	if err != nil || !root.IsDir() {
		return nil
	}

	folders, err := os.ReadDir(path)
	if err != nil {
		return nil
	}

	for _, folder := range folders {
		abs := filepath.Join(path, folder.Name())

		matches, err := matchSystemFolder(abs)
		if err != nil {
			continue
		} else {
			found = append(found, matches...)
		}
	}

	return found
}

func GetSystemPaths() map[string][]string {
	var paths = make(map[string][]string)

	for _, rootPath := range config.GamesFolders {
		for _, result := range findSystemFolders(rootPath) {
			paths[result[0]] = append(paths[result[0]], result[1])
		}
	}

	return paths
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
		systemPath := strings.ToLower(filepath.Join(gamesFolder, system.Folder))
		// TODO: match not prefix
		if strings.HasPrefix(path, systemPath) {
			systems = append(systems, system)
		}
	}

	return systems
}

type pathResult struct {
	System System
	Path   string
}

// Return the active path for each system.
func GetActiveSystemPaths(systems []System) []pathResult {
	var matches []pathResult

	listFolder := memoListDir()

	for _, system := range systems {
		for _, gamesFolder := range config.GamesFolders {
			gf, err := getCaseInsensitiveDir(listFolder, gamesFolder)
			if err != nil {
				continue
			}

			systemFolder := filepath.Join(gf, system.Folder)
			path, err := getCaseInsensitiveDir(listFolder, systemFolder)
			if err != nil {
				continue
			}

			matches = append(matches, pathResult{system, path})
			break
		}

		if len(matches) == len(systems) {
			break
		}
	}

	return matches
}
