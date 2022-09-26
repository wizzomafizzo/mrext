package games

import (
	"os"

	"github.com/wizzomafizzo/mrext/pkg/utils"
)

type fileChecker struct {
	zipCache map[string]map[string]struct{}
}

func (fc *fileChecker) cacheZip(zipPath string, files []string) {
	fc.zipCache[zipPath] = make(map[string]struct{})
	for _, file := range files {
		fc.zipCache[zipPath][file] = struct{}{}
	}
}

func (fc *fileChecker) existsZip(zipPath string, file string) bool {
	if _, ok := fc.zipCache[zipPath]; !ok {
		files, err := utils.ListZip(zipPath)
		if err != nil {
			return false
		}

		fc.cacheZip(zipPath, files)
	}

	if _, ok := fc.zipCache[zipPath][file]; !ok {
		return false
	}

	return true
}

func (fc *fileChecker) Exists(path string) bool {
	_, err := os.Stat(path)
	if err == nil {
		return true
	}

	zipMatch := zipRe.FindStringSubmatch(path)
	if zipMatch != nil {
		zipPath := zipMatch[1]
		file := zipMatch[2]

		return fc.existsZip(zipPath, file)
	}

	return false
}

func NewFileChecker() *fileChecker {
	return &fileChecker{
		zipCache: make(map[string]map[string]struct{}),
	}
}
