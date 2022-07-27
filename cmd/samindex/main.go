package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/wizzomafizzo/mext/pkg/games"
	"github.com/wizzomafizzo/mext/pkg/utils"
)

func main() {
	start := time.Now()

	gameFiles := games.GetSystemFiles(func(system string) {
		fmt.Println("Scanning", system)
	})

	fmt.Println(len(gameFiles), "games found")

	tmpDir, err := os.MkdirTemp(os.TempDir(), "sam-")
	if err != nil {
		panic(err)
	}

	fmt.Println("Creating game lists")
	listFiles := make(map[string]*os.File)
	for _, game := range gameFiles {
		systemId, path := game[0], game[1]

		if _, ok := listFiles[systemId]; !ok {
			filename := strings.ToLower(systemId) + "_gamelist.txt"

			file, err := os.Create(filepath.Join(tmpDir, filename))
			if err != nil {
				panic(err)
			}
			defer file.Close()

			listFiles[systemId] = file
		}

		listFiles[systemId].WriteString(path + "\n")
	}

	pwd, err := os.Getwd()
	if err != nil {
		panic(err)
	}

	lists, err := os.ReadDir(tmpDir)
	if err != nil {
		panic(err)
	}

	fmt.Println("Writing to disk")
	for _, file := range lists {
		src := filepath.Join(tmpDir, file.Name())
		dest := filepath.Join(pwd, file.Name())

		if err := utils.MoveFile(src, dest); err != nil {
			panic(err)
		}
	}

	if err := os.RemoveAll(tmpDir); err != nil {
		panic(err)
	}

	fmt.Printf("Completed in %d seconds\n", int(time.Since(start).Seconds()))
}
