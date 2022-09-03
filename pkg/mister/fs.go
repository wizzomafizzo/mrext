package mister

import (
	"os"
	"path/filepath"
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
