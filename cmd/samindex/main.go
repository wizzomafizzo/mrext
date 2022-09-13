package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/wizzomafizzo/mrext/pkg/games"
	"github.com/wizzomafizzo/mrext/pkg/utils"
)

var idMap = map[string]string{
	"Gameboy":         "gb",
	"GameboyColor":    "gbc",
	"GameGear":        "gg",
	"MasterSystem":    "sms",
	"Sega32X":         "s32x",
	"TurboGraphx16":   "tgfx16",
	"TurboGraphx16CD": "tgfx16cd",
}

func reverseId(id string) string {
	for k, v := range idMap {
		if strings.EqualFold(v, id) {
			return k
		}
	}

	return id
}

func gamelistFilename(systemId string) string {
	var prefix string
	if id, ok := idMap[systemId]; ok {
		prefix = id
	} else {
		prefix = systemId
	}

	return strings.ToLower(prefix) + "_gamelist.txt"
}

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

func createGamelists(gamelistDir string, systemPaths map[string][]string, progress bool, quiet bool, filter bool) {
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
}

func main() {
	gamelistDir := flag.String("o", ".", "gamelist files directory")
	filter := flag.String("s", "all", "list of systems to index (comma separated)")
	progress := flag.Bool("p", false, "print output for dialog gauge")
	quiet := flag.Bool("q", false, "suppress all output")
	detect := flag.Bool("d", false, "list system folders")
	noFilter := flag.Bool("nofilter", false, "don't filter out duplicate games")
	flag.Parse()

	systemPaths := games.GetSystemPaths()

	// filter systems if required
	filteredPaths := make(map[string][]string)
	if *filter == "all" {
		filteredPaths = systemPaths
	} else {
		filteredSystems := strings.Split(*filter, ",")
		for _, system := range filteredSystems {
			for systemId, paths := range systemPaths {
				if strings.EqualFold(system, systemId) {
					filteredPaths[systemId] = paths
				}
			}
			// also support sam's system ids
			for origId, samId := range idMap {
				if strings.EqualFold(system, samId) {
					filteredPaths[origId] = systemPaths[origId]
				}
			}
		}
	}

	if *detect {
		for systemId, paths := range filteredPaths {
			for _, path := range paths {
				files, err := os.ReadDir(path)
				if err != nil {
					continue
				}

				if len(files) > 0 {
					fmt.Printf("%s:%s\n", strings.ToLower(reverseId(systemId)), path)
				}
			}
		}
		return
	}

	createGamelists(*gamelistDir, filteredPaths, *progress, *quiet, !*noFilter)
}
