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

func gamelistFilename(systemId string) string {
	var prefix string
	if id, ok := idMap[systemId]; ok {
		prefix = id
	} else {
		prefix = systemId
	}

	return strings.ToLower(prefix) + "_gamelist.txt"
}

func main() {
	outDir := flag.String("o", ".", "output directory for gamelist files")
	filter := flag.String("s", "all", "list of systems to index (comma delimited)")
	progress := flag.Bool("p", false, "print output for dialog gauge")
	quiet := flag.Bool("q", false, "suppress all output")
	flag.Parse()

	start := time.Now()

	if !*quiet && !*progress {
		fmt.Println("Finding system folders...")
	}

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

	// prep calculating progress
	totalSteps := 1
	for _, v := range filteredPaths {
		totalSteps += len(v)
	}
	currentStep := 0

	// generate file list
	systemFiles := make([][2]string, 0)
	for systemId, paths := range filteredPaths {
		for _, path := range paths {
			if !*quiet {
				if *progress {
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

			files = games.FilterUniqueFilenames(files)

			for _, file := range files {
				systemFiles = append(systemFiles, [2]string{systemId, file})
			}

			currentStep++
		}
	}

	// write gamelist files to tmp
	if !*quiet {
		if *progress {
			fmt.Println("XXX")
			fmt.Println(int(float64(currentStep) / float64(totalSteps) * 100))
			fmt.Println("Creating game lists...")
			fmt.Println("XXX")
		} else {
			fmt.Println("Creating game lists...")
		}
	}
	currentStep++

	tmpDir, err := os.MkdirTemp(os.TempDir(), "sam-")
	if err != nil {
		panic(err)
	}

	gamelists := make(map[string]*os.File)
	for _, game := range systemFiles {
		systemId, path := game[0], game[1]

		if _, ok := gamelists[systemId]; !ok {
			filename := gamelistFilename(systemId)

			file, err := os.Create(filepath.Join(tmpDir, filename))
			if err != nil {
				panic(err)
			}
			defer file.Close()

			gamelists[systemId] = file
		}

		gamelists[systemId].WriteString(path + "\n")
	}

	// move gamelist files to final destination
	gamelistFiles, err := os.ReadDir(tmpDir)
	if err != nil {
		panic(err)
	}

	for _, file := range gamelistFiles {
		src := filepath.Join(tmpDir, file.Name())
		dest := filepath.Join(*outDir, file.Name())

		if err := utils.MoveFile(src, dest); err != nil {
			panic(err)
		}
	}

	if err := os.RemoveAll(tmpDir); err != nil {
		panic(err)
	}

	if !*quiet {
		taken := int(time.Since(start).Seconds())
		if *progress {
			fmt.Println("XXX")
			fmt.Println("100")
			fmt.Printf("Indexing complete (%d games in %ds)\n", len(systemFiles), taken)
			fmt.Println("XXX")
		} else {
			fmt.Printf("Indexing complete (%d games in %ds)\n", len(systemFiles), taken)
		}
	}
}
