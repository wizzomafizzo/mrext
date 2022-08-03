package games

import (
	"fmt"
	"io/fs"
	"io/ioutil"
	"os"
	"path/filepath"
	s "strings"

	"github.com/wizzomafizzo/mrext/pkg/utils"
)

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

type resultsStack [][]string

func (r *resultsStack) new() {
	*r = append(*r, []string{})
}

func (r *resultsStack) pop() {
	if len(*r) == 0 {
		return
	}
	*r = (*r)[:len(*r)-1]
}

func (r *resultsStack) get() (*[]string, error) {
	if len(*r) == 0 {
		return nil, fmt.Errorf("nothing on stack")
	}
	return &(*r)[len(*r)-1], nil
}

func GetFiles(systemId string, path string) ([]string, error) {
	var allResults []string
	var stack resultsStack
	visited := make(map[string]struct{})

	system, err := getSystem(systemId)
	if err != nil {
		return nil, err
	}

	cwd, err := os.Getwd()
	if err != nil {
		return nil, err
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
			err = os.Chdir(filepath.Dir(path))
			if err != nil {
				return err
			}

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
					result = s.Replace(result, realPath, path, 1)
					allResults = append(allResults, result)
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
				if matchSystemFile(*system, zipPath) {
					abs := filepath.Join(path, zipPath)
					*results = append(*results, string(abs))

				}
			}
		} else {
			// regular files
			if matchSystemFile(*system, path) {
				*results = append(*results, path)
			}
		}

		return nil
	}

	err = os.Chdir(cwd)
	if err != nil {
		return nil, err
	}

	stack.new()
	filepath.WalkDir(path, scanner)

	results, err := stack.get()
	if err != nil {
		return nil, err
	}

	allResults = append(allResults, *results...)
	stack.pop()

	return allResults, nil
}

// Search for all valid games in given paths and return a single list of files
// with their corresponding system names. An optional function can be given
// which is simply triggered before each system path is searched (for use in
// progress displays).
// This function supports deep searching in .zip files and symlinked directories.
func GetAllFiles(systemPaths map[string][]string, statusFn func(systemId string, path string)) ([][2]string, error) {
	var allFiles [][2]string

	cwd, err := os.Getwd()
	if err != nil {
		return nil, err
	}

	for systemId, paths := range systemPaths {
		for _, path := range paths {
			statusFn(systemId, path)

			err = os.Chdir(filepath.Dir(path))
			if err != nil {
				return nil, err
			}

			folder, err := os.Lstat(path)
			if err != nil {
				continue
			}

			if folder.Mode()&os.ModeSymlink == 0 {
				// regular folders
				results, err := GetFiles(systemId, path)
				if err != nil {
					return nil, err
				}

				for _, filePath := range results {
					allFiles = append(allFiles, [2]string{systemId, filePath})
				}
			} else {
				// handle symlinked games folders
				realPath, err := filepath.EvalSymlinks(path)
				if err != nil {
					continue
				}

				results, err := GetFiles(systemId, path)
				if err != nil {
					return nil, err
				}

				for _, filePath := range results {
					allFiles = append(allFiles, [2]string{
						systemId,
						s.Replace(filePath, realPath, path, 1),
					})
				}
			}
		}
	}

	err = os.Chdir(cwd)
	if err != nil {
		return nil, err
	}

	return allFiles, nil
}
