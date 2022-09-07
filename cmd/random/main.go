package main

import (
	"flag"
	"fmt"
	"math/rand"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/wizzomafizzo/mrext/pkg/games"
	"github.com/wizzomafizzo/mrext/pkg/mister"
	"github.com/wizzomafizzo/mrext/pkg/utils"
)

const MAX_ATTEMPTS = 100

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

	file, err := utils.RandomElem(validFiles)
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
		randomZip, err := utils.RandomElem(zipFiles)
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

func init() {
	rand.Seed(time.Now().UnixNano())
}

func main() {
	// TODO: support an ini file for default values

	filter := flag.String("filter", "", "list of system folders to filter (ex. gba,psx,nes)")
	ignore := flag.String("ignore", "", "list of system folders to ignore (ex. tgfx16-cd)")
	noscan := flag.Bool("noscan", false, "don't index entire system (faster, but less random)")
	flag.Parse()

	filteredFolders := strings.Split(*filter, ",")
	ignoredFolders := strings.Split(*ignore, ",")

	folders := games.GetSystemPaths()
	if len(folders) == 0 {
		fmt.Println("No games folders found.")
		return
	}

	var filteredSystemIds []string
	for _, system := range games.Systems {
		for _, folder := range filteredFolders {
			// exception for arcade folder
			if strings.EqualFold(folder, "arcade") {
				folder = "_Arcade"
			}

			if strings.EqualFold(folder, system.Folder) {
				filteredSystemIds = append(filteredSystemIds, system.Id)
			}
		}
	}

	var ignoredSystemIds []string
	for _, system := range games.Systems {
		for _, folder := range ignoredFolders {
			// exception for arcade folder
			if strings.EqualFold(folder, "arcade") {
				folder = "_Arcade"
			}

			if strings.EqualFold(folder, system.Folder) {
				ignoredSystemIds = append(ignoredSystemIds, system.Id)
			}
		}
	}

	if *filter != "" {
		for _, systemId := range utils.MapKeys(folders) {
			if !utils.Contains(filteredSystemIds, systemId) {
				delete(folders, systemId)
			}
		}
	}

	if *ignore != "" {
		for _, systemId := range ignoredSystemIds {
			delete(folders, systemId)
		}
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

	if *noscan {
		for i := 0; i < MAX_ATTEMPTS; i++ {
			// random system
			systemId, err := utils.RandomElem(utils.MapKeys(populated))
			if err != nil {
				continue
			}

			// random folder from that system
			folder, err := utils.RandomElem(populated[systemId])
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
				mister.LaunchGame(system, game)
				return
			}
		}
	} else {
		for i := 0; i < MAX_ATTEMPTS; i++ {
			// random system
			systemId, err := utils.RandomElem(utils.MapKeys(populated))
			if err != nil {
				continue
			}

			// scan all system folders
			var files []string
			for _, path := range populated[systemId] {
				results, err := games.GetFiles(systemId, path)
				if err != nil {
					continue
				} else {
					files = append(files, results...)
				}
			}

			if len(files) == 0 {
				continue
			}

			system, err := games.GetSystem(systemId)
			if err != nil {
				continue
			}

			game, err := utils.RandomElem(files)
			if err != nil {
				continue
			} else {
				// we did it
				fmt.Printf("Launching %s: %s\n", system.Id, game)
				mister.LaunchGame(system, game)
				return
			}
		}
	}

	fmt.Println("No games found.")
}
