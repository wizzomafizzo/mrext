package utils

import (
	"archive/zip"
	"fmt"
	"io"
	"math/rand"
	"os"
	"path/filepath"
	"strings"

	"golang.org/x/exp/constraints"
)

func IsZip(path string) bool {
	// TODO: this should check the file header
	return filepath.Ext(strings.ToLower(path)) == ".zip"
}

// Return a slice of all filenames in a zip file.
func ListZip(path string) ([]string, error) {
	r, err := zip.OpenReader(path)
	if err != nil {
		return nil, err
	}
	defer r.Close()

	var files []string
	for _, f := range r.File {
		files = append(files, f.Name)
	}

	return files, nil
}

// Move a file. Supports moving between filesystems.
func MoveFile(sourcePath, destPath string) error {
	inputFile, err := os.Open(sourcePath)
	if err != nil {
		return err
	}

	outputFile, err := os.Create(destPath)
	if err != nil {
		inputFile.Close()
		return err
	}

	defer outputFile.Close()
	_, err = io.Copy(outputFile, inputFile)
	inputFile.Close()
	if err != nil {
		return err
	}

	err = os.Remove(sourcePath)
	if err != nil {
		return err
	}

	return nil
}

// Return the highest value in a slice.
func Max[T constraints.Ordered](xs []T) T {
	if len(xs) == 0 {
		var zv T
		return zv
	}
	max := xs[0]
	for _, x := range xs {
		if x > max {
			max = x
		}
	}
	return max
}

// Return the lowest value in a slice.
func Min[T constraints.Ordered](xs []T) T {
	if len(xs) == 0 {
		var zv T
		return zv
	}
	min := xs[0]
	for _, x := range xs {
		if x < min {
			min = x
		}
	}
	return min
}

// Return true if slice contains value.
func Contains[T comparable](xs []T, x T) bool {
	for _, v := range xs {
		if v == x {
			return true
		}
	}
	return false
}

// Pick and return a random element from a slice.
func RandomElem[T any](xs []T) (T, error) {
	var item T
	if len(xs) == 0 {
		return item, fmt.Errorf("empty slice")
	} else {
		item = xs[rand.Intn(len(xs))]
		return item, nil
	}
}

// Return a list of all keys in a map.
func MapKeys[K comparable, V any](m map[K]V) []K {
	keys := make([]K, len(m))
	i := 0
	for k := range m {
		keys[i] = k
		i++
	}
	return keys
}
