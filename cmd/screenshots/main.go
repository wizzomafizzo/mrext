package main

import (
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"os"

	"github.com/wizzomafizzo/mrext/pkg/framebuffer"
)

func getImageFromFilePath(filePath string) (image.Image, error) {
	f, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	image, err := png.Decode(f)
	return image, err
}

func main() {
	var fb framebuffer.Framebuffer

	err := fb.Open()
	if err != nil {
		panic(err)
	}
	defer fb.Close()

	fb.Fill(color.White)

	img, err := getImageFromFilePath("/media/fat/screenshots/MENU/20221004_144640-screen.png")
	if err != nil {
		panic(err)
	}

	draw.Draw(&fb, fb.Bounds(), img, image.Point{}, draw.Src)

	fb.ReadKey()
}
