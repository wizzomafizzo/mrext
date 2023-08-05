package main

import (
	"flag"
	"fmt"
	"github.com/wizzomafizzo/mrext/pkg/config"
	"github.com/wizzomafizzo/mrext/pkg/mister"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/wizzomafizzo/mrext/pkg/games"
	"github.com/wizzomafizzo/mrext/pkg/utils"
)

// Alternate names for systems.
var idMap = map[string]string{
	// "TurboGrafx16":   "tgfx16",
}

// Only allow these extensions to be indexed.
// Any systems not listed will allow all extensions.
var extMap = map[string][]string{
	// "Atari5200":    {".a52", ".car"},
}

// Convert an internal system ID to a mistercon ID if possible.
func conId(id string) string {
	if id, ok := idMap[id]; ok {
		return id
	}

	return id
}

// Convert a mistercon system ID to an internal ID if possible.
func reverseId(id string) string {
	for k, v := range idMap {
		if strings.EqualFold(v, id) {
			return k
		}
	}

	return id
}

// Return the filename of the gamelist for a given system ID.
func gamelistFilename(systemId string) string {
	var prefix string
	if id, ok := idMap[systemId]; ok {
		prefix = id
	} else {
		prefix = systemId
	}

	return strings.ToLower(prefix) + ".txt"
}

// Generate a gamelist file for a system with given results.
func writeGamelist(gamelistDir string, systemId string, files []string) {
	gamelistPath := filepath.Join(gamelistDir, gamelistFilename(systemId))
	tmpPath, err := os.CreateTemp("", "gamelist-*.txt")
	if err != nil {
		panic(err)
	}

	for _, file := range files {
		tmpPath.WriteString(file + "\n")
	}
	tmpPath.Sync()
	tmpPath.Close()

	err = utils.MoveFile(tmpPath.Name(), gamelistPath)
	if err != nil {
		panic(err)
	}
}

// Generate gamelists for all systems. Main workflow of app.
func createGamelists(gamelistDir string, systemPaths map[string][]string, progress bool, quiet bool, filter bool) int {
	start := time.Now()

	if !quiet && !progress {
		fmt.Println("Finding system folders...")
	}

	// prep calculating progress
	totalPaths := 0
	for _, v := range systemPaths {
		totalPaths += len(v)
	}
	totalSteps := totalPaths
	currentStep := 0

	// generate file list
	totalGames := 0
	for systemId, paths := range systemPaths {
		var systemFiles []string

		for _, path := range paths {
			if !quiet {
				if progress {
					fmt.Println("XXX")
					fmt.Println(int(float64(currentStep) / float64(totalSteps) * 100))
					fmt.Printf("Scanning %s (%s)\n", systemId, path)
					fmt.Println("XXX")
				} else {
					fmt.Printf("Scanning %s: %s\n", systemId, path)
				}
			}

			files, err := games.GetFiles(systemId, path)
			if err != nil {
				log.Println(err)
				continue
			}
			systemFiles = append(systemFiles, files...)

			currentStep++
		}

		if filter {
			systemFiles = games.FilterUniqueFilenames(systemFiles)
		}

		// filter out certain extensions
		var filteredFiles []string
		if filterExts, ok := extMap[systemId]; ok {
			for _, file := range systemFiles {
				path := strings.ToLower(file)
				for _, ext := range filterExts {
					if strings.HasSuffix(path, ext) {
						filteredFiles = append(filteredFiles, file)
						break
					}
				}
			}
			systemFiles = filteredFiles
		}

		if len(systemFiles) > 0 {
			totalGames += len(systemFiles)
			writeGamelist(gamelistDir, systemId, systemFiles)
		}
	}

	if !quiet {
		taken := int(time.Since(start).Seconds())
		if progress {
			fmt.Println("XXX")
			fmt.Println("100")
			fmt.Printf("Indexing complete (%d games in %ds)\n", totalGames, taken)
			fmt.Println("XXX")
		} else {
			fmt.Printf("Indexing complete (%d games in %ds)\n", totalGames, taken)
		}
	}

	return totalGames
}

func tryLaunchGame(launchPath string) error {
	system, err := games.BestSystemMatch(&config.UserConfig{}, launchPath)
	if err != nil {
		return fmt.Errorf("error during launch: %s", err)
	}

	err = mister.LaunchGame(system, launchPath)
	if err != nil {
		return fmt.Errorf("error during launch: %s", err)
	}

	return nil
}

func main() {
	gamelistDir := flag.String("out", ".", "gamelist files directory")
	filter := flag.String("filter", "all", "list of systems to index (comma separated)")
	progress := flag.Bool("progress", false, "print output for dialog gauge")
	quiet := flag.Bool("quiet", false, "suppress all status output")
	detect := flag.Bool("detect", false, "list active system folders")
	noDupes := flag.Bool("nodupes", false, "filter out duplicate games")
	launchPath := flag.String("launch", "", "launch game with given path")
	flag.Parse()

	// launch game
	if *launchPath != "" {
		err := tryLaunchGame(*launchPath)
		if err != nil {
			fmt.Println("Error launching game:", err)
			os.Exit(1)
		}
		os.Exit(0)
	}

	// filter systems
	var systems []games.System
	if *filter == "all" {
		systems = games.AllSystems()
	} else {
		for _, filterId := range strings.Split(*filter, ",") {
			systemId := reverseId(filterId)

			if system, ok := games.Systems[systemId]; ok {
				systems = append(systems, system)
				continue
			}

			system, err := games.LookupSystem(systemId)
			if err != nil {
				continue
			}

			systems = append(systems, *system)
		}
	}

	// find active system paths
	if *detect {
		results := games.GetActiveSystemPaths(&config.UserConfig{}, systems)
		for _, r := range results {
			fmt.Printf("%s:%s\n", strings.ToLower(conId(r.System.Id)), r.Path)
		}
		os.Exit(0)
	}

	systemPaths := games.GetSystemPaths(&config.UserConfig{}, systems)
	systemPathsMap := make(map[string][]string)

	for _, p := range systemPaths {
		systemPathsMap[p.System.Id] = append(systemPathsMap[p.System.Id], p.Path)
	}

	total := createGamelists(*gamelistDir, systemPathsMap, *progress, *quiet, *noDupes)

	if total == 0 {
		os.Exit(8)
	} else {
		os.Exit(0)
	}
}
