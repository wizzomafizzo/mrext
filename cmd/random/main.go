package main

import (
	"flag"
	"fmt"
	"github.com/wizzomafizzo/mrext/pkg/config"
	"github.com/wizzomafizzo/mrext/pkg/games"
	"github.com/wizzomafizzo/mrext/pkg/mister"
	"github.com/wizzomafizzo/mrext/pkg/utils"
	"os"
	"strings"
)

const (
	appName         = "random"
	maxPickAttempts = 100
)

func main() {
	// TODO: support an ini file for default values

	filter := flag.String("filter", "", "list of systems to filter (ex. gba,psx,nes)")
	ignore := flag.String("ignore", "", "list of systems to ignore (ex. tgfx16-cd)")
	noscan := flag.Bool("noscan", false, "don't index entire system (faster, but less random)")
	flag.Parse()

	cfg, err := config.LoadUserConfig(appName, &config.UserConfig{})
	if err != nil {
		fmt.Println("Error loading config file:", err)
		os.Exit(1)
	}

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

	results := games.GetSystemPaths(cfg, systems)
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
		for i := 0; i < maxPickAttempts; i++ {
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

			game, err := mister.TryPickRandomGame(system, folder)
			if err != nil || game == "" {
				continue
			} else {
				// we did it
				fmt.Printf("Launching %s: %s\n", system.Id, game)
				err := mister.LaunchGame(cfg, *system, game)
				if err != nil {
					fmt.Println(err)
				}
				return
			}
		}
	} else {
		for i := 0; i < maxPickAttempts; i++ {
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
				err := mister.LaunchGame(cfg, *system, game)
				if err != nil {
					fmt.Println(err)
				}
				return
			}
		}
	}

	fmt.Println("No games found.")
}
