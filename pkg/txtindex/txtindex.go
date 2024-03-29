package txtindex

import (
	"archive/tar"
	"bufio"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/wizzomafizzo/mrext/pkg/config"
	"github.com/wizzomafizzo/mrext/pkg/utils"
)

func Exists() bool {
	_, err := os.Stat(config.SearchDbFile)
	return err == nil
}

func Generate(files [][2]string, indexPath string) error {
	tmpDir, err := os.MkdirTemp(os.TempDir(), "search-")
	if err != nil {
		return err
	}

	tmpFilesDir := filepath.Join(tmpDir, "files")
	if err := os.Mkdir(tmpFilesDir, 0755); err != nil {
		return err
	}

	indexFiles := make(map[string]*os.File)
	getIndexFile := func(fn string) (*os.File, error) {
		if _, ok := indexFiles[fn]; !ok {
			indexFiles[fn], err = os.Create(filepath.Join(tmpFilesDir, fn))
			if err != nil {
				return nil, err
			}
		}

		return indexFiles[fn], nil
	}

	for i := range files {
		pathsFile, err := getIndexFile(files[i][0] + "__paths")
		if err != nil {
			return err
		}

		namesFile, err := getIndexFile(files[i][0] + "__names")
		if err != nil {
			return err
		}

		basename := filepath.Base(files[i][1])
		name := strings.TrimSuffix(basename, filepath.Ext(basename))

		_, err = pathsFile.WriteString(files[i][1] + "\n")
		if err != nil {
			return err
		}

		_, err = namesFile.WriteString(name + "\n")
		if err != nil {
			return err
		}
	}

	for _, f := range indexFiles {
		_ = f.Sync()
		_ = f.Close()
	}

	tmpIndexPath := filepath.Join(tmpDir, filepath.Base(config.SearchDbFile))

	indexTar, err := os.Create(tmpIndexPath)
	if err != nil {
		return err
	}

	tarw := tar.NewWriter(indexTar)
	defer func(tarw *tar.Writer) {
		_ = tarw.Close()
	}(tarw)

	tmpFiles, err := os.ReadDir(tmpFilesDir)
	if err != nil {
		return err
	}

	for _, indexFile := range tmpFiles {
		file, err := os.Open(filepath.Join(tmpFilesDir, indexFile.Name()))
		if err != nil {
			return err
		}

		fileInfo, err := indexFile.Info()
		if err != nil {
			_ = file.Close()
			return err
		}

		header := &tar.Header{
			Name:    indexFile.Name(),
			Size:    fileInfo.Size(),
			Mode:    int64(fileInfo.Mode()),
			ModTime: fileInfo.ModTime(),
		}

		err = tarw.WriteHeader(header)
		if err != nil {
			_ = file.Close()
			return err
		}

		if _, err := io.Copy(tarw, file); err != nil {
			_ = file.Close()
			return err
		}
	}

	err = utils.MoveFile(tmpIndexPath, indexPath)
	if err != nil {
		return err
	}

	err = os.RemoveAll(tmpDir)
	if err != nil {
		return err
	}

	return nil
}

type indexMap map[string]map[string][]string

type Index struct {
	Path  string
	files indexMap
}

func Open(indexPath string) (Index, error) {
	var index Index

	_, err := os.Stat(indexPath)
	if err != nil {
		return index, err
	}

	indexTar, err := os.Open(indexPath)
	if err != nil {
		return index, err
	}
	defer func(indexTar *os.File) {
		_ = indexTar.Close()
	}(indexTar)

	index.Path = indexPath
	index.files = make(map[string]map[string][]string)

	r := tar.NewReader(indexTar)
	for {
		header, err := r.Next()

		if err == io.EOF {
			break
		} else if err != nil {
			return index, err
		}

		if header.Typeflag == tar.TypeReg {
			bs := bufio.NewScanner(r)
			lines := make([]string, 0)

			for bs.Scan() {
				lines = append(lines, bs.Text())
			}

			if err := bs.Err(); err != nil {
				return index, err
			}

			hp := strings.Split(header.Name, "__")

			if len(hp) != 2 {
				return index, fmt.Errorf("invalid index file: %s", header.Name)
			}

			if _, ok := index.files[hp[0]]; !ok {
				index.files[hp[0]] = make(map[string][]string)
			}

			index.files[hp[0]][hp[1]] = lines
		}
	}

	return index, nil
}

type SearchResult struct {
	System string
	Name   string
	Path   string
}

func (idx *Index) searchSystemByNameGeneric(test func(string, string) bool, system string, query string) []SearchResult {
	var results []SearchResult
	for i, name := range idx.files[system]["names"] {
		if test(name, query) {
			results = append(results, SearchResult{
				System: system,
				Name:   name,
				Path:   idx.files[system]["paths"][i],
			})
		}
	}
	return results
}

func searchByNameTest(name string, query string) bool {
	return strings.Contains(strings.ToLower(name), query)
}

func (idx *Index) SearchAllByName(query string) []SearchResult {
	var results []SearchResult
	query = strings.ToLower(query)
	for _, system := range utils.SortedMapKeys(idx.files) {
		results = append(results, idx.searchSystemByNameGeneric(searchByNameTest, system, query)...)
	}
	return results
}

func (idx *Index) SearchSystemByName(system string, query string) []SearchResult {
	query = strings.ToLower(query)
	return idx.searchSystemByNameGeneric(searchByNameTest, system, query)
}

func searchByNameReTest(name string, query string) bool {
	re, err := regexp.Compile(query)
	if err != nil {
		return false
	}
	return re.MatchString(name)
}

func (idx *Index) SearchSystemByNameRe(system string, query string) []SearchResult {
	return idx.searchSystemByNameGeneric(searchByNameReTest, system, query)
}

func (idx *Index) SearchSystemByWords(system string, query string) []SearchResult {
	var results []SearchResult
	words := strings.Split(strings.ToLower(query), " ")
	if len(words) == 0 {
		return results
	}

	for i, name := range idx.files[system]["names"] {
		if searchByNameTest(name, words[0]) {
			results = append(results, SearchResult{
				System: system,
				Name:   name,
				Path:   idx.files[system]["paths"][i],
			})
		}
	}

	for _, word := range words[1:] {
		var newResults []SearchResult
		for _, result := range results {
			if searchByNameTest(result.Name, word) {
				newResults = append(newResults, result)
			}
		}
		results = newResults
	}

	return results
}

func (idx *Index) SearchAllByWords(query string) []SearchResult {
	var results []SearchResult
	for _, system := range utils.SortedMapKeys(idx.files) {
		results = append(results, idx.SearchSystemByWords(system, query)...)
	}
	return results
}

func (idx *Index) Total() int {
	total := 0
	for system := range idx.files {
		total += len(idx.files[system]["paths"])
	}
	return total
}

func (idx *Index) Systems() []string {
	return utils.SortedMapKeys(idx.files)
}
