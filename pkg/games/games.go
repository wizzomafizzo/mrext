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
	filenames map[string]struct{}
	enabled   bool
}

func (d *dupeChecker) isDupe(path string) bool {
	if !d.enabled {
		return false
	}

	fn := filepath.Base(path)
	_, exists := d.filenames[fn]

	if exists {
		return true
	} else {
		d.filenames[fn] = struct{}{}
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

type resultsStack [][][2]string

func (r *resultsStack) new() {
	*r = append(*r, [][2]string{})
}

func (r *resultsStack) pop() {
	if len(*r) == 0 {
		return
	}
	*r = (*r)[:len(*r)-1]
}

func (r *resultsStack) get() (*[][2]string, error) {
	if len(*r) == 0 {
		return nil, fmt.Errorf("nothing on stack")
	}
	return &(*r)[len(*r)-1], nil
}

// Search for all valid games in given paths and return a single list of files
// with their corresponding system names. An optional function can be given
// which is simply triggered before each system is searched (for use in progress
// displays).
// This function supports deep searching in .zip files and symlinked directories.
func GetSystemFiles(systemPaths map[string][]string, statusFn func(systemId string, path string), removeDupes bool) ([][2]string, error) {
	dupes := &dupeChecker{filenames: make(map[string]struct{}), enabled: removeDupes}
	visited := make(map[string]struct{})
	var allFiles [][2]string
	var stack resultsStack

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

		var scanner func(path string, file fs.DirEntry, err error) error
		scanner = func(path string, file fs.DirEntry, _ error) error {
			// avoid recursive symlinks
			if file.IsDir() {
				if _, ok := visited[path]; ok {
					return filepath.SkipDir
				} else {
					visited[path] = struct{}{}
				}
			}

			// handle symlinked directories
			if file.Type()&os.ModeSymlink != 0 {
				realPath, err := filepath.EvalSymlinks(path)
				if err != nil {
					return err
				}

				file, err := os.Stat(realPath)
				if err != nil {
					return err
				}

				if file.IsDir() {
					err = os.Chdir(path)
					if err != nil {
						return err
					}

					stack.new()

					filepath.WalkDir(realPath, scanner)

					results, err := stack.get()
					if err != nil {
						return err
					}

					for _, result := range *results {
						result[1] = s.Replace(result[1], realPath, path, 1)
						allFiles = append(allFiles, result)
					}

					stack.pop()
					return nil
				}
			}

			results, err := stack.get()
			if err != nil {
				return err
			}

			if s.HasSuffix(s.ToLower(path), ".zip") {
				// zip files
				zipFiles, err := utils.ListZip(path)
				if err != nil {
					return err
				}

				for _, zipPath := range zipFiles {
					if matchSystemFile(*system, zipPath) && !dupes.isDupe(zipPath) {
						abs := filepath.Join(path, zipPath)
						*results = append(*results, [2]string{systemId, string(abs)})

					}
				}
			} else {
				// regular files
				if matchSystemFile(*system, path) && !dupes.isDupe(path) {
					*results = append(*results, [2]string{systemId, path})
				}
			}

			return nil
		}

		for _, path := range paths {
			statusFn(systemId, path)

			stack.new()
			visited = make(map[string]struct{})

			err = os.Chdir(cwd)
			if err != nil {
				return nil, err
			}

			folder, err := os.Lstat(path)
			if err != nil {
				continue
			}

			if folder.Mode()&os.ModeSymlink == 0 {
				// handle symlinked games folders
				filepath.WalkDir(path, scanner)
				results, err := stack.get()
				if err != nil {
					return nil, err
				}
				allFiles = append(allFiles, *results...)
			} else {
				realPath, err := filepath.EvalSymlinks(path)
				if err != nil {
					continue
				}

				filepath.WalkDir(realPath, scanner)

				results, err := stack.get()
				if err != nil {
					return nil, err
				}

				for _, result := range *results {
					result[1] = s.Replace(result[1], realPath, path, 1)
					allFiles = append(allFiles, result)
				}
			}

			stack.pop()
		}
	}

	err = os.Chdir(cwd)
	if err != nil {
		return nil, err
	}

	return allFiles, nil
}
