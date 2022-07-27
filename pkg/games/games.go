package games

import (
	"fmt"
	"io/fs"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	s "strings"

	"github.com/wizzomafizzo/mext/pkg/utils"
)

func getSystem(name string) (*System, error) {
	if system, ok := SYSTEMS[name]; ok {
		return &system, nil
	} else {
		return nil, fmt.Errorf("unknown system: %s", name)
	}
}

func matchSystemFolder(folder fs.FileInfo) ([]string, error) {
	var found []string
	for k, v := range SYSTEMS {
		if folder.IsDir() && s.EqualFold(folder.Name(), v.folder) {
			found = append(found, k)
		}
	}
	if len(found) == 0 {
		return nil, fmt.Errorf("unknown system: %s", folder)
	} else {
		return found, nil
	}
}

func matchSystemFile(system System, path string) bool {
	for _, args := range system.fileTypes {
		for _, ext := range args.extensions {
			if s.HasSuffix(s.ToLower(path), ext) {
				return true
			}
		}
	}
	return false
}

func findSystemFolders(path string) [][2]string {
	var found [][2]string

	root, err := os.Stat(path)
	if err != nil || !root.IsDir() {
		return found
	}

	folders, err := ioutil.ReadDir(path)
	if err != nil {
		return found
	}

	for _, folder := range folders {
		abs := filepath.Join(path, folder.Name())

		if folder.IsDir() && s.ToLower(folder.Name()) == "games" {
			found = append(found, findSystemFolders(abs)...)
		}

		matches, err := matchSystemFolder(folder)
		if err != nil {
			continue
		} else {
			for _, match := range matches {
				found = append(found, [2]string{match, abs})
			}
		}
	}

	return found
}

func GetSystemPaths() map[string][]string {
	var paths = make(map[string][]string)

	for _, rootPath := range GAMES_FOLDERS {
		for _, result := range findSystemFolders(rootPath) {
			paths[result[0]] = append(paths[result[0]], result[1])
		}
	}

	return paths
}

func GetSystemFiles(statusFn func(system string)) [][2]string {
	var found [][2]string

	for systemId, paths := range GetSystemPaths() {
		statusFn(systemId)

		system, err := getSystem(systemId)
		if err != nil {
			log.Println(err)
			continue
		}

		scanner := func(path string, _ fs.DirEntry, _ error) error {
			if s.HasSuffix(s.ToLower(path), ".zip") {
				zipFiles, err := utils.ListZip(path)
				if err != nil {
					return err
				}

				for _, zipPath := range zipFiles {
					if matchSystemFile(*system, zipPath) {
						abs := filepath.Join(path, zipPath)
						found = append(found, [2]string{systemId, string(abs)})

					}
				}
			} else {
				if matchSystemFile(*system, path) {
					found = append(found, [2]string{systemId, path})
				}
			}
			return nil
		}

		for _, path := range paths {
			filepath.WalkDir(path, scanner)
		}
	}

	return found
}
