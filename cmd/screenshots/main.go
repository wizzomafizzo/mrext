package main

import (
	"fmt"

	"github.com/wizzomafizzo/mrext/pkg/framebuffer"
)

func main() {
	var fb framebuffer.Framebuffer

	err := fb.Open()
	if err != nil {
		panic(err)
	}
	defer fb.Close()

	fb.Fill(255, 255, 255, 0)
	fmt.Scanln()
	fb.Fill(0, 0, 0, 0)
}
