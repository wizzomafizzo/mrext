package games

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/wizzomafizzo/mrext/pkg/utils"
)

// Lookup up exact system id definition.
func GetSystem(id string) (*System, error) {
	if system, ok := Systems[id]; ok {
		return &system, nil
	} else {
		return nil, fmt.Errorf("unknown system: %s", id)
	}
}

func GetGroup(groupId string) (System, error) {
	var merged System
	if _, ok := CoreGroups[groupId]; !ok {
		return merged, fmt.Errorf("no system group found for %s", groupId)
	}

	if len(CoreGroups[groupId]) < 1 {
		return merged, fmt.Errorf("no systems in %s", groupId)
	} else if len(CoreGroups[groupId]) == 1 {
		return CoreGroups[groupId][0], nil
	}

	merged = CoreGroups[groupId][0]
	merged.FileTypes = make([]FileType, 0)
	for _, s := range CoreGroups[groupId] {
		merged.FileTypes = append(merged.FileTypes, s.FileTypes...)
	}

	return merged, nil
}

// Lookup case insensitive system id definition including aliases.
func LookupSystem(id string) (*System, error) {
	if system, err := GetGroup(id); err == nil {
		return &system, nil
	}

	for k, v := range Systems {
		if strings.EqualFold(k, id) {
			return &v, nil
		}

		for _, alias := range v.Alias {
			if strings.EqualFold(alias, id) {
				return &v, nil
			}
		}
	}

	return nil, fmt.Errorf("unknown system: %s", id)
}

// Return true if a given files extension is valid for a system.
func MatchSystemFile(system System, path string) bool {
	for _, args := range system.FileTypes {
		for _, ext := range args.Exts {
			if strings.HasSuffix(strings.ToLower(path), ext) {
				return true
			}
		}
	}
	return false
}

// Return a slice of all systems.
func AllSystems() []System {
	var systems []System

	for _, system := range Systems {
		systems = append(systems, system)
	}

	return systems
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

// Search for all valid games in a given path and return a list of files.
// This function deep searches .zip files and handles symlinks at all levels.
func GetFiles(systemId string, path string) ([]string, error) {
	var allResults []string
	var stack resultsStack
	visited := make(map[string]struct{})

	system, err := GetSystem(systemId)
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

				for i := range *results {
					allResults = append(allResults, strings.Replace((*results)[i], realPath, path, 1))
				}

				stack.pop()
				return nil
			}
		}

		results, err := stack.get()
		if err != nil {
			return err
		}

		if strings.HasSuffix(strings.ToLower(path), ".zip") {
			// zip files
			zipFiles, err := utils.ListZip(path)
			if err != nil {
				return err
			}

			for i := range zipFiles {
				if MatchSystemFile(*system, zipFiles[i]) {
					abs := filepath.Join(path, zipFiles[i])
					*results = append(*results, string(abs))

				}
			}
		} else {
			// regular files
			if MatchSystemFile(*system, path) {
				*results = append(*results, path)
			}
		}

		return nil
	}

	stack.new()

	root, err := os.Lstat(path)
	if err != nil {
		return nil, err
	}

	err = os.Chdir(filepath.Dir(path))
	if err != nil {
		return nil, err
	}

	// handle symlinks on root game folder because WalkDir fails silently on them
	var realPath string
	if root.Mode()&os.ModeSymlink == 0 {
		realPath = path
	} else {
		realPath, err = filepath.EvalSymlinks(path)
		if err != nil {
			return nil, err
		}
	}

	realRoot, err := os.Stat(realPath)
	if err != nil {
		return nil, err
	}

	if !realRoot.IsDir() {
		return nil, fmt.Errorf("root is not a directory")
	}

	filepath.WalkDir(realPath, scanner)

	results, err := stack.get()
	if err != nil {
		return nil, err
	}

	allResults = append(allResults, *results...)
	stack.pop()

	// change root back to symlink
	if realPath != path {
		for i := range allResults {
			allResults[i] = strings.Replace(allResults[i], realPath, path, 1)
		}
	}

	err = os.Chdir(cwd)
	if err != nil {
		return nil, err
	}

	return allResults, nil
}

func GetAllFiles(systemPaths map[string][]string, statusFn func(systemId string, path string)) ([][2]string, error) {
	var allFiles [][2]string

	for systemId, paths := range systemPaths {
		for i := range paths {
			statusFn(systemId, paths[i])

			files, err := GetFiles(systemId, paths[i])
			if err != nil {
				return nil, err
			}

			for i := range files {
				allFiles = append(allFiles, [2]string{systemId, files[i]})
			}
		}
	}

	return allFiles, nil
}

func FilterUniqueFilenames(files []string) []string {
	var filtered []string
	filenames := make(map[string]struct{})
	for i := range files {
		fn := filepath.Base(files[i])
		if _, ok := filenames[fn]; ok {
			continue
		} else {
			filenames[fn] = struct{}{}
			filtered = append(filtered, files[i])
		}
	}
	return filtered
}

var zipRe = regexp.MustCompile(`^(.*\.zip)/(.+)$`)

func FileExists(path string) bool {
	_, err := os.Stat(path)
	if err == nil {
		return true
	}

	zipMatch := zipRe.FindStringSubmatch(path)
	if zipMatch != nil {
		zipPath := zipMatch[1]
		file := zipMatch[2]

		zipFiles, err := utils.ListZip(zipPath)
		if err != nil {
			return false
		}

		for i := range zipFiles {
			if zipFiles[i] == file {
				return true
			}
		}
	}

	return false
}
