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

	filter := flag.String("filter", "", "list of systems to filter (ex. gba,psx,nes)")
	ignore := flag.String("ignore", "", "list of systems to ignore (ex. tgfx16-cd)")
	noscan := flag.Bool("noscan", false, "don't index entire system (faster, but less random)")
	flag.Parse()

	filteredIds := strings.Split(*filter, ",")
	var filteredSystems []games.System
	for _, id := range filteredIds {
		system, _ := games.LookupSystem(id)
		if system != nil {
			filteredSystems = append(filteredSystems, *system)
		}
	}

	ignoredIds := strings.Split(*ignore, ",")
	var ignoredSystems []games.System
	for _, id := range ignoredIds {
		found, _ := games.LookupSystem(id)
		if found != nil {
			ignoredSystems = append(ignoredSystems, *found)
		}
	}

	systems := games.AllSystems()

	// filter systems
	if len(filteredSystems) > 0 {
		systems = filteredSystems
	}

	// ignore systems
	if len(ignoredSystems) > 0 {
		var filtered []games.System
		for _, system := range systems {
			ignore := false
			for _, ignored := range ignoredSystems {
				if system.Id == ignored.Id {
					ignore = true
					break
				}
			}
			if !ignore {
				filtered = append(filtered, system)
			}
		}
		systems = filtered
	}

	results := games.GetSystemPaths(systems)
	if len(results) == 0 {
		fmt.Println("No games folders found.")
		os.Exit(1)
	}

	// pick out the folders that actually have stuff in them
	populated := make(map[string][]string)
	for _, folder := range results {
		files, err := os.ReadDir(folder.Path)
		if err != nil {
			continue
		}
		if len(files) > 0 {
			populated[folder.System.Id] = append(populated[folder.System.Id], folder.Path)
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
				fmt.Printf("Launching %s: %s\n", system.Id, game)
				mister.LaunchGame(*system, game)
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
				mister.LaunchGame(*system, game)
				return
			}
		}
	}

	fmt.Println("No games found.")
}
