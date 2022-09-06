package main

import (
	"fmt"
	"os"

	"github.com/wizzomafizzo/mrext/pkg/config"
	"github.com/wizzomafizzo/mrext/pkg/mister"
)

func main() {
	fmt.Print("Searching for sync files... ")
	menuFolders := mister.GetMenuFolders(config.SD_ROOT)
	syncFiles := getSyncFiles(menuFolders)
	var syncs []*syncFile

	for _, path := range syncFiles {
		sf, err := readSyncFile(path)
		if err != nil {
			fmt.Printf("Error reading %s: %s\n", path, err)
			continue
		}
		syncs = append(syncs, sf)
	}

	if len(syncs) == 0 {
		fmt.Println("no sync files found")
		os.Exit(1)
	}
	fmt.Printf("found %d files\n", len(syncs))

	fmt.Println("Checking for updates...")
	for i, sync := range syncs {
		fmt.Printf("%d/%d: %s... ", i+1, len(syncs), sync.name)
		newSync, updated, err := checkForUpdate(sync)
		if err != nil {
			fmt.Printf("error: %s\n", err)
		} else if updated {
			syncs[i] = newSync
			fmt.Println("updated")
		} else {
			fmt.Println("no update")
		}
	}

	fmt.Print("Building games index... ")
	index, err := makeIndex(syncs)
	if err != nil {
		fmt.Printf("error generating index: %s\n", err)
		os.Exit(1)
	}
	fmt.Println("done")

	for _, sync := range syncs {
		fmt.Println("---")
		fmt.Printf("Name:    %s\n", sync.name)
		fmt.Printf("Author:  %s\n", sync.author)
		fmt.Printf("URL:     %s\n", sync.url)
		fmt.Printf("Updated: %s\n", sync.updated)
		fmt.Printf("Folder:  %s\n", sync.folder)
		fmt.Println("Games:")

		for _, game := range sync.games {
			fmt.Print("- " + game.name + "... ")
			file, found, err := tryLinkGame(sync, game, index)
			if err != nil {
				fmt.Printf("error: %s\n", err)
			} else if found {
				fmt.Printf("found %s\n", file)
			} else {
				fmt.Println("not found")
			}
		}
	}
}
