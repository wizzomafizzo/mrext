package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/wizzomafizzo/mrext/pkg/config"
	"github.com/wizzomafizzo/mrext/pkg/games"
	"github.com/wizzomafizzo/mrext/pkg/mister"
	"github.com/wizzomafizzo/mrext/pkg/utils"
)

func main() {
	systemId := flag.String("system", "auto", "system for game (optional)")
	favName := flag.String("name", "", "name of entry in menu")
	gameFile := flag.String("game", "", "path to game file")
	menuFolder := flag.String("folder", "", "path to menu folder")
	flag.Parse()

	if *favName == "" {
		fmt.Printf("Favorite name is required.\n")
		os.Exit(1)
	} else {
		*favName = utils.StripBadFileChars(*favName)
	}

	if !games.FileExists(*gameFile) {
		fmt.Printf("Game file does not exist: %s\n", *gameFile)
		os.Exit(1)
	}

	if !filepath.IsAbs(*menuFolder) {
		*menuFolder = filepath.Join(config.SdFolder, *menuFolder)
	}

	if _, err := os.Stat(*menuFolder); os.IsNotExist(err) {
		fmt.Printf("Menu folder does not exist: %s\n", *menuFolder)
		os.Exit(1)
	}

	var system games.System

	if *systemId == "auto" {
		match, err := games.PathBestMatch(*gameFile)
		if err != nil {
			fmt.Printf("Could not determine system for game: %s\n", *gameFile)
			os.Exit(1)
		}
		system = match
	} else {
		lookup, err := games.LookupSystem(*systemId)
		if err != nil {
			fmt.Printf("Invalid system ID: %s\n", *systemId)
			os.Exit(1)
		}

		system = *lookup
	}

	path, err := mister.CreateLauncher(&system, *gameFile, *menuFolder, *favName)
	if err != nil {
		fmt.Printf("Error creating favorite: %s\n", err)
		os.Exit(1)
	}

	fmt.Printf("Favorite created: %s\n", path)
}
