package utils

import (
	"archive/zip"
	"io"
	"os"
)

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
