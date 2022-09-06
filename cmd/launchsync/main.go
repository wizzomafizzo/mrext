package main

import (
	"flag"
	"fmt"
	"net/http"
	"os"

	"github.com/wizzomafizzo/mrext/pkg/config"
	"github.com/wizzomafizzo/mrext/pkg/mister"
)

func testSyncFile(path string) {
	sf, err := readSyncFile(path)
	if err != nil {
		fmt.Printf("Error reading %s: %s\n", path, err)
		os.Exit(1)
	}

	fmt.Printf("Name:    %s\n", sf.name)
	fmt.Printf("Author:  %s\n", sf.author)
	fmt.Printf("URL:     %s\n", sf.url)
	fmt.Printf("Updated: %s\n", sf.updated)
	fmt.Printf("Folder:  %s\n", sf.folder)
	fmt.Printf("Games:   %d\n", len(sf.games))
	fmt.Println("---")

	if sf.url != "" {
		fmt.Printf("Testing URL... ")

		resp, err := http.Get(sf.url)
		if err != nil {
			fmt.Printf("error: %s\n", err)
		} else if resp.StatusCode != 200 {
			fmt.Printf("bad response: %s\n", resp.Status)
		} else {
			fmt.Println("tested OK")
		}
	}

	if len(sf.games) == 0 {
		fmt.Println("---")
		fmt.Println("No games")
		return
	}

	fmt.Print("Building games index... ")
	index, err := makeIndex([]*syncFile{sf})
	if err != nil {
		fmt.Printf("error generating index: %s\n", err)
		os.Exit(1)
	}
	fmt.Println("done")

	for _, game := range sf.games {
		fmt.Println("---")
		fmt.Printf("Game:    %s\n", game.name)
		fmt.Printf("System:  %s\n", game.system.Id)
		fmt.Printf("Matches: %d\n", len(game.matches))

		for _, match := range game.matches {
			fmt.Printf("- %s\n", match[4:])
			results := index.SearchSystemByNameRe(game.system.Id, match)
			for i := 0; i < 5 && i < len(results); i++ {
				if i == 0 {
					fmt.Printf(" *%s\n", results[i].Path)
				} else {
					fmt.Printf("  %s\n", results[i].Path)
				}
			}
		}
	}
}

func main() {
	update := flag.Bool("update", false, "find, update and link all sync files on system")
	verbose := flag.Bool("verbose", false, "print status information during update")
	test := flag.String("test", "", "report if specified sync file is valid and display match results")
	flag.Parse()

	if *test != "" {
		testSyncFile(*test)
		return
	}

	if *verbose || !*update {
		fmt.Print("Searching for sync files... ")
	}
	menuFolders := mister.GetMenuFolders(config.SD_ROOT)
	syncFiles := getSyncFiles(menuFolders)
	var syncs []*syncFile

	for _, path := range syncFiles {
		sf, err := readSyncFile(path)
		if err != nil {
			if *verbose || !*update {
				fmt.Printf("Error reading %s: %s\n", path, err)
			}
			continue
		}
		syncs = append(syncs, sf)
	}

	if len(syncs) == 0 {
		if *verbose || !*update {
			fmt.Println("no sync files found")
		}
		os.Exit(1)
	}
	if *verbose || !*update {
		fmt.Printf("found %d\n", len(syncs))
	}

	if *verbose || !*update {
		fmt.Println("Checking for updates...")
	}
	for i, sync := range syncs {
		if *verbose || !*update {
			fmt.Printf("%d/%d: %s... ", i+1, len(syncs), sync.name)
		}
		newSync, updated, err := checkForUpdate(sync)
		if err != nil {
			if *verbose || !*update {
				fmt.Printf("error: %s\n", err)
			}
		} else if updated {
			syncs[i] = newSync
			if *verbose || !*update {
				fmt.Println("updated")
			}
		} else {
			if *verbose || !*update {
				fmt.Println("no update")
			}
		}
	}

	if *verbose || !*update {
		fmt.Print("Building games index... ")
	}
	index, err := makeIndex(syncs)
	if err != nil {
		if *verbose || !*update {
			fmt.Printf("error generating index: %s\n", err)
		}
		os.Exit(1)
	}
	if *verbose || !*update {
		fmt.Println("done")
	}

	for _, sync := range syncs {
		if *verbose || !*update {
			fmt.Println("---")
			fmt.Printf("Name:    %s\n", sync.name)
			fmt.Printf("Author:  %s\n", sync.author)
			fmt.Printf("URL:     %s\n", sync.url)
			fmt.Printf("Updated: %s\n", sync.updated)
			fmt.Printf("Folder:  %s\n", sync.folder)
			fmt.Println("Games:")
		}

		for _, game := range sync.games {
			if *verbose || !*update {
				fmt.Print("- " + game.name + "... ")
			}
			file, found, err := tryLinkGame(sync, game, index)
			if *verbose || !*update {
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
}
