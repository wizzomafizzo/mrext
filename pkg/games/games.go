package games

import (
	"fmt"
	"io/fs"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	s "strings"

	"github.com/wizzomafizzo/mrext/pkg/utils"
)

type dupeChecker struct {
	filenames map[string]bool
}

func (d *dupeChecker) isDupe(path string) bool {
	fn := filepath.Base(path)
	_, exists := d.filenames[fn]

	if exists {
		return true
	} else {
		d.filenames[fn] = true
		return false
	}
}

func getSystem(name string) (*System, error) {
	if system, ok := SYSTEMS[name]; ok {
		return &system, nil
	} else {
		return nil, fmt.Errorf("unknown system: %s", name)
	}
}

func matchSystemFolder(path string) ([][2]string, error) {
	var matches [][2]string

	folder, err := os.Stat(path)
	if err != nil {
		return nil, err
	}

	name := folder.Name()

	if !folder.IsDir() {
		return nil, fmt.Errorf("not a directory: %s", path)
	}

	for k, v := range SYSTEMS {
		if s.EqualFold(name, v.folder) {
			matches = append(matches, [2]string{k, path})
		}
	}

	if len(matches) == 0 {
		return nil, fmt.Errorf("unknown system: %s", name)
	} else {
		return matches, nil
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
		return nil
	}

	folders, err := ioutil.ReadDir(path)
	if err != nil {
		return nil
	}

	for _, folder := range folders {
		abs := filepath.Join(path, folder.Name())

		if folder.IsDir() && s.ToLower(folder.Name()) == "games" {
			found = append(found, findSystemFolders(abs)...)
		}

		matches, err := matchSystemFolder(abs)
		if err != nil {
			continue
		} else {
			found = append(found, matches...)
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

func GetSystemFiles(systemPaths map[string][]string, statusFn func(systemId string, path string)) ([][2]string, error) {
	var dupes = &dupeChecker{filenames: make(map[string]bool)}
	var allFound [][2]string
	var folderFound [][2]string

	cwd, err := os.Getwd()
	if err != nil {
		return nil, err
	}

	for systemId, paths := range systemPaths {
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
					if matchSystemFile(*system, zipPath) && !dupes.isDupe(zipPath) {
						abs := filepath.Join(path, zipPath)
						folderFound = append(folderFound, [2]string{systemId, string(abs)})

					}
				}
			} else {
				if matchSystemFile(*system, path) && !dupes.isDupe(path) {
					folderFound = append(folderFound, [2]string{systemId, path})
				}
			}
			return nil
		}

		for _, path := range paths {
			statusFn(systemId, path)

			folderFound = nil

			err = os.Chdir(cwd)
			if err != nil {
				return nil, err
			}

			folder, err := os.Lstat(path)
			if err != nil {
				continue
			}

			if folder.Mode()&os.ModeSymlink == 0 {
				filepath.WalkDir(path, scanner)
				allFound = append(allFound, folderFound...)
			} else {
				realPath, err := filepath.EvalSymlinks(path)
				if err != nil {
					continue
				}
				filepath.WalkDir(realPath, scanner)
				for _, result := range folderFound {
					result[1] = s.Replace(result[1], realPath, path, 1)
					allFound = append(allFound, result)
				}
			}
		}
	}

	err = os.Chdir(cwd)
	if err != nil {
		return nil, err
	}

	return allFound, nil
}
