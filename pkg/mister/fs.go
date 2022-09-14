package mister

import (
	"os"
	"path/filepath"
	"strings"
)

// Search for directories in root that start with "_".
func GetMenuFolders(root string) []string {
	var folders []string

	// TODO: confirm menu can't traverse symlinks
	var scan func(path string)
	scan = func(folder string) {
		files, err := os.ReadDir(folder)
		if err != nil {
			return
		}
		for _, file := range files {
			if file.IsDir() && file.Name()[0] == '_' {
				path := filepath.Join(folder, file.Name())
				folders = append(folders, path)
				scan(path)
			}
		}
	}

	scan(root)
	return folders
}

func isRbf(path string) bool {
	return filepath.Ext(strings.ToLower(path)) == ".rbf"
}

// Do a shallow search of RBF files in root and return list of relative paths.
func GetRbfs(root string) []string {
	var rbfs []string

	rootFiles, err := os.ReadDir(root)
	if err != nil {
		return nil
	}

	for _, rootFile := range rootFiles {
		if !rootFile.IsDir() && isRbf(rootFile.Name()) {
			rbfs = append(rbfs, rootFile.Name())
		} else if rootFile.IsDir() && rootFile.Name()[0] == '_' {
			subFiles, err := os.ReadDir(filepath.Join(root, rootFile.Name()))
			if err != nil {
				continue
			}

			for _, subFile := range subFiles {
				if !subFile.IsDir() && isRbf(subFile.Name()) {
					rbfs = append(rbfs, rootFile.Name()+"/"+subFile.Name())
				}
			}
		}
	}

	return rbfs
}

// Find an RBF in a list of all RBFs and return a value suitable for MGL.
func MatchRbf(rbfs []string, match string) string {
	if len(rbfs) == 0 {
		return ""
	}

	for _, rbf := range rbfs {
		parts := strings.Split(rbf, "/")
		file := parts[len(parts)-1]
		if strings.HasPrefix(strings.ToLower(file), strings.ToLower(match)) {
			if len(parts) == 1 {
				return match
			} else {
				return strings.Join(append(parts[0:len(parts)-1], match), "/")
			}
		}
	}

	return ""
}
