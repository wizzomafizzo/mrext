package main

import (
	"flag"
	"fmt"
	"github.com/wizzomafizzo/mrext/pkg/config"
	"github.com/wizzomafizzo/mrext/pkg/games"
	"github.com/wizzomafizzo/mrext/pkg/mister"
	"math/rand"
	"os"
	"path/filepath"
	"time"
)

func main() {
	system := flag.String("system", "n64", "system to load games from")
	delay := flag.Int("delay", 40, "number of seconds between loading each game")
	random := flag.Bool("random", false, "randomize the order of games")
	offset := flag.Int("offset", 0, "offset of games list to start at (not used for random)")
	path := flag.String("path", "", "custom additional path to scan for games")
	flag.Parse()

	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	cfg := config.UserConfig{}

	if *path != "" {
		cfg.Systems.GamesFolder = []string{*path}
	}

	sys, err := games.LookupSystem(*system)
	if err != nil {
		fmt.Printf("Error loading system %s: %s\n", *system, err)
		os.Exit(1)
	}

	fmt.Println("Scanning games...")

	folders := games.GetSystemPaths(&cfg, []games.System{*sys})
	var files []string

	for _, folder := range folders {
		result, err := games.GetFiles(sys.Id, folder.Path)
		if err != nil {
			fmt.Printf("Error scanning folder %s: %s\n", folder.Path, err)
			continue
		}
		files = append(files, result...)
	}

	if len(files) == 0 {
		fmt.Println("No games found")
		os.Exit(1)
	} else {
		fmt.Printf("Found %d games\n", len(files))
	}

	index := *offset
	if index >= len(files) {
		index = 0
	}

	fmt.Printf("Runnning with a %ds delay. Ctrl-c to exit\n", *delay)

	for {
		if index >= len(files) {
			index = 0
		}

		if *random {
			index = r.Intn(len(files))
		}

		name := filepath.Base(files[index])
		name = name[:len(name)-len(filepath.Ext(name))]

		fmt.Printf(
			"%s - %d: %s <%s>\n",
			time.Now().Format("15:04:05"),
			index,
			name,
			files[index],
		)

		err := mister.LaunchGenericFile(&cfg, files[index])
		if err != nil {
			fmt.Printf("Error launching game: %s\n", err)
			index++
			continue
		}

		index++
		time.Sleep(time.Duration(*delay) * time.Second)
	}
}
