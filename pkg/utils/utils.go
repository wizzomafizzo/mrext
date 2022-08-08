package utils

import (
	"archive/zip"
	"io"
	"math/rand"
	"os"
	"path/filepath"
	"strings"
)

func IsZip(path string) bool {
	if filepath.Ext(strings.ToLower(path)) == ".zip" {
		return true
	} else {
		return false
	}
}

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

func MaxInt(xs []int) int {
	max := 0
	for _, x := range xs {
		if x > max {
			max = x
		}
	}
	return max
}

func MinInt(xs []int) int {
	min := int(^uint(0) >> 1)
	for _, x := range xs {
		if x < min {
			min = x
		}
	}
	return min
}

func Contains[T comparable](xs []T, x T) bool {
	for _, v := range xs {
		if v == x {
			return true
		}
	}
	return false
}

func RandomItem[T any](xs []T) (T, error) {
	var item T
	if len(xs) == 0 {
		return item, nil
	} else {
		item = xs[rand.Intn(len(xs))]
		return item, nil
	}
}

func MapKeys[K comparable, V any](m map[K]V) []K {
	keys := make([]K, len(m))
	i := 0
	for k := range m {
		keys[i] = k
		i++
	}
	return keys
}
