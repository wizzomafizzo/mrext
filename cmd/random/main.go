package main

import (
	"fmt"
	"math/rand"
	"os"

	"github.com/wizzomafizzo/mrext/pkg/games"
)

func pickRandomFile(path string) string {
	files, err := os.ReadDir(path)
	if err != nil {
		return ""
	}

	if len(files) == 0 {
		return ""
	}

	return files[rand.Intn(len(files))].Name()
}

func main() {
	folders := games.GetSystemPaths()
	if len(folders) == 0 {
		fmt.Println("No games found, exiting...")
		return
	}

}
