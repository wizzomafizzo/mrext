package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/wizzomafizzo/mrext/pkg/config"
	"github.com/wizzomafizzo/mrext/pkg/games"
	"github.com/wizzomafizzo/mrext/pkg/mister"
	"github.com/wizzomafizzo/mrext/pkg/txtindex"
	"github.com/wizzomafizzo/mrext/pkg/utils"
)

func makeIndex(systems []*games.System) (txtindex.Index, error) {
	var index txtindex.Index
	indexFile := filepath.Join(os.TempDir(), "launchsync-index.tar")

	systemPaths := make(map[string][]string)
	for systemId, path := range games.GetSystemPaths() {
		for _, system := range systems {
			if system.Id == systemId {
				systemPaths[systemId] = path
				break
			}
		}
	}

	systemFiles := make([][2]string, 0)
	for systemId, paths := range systemPaths {
		for _, path := range paths {
			files, err := games.GetFiles(systemId, path)
			if err != nil {
				return index, err
			}

			for _, file := range files {
				systemFiles = append(systemFiles, [2]string{systemId, file})
			}
		}
	}

	err := txtindex.Generate(systemFiles, indexFile)
	if err != nil {
		return index, err
	}

	index, err = txtindex.Open(indexFile)
	if err != nil {
		return index, err
	}
	os.Remove(indexFile)

	return index, nil
}

func notFoundFilename(folder string, name string) string {
	return filepath.Join(folder, name+" [NOT FOUND].mgl")
}

func main() {
	fmt.Println("Searching for sync files...")
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
		fmt.Println("No sync files found")
		os.Exit(1)
	}
	fmt.Printf("Found %d sync files\n", len(syncs))

	fmt.Println("Checking for updates...")
	for i, sync := range syncs {
		fmt.Printf("%d/%d: %s\n", i+1, len(syncs), sync.name)

		newSync, updated, err := updateSyncFile(sync)
		if err != nil {
			fmt.Printf("Error updating %s: %s\n", sync.name, err)
			continue
		}

		var newNames []string
		for _, game := range newSync.games {
			newNames = append(newNames, game.name)
		}

		for _, game := range sync.games {
			if !utils.Contains(newNames, game.name) {
				mister.DeleteLauncher(mister.GetLauncherFilename(game.system, sync.folder, game.name))
				os.Remove(notFoundFilename(sync.folder, game.name))
			}
		}

		if updated {
			syncs[i] = newSync
		}
	}

	// Restrict index to necessary systems
	var indexSystems []*games.System
	for _, sync := range syncs {
		for _, game := range sync.games {
			indexSystems = append(indexSystems, game.system)
		}
	}

	fmt.Println("Building index...")
	index, err := makeIndex(indexSystems)
	if err != nil {
		fmt.Printf("Error generating index: %s\n", err)
		os.Exit(1)
	}
	fmt.Println("Index built")

	for _, sync := range syncs {
		fmt.Println("---")
		fmt.Printf("Name:    %s\n", sync.name)
		fmt.Printf("Author:  %s\n", sync.author)
		fmt.Printf("URL:     %s\n", sync.url)
		fmt.Printf("Updated: %s\n", sync.updated)
		fmt.Printf("Folder:  %s\n", sync.folder)
		fmt.Println("Games:")

		for _, game := range sync.games {
			var match txtindex.SearchResult
			fmt.Println("- " + game.name)

			for _, re := range game.matches {
				results := index.SearchSystemNameRe(game.system.Id, re)
				if len(results) > 0 {
					match = results[0]
					break
				}
			}

			if match.Name != "" {
				fmt.Println(filepath.Base(match.Path))
				// TODO: don't write if it's the same file
				_, err := mister.CreateLauncher(game.system, match.Path, sync.folder, game.name)
				if err != nil {
					fmt.Printf("Error creating launcher: %s\n", err)
				} else {
					if _, err := os.Stat(notFoundFilename(sync.folder, game.name)); err == nil {
						os.Remove(notFoundFilename(sync.folder, game.name))
					}
				}
			} else {
				fp, err := os.Create(notFoundFilename(sync.folder, game.name))
				if err != nil {
					fmt.Printf("Error creating not found placeholder: %s\n", err)
				}
				fp.Close()
			}
		}
	}
}
