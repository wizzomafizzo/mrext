package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/wizzomafizzo/mrext/pkg/games"
	"github.com/wizzomafizzo/mrext/pkg/utils"
)

const MAX_ATTEMPTS = 10

// Recursively search through given folder for a valid game file for that system.
func tryPickRandomGame(system *games.System, folder string) (string, error) {
	files, err := os.ReadDir(folder)
	if err != nil {
		return "", err
	}

	if len(files) == 0 {
		return "", fmt.Errorf("no files in %s", folder)
	}

	var validFiles []os.DirEntry
	for _, file := range files {
		if file.IsDir() {
			validFiles = append(validFiles, file)
		} else if utils.IsZip(file.Name()) {
			validFiles = append(validFiles, file)
		} else if games.MatchSystemFile(*system, file.Name()) {
			validFiles = append(validFiles, file)
		}
	}

	if len(validFiles) == 0 {
		return "", fmt.Errorf("no valid files in %s", folder)
	}

	file, err := utils.RandomItem(validFiles)
	if err != nil {
		return "", err
	}

	path := filepath.Join(folder, file.Name())
	if file.IsDir() {
		return tryPickRandomGame(system, path)
	} else if utils.IsZip(path) {
		// zip files
		zipFiles, err := utils.ListZip(path)
		if err != nil {
			return "", err
		}
		if len(zipFiles) == 0 {
			return "", fmt.Errorf("no files in %s", path)
		}
		// just shoot our shot on a zip instead of checking every file
		randomZip, err := utils.RandomItem(zipFiles)
		if err != nil {
			return "", err
		}
		zipPath := filepath.Join(path, randomZip)
		if games.MatchSystemFile(*system, zipPath) {
			return zipPath, nil
		} else {
			return "", fmt.Errorf("invalid file picked in %s", path)
		}
	} else {
		return path, nil
	}
}

func main() {
	folders := games.GetSystemPaths()
	if len(folders) == 0 {
		fmt.Println("No games folders found.")
		return
	}

	// pick out the folders that actually have stuff in them
	populated := make(map[string][]string)
	for systemId, folders := range folders {
		for _, folder := range folders {
			files, err := os.ReadDir(folder)
			if err != nil {
				continue
			}
			if len(files) > 0 {
				populated[systemId] = append(populated[systemId], folder)
			}
		}
	}

	if len(populated) == 0 {
		fmt.Println("No games found.")
		return
	}

	for i := 0; i < MAX_ATTEMPTS; i++ {
		// random system
		systemId, err := utils.RandomItem(utils.MapKeys(populated))
		if err != nil {
			continue
		}
		// random folder from that system
		folder, err := utils.RandomItem(populated[systemId])
		if err != nil {
			continue
		}
		// search for a random game
		system, err := games.GetSystem(systemId)
		if err != nil {
			continue
		}
		game, err := tryPickRandomGame(system, folder)
		if err != nil || game == "" {
			continue
		} else {
			// we did it
			fmt.Println(system.Id, game)
			return
		}
	}

	fmt.Println("No games found.")
}
