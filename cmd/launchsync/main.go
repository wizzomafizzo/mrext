package main

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"gopkg.in/ini.v1"

	"github.com/wizzomafizzo/mrext/pkg/config"
	"github.com/wizzomafizzo/mrext/pkg/games"
	"github.com/wizzomafizzo/mrext/pkg/mister"
	"github.com/wizzomafizzo/mrext/pkg/txtindex"
)

type syncFileGame struct {
	name   string
	system *games.System
	files  []string
}

type syncFile struct {
	name    string
	author  string
	url     string
	updated time.Time
	folder  string
	games   []syncFileGame
}

func getSyncFiles(paths []string) []string {
	var syncFiles []string
	for _, path := range paths {
		files, err := filepath.Glob(path + "/*.sync")
		if err != nil {
			continue
		}
		syncFiles = append(syncFiles, files...)
	}
	return syncFiles
}

func readSyncFile(path string) (*syncFile, error) {
	var sf syncFile

	cfg, err := ini.ShadowLoad(path)
	if err != nil {
		return nil, err
	}

	sf.folder = filepath.Dir(path)

	sf.name = cfg.Section("DEFAULT").Key("name").String()
	if sf.name == "" {
		return nil, fmt.Errorf("missing name in %s", path)
	}

	sf.author = cfg.Section("DEFAULT").Key("author").String()
	if sf.author == "" {
		return nil, fmt.Errorf("missing author in %s", path)
	}

	sf.url = cfg.Section("DEFAULT").Key("url").String()

	if !cfg.Section("DEFAULT").HasKey("updated") {
		return nil, fmt.Errorf("missing updated in %s", path)
	}
	// TODO: support time
	sf.updated, err = cfg.Section("DEFAULT").Key("updated").TimeFormat("2006-01-02")
	if err != nil {
		return nil, err
	}

	for _, section := range cfg.Sections() {
		if section.Name() == "DEFAULT" {
			continue
		}

		var game syncFileGame

		game.name = section.Name()
		if game.name == "" {
			return nil, fmt.Errorf("missing name in %s -> %s", path, section.Name())
		}

		systemName := section.Key("system").String()
		system, err := games.LookupSystem(systemName)
		if err != nil {
			return nil, fmt.Errorf("invalid system in %s -> %s: %s", path, section.Name(), err)
		} else {
			game.system = system
		}

		game.files = section.Key("file").ValueWithShadows()
		if len(game.files) == 0 {
			return nil, fmt.Errorf("missing files in %s -> %s", path, section.Name())
		}

		sf.games = append(sf.games, game)
	}

	if len(sf.games) == 0 {
		return nil, fmt.Errorf("no games in %s", path)
	}

	return &sf, nil
}

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

func main() {
	menuFolders := mister.GetMenuFolders(config.SD_ROOT)
	syncFiles := getSyncFiles(menuFolders)
	var syncs []*syncFile

	for _, path := range syncFiles {
		sf, err := readSyncFile(path)
		if err != nil {
			fmt.Printf("Error reading %s: %s", path, err)
			continue
		}
		syncs = append(syncs, sf)
	}

	if len(syncs) == 0 {
		fmt.Println("No sync files found")
		os.Exit(1)
	}

	var indexSystems []*games.System
	for _, sync := range syncs {
		for _, game := range sync.games {
			indexSystems = append(indexSystems, game.system)
		}
	}

	index, err := makeIndex(indexSystems)
	if err != nil {
		fmt.Printf("Error generating index: %s\n", err)
		os.Exit(1)
	}

	for _, sync := range syncs {
		fmt.Println("---")
		fmt.Printf("Name:    %s\n", sync.name)
		fmt.Printf("Author:  %s\n", sync.author)
		fmt.Printf("URL:     %s\n", sync.url)
		fmt.Printf("Updated: %s\n", sync.updated)
		fmt.Printf("Folder:  %s\n", sync.folder)
		fmt.Printf("Games:   %d\n", len(sync.games))

		for _, game := range sync.games {
			var match txtindex.SearchResult
			fmt.Println("> " + game.name)

			for _, file := range game.files {
				results := index.SearchSystemName(game.system.Id, file)
				if len(results) > 0 {
					match = results[0]
					break
				}
			}

			if match.Name != "" {
				fmt.Println(filepath.Base(match.Path))

				// TODO: handle arcade
				mglContent, err := mister.GenerateMgl(game.system, match.Path)
				if err != nil {
					fmt.Println(err)
					continue
				}

				mglPath := filepath.Join(sync.folder, game.name+".mgl")
				fp, err := os.Create(mglPath)
				if err != nil {
					fmt.Println(err)
					continue
				}
				defer fp.Close()
				fp.WriteString(mglContent)
			} else {
				// TODO: generate missing mgl placeholder
				fmt.Println("No match found")
			}
		}
	}
}
