package main

import (
	"golang.org/x/image/font"
	"golang.org/x/image/font/basicfont"
	"golang.org/x/image/math/fixed"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"os"
	"os/exec"

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

func addLabel(img *image.RGBA, x, y int, label string) {
	col := color.RGBA{R: 200, G: 100, A: 255}
	point := fixed.Point26_6{X: fixed.I(x), Y: fixed.I(y)}

	d := &font.Drawer{
		Dst:  img,
		Src:  image.NewUniform(col),
		Face: basicfont.Face7x13,
		Dot:  point,
	}
	d.DrawString(label)
}

func drawImage(fb *framebuffer.Framebuffer, atX int, atY int) {
	labelImg := image.NewRGBA(image.Rect(0, 0, 100, 20))
	draw.Draw(labelImg, labelImg.Bounds(), &image.Uniform{C: color.RGBA{R: 255, A: 255}}, image.Point{}, draw.Src)
	draw.Draw(fb, labelImg.Bounds().Add(image.Point{X: atX, Y: atY}), labelImg, image.Point{}, draw.Src)
}

func drawBackgroundImage(fb *framebuffer.Framebuffer, img image.Image) {
	draw.Draw(fb, fb.Bounds(), img, image.Point{}, draw.Src)
}

func main() {
	_ = exec.Command("vmode", "-r", "640", "360", "rgb32").Run()

	var fb framebuffer.Framebuffer

	err := fb.Open()
	if err != nil {
		panic(err)
	}
	defer fb.Close()

	fb.Fill(color.Black)

	backgroundImg, err := getImageFromFilePath("/media/fat/monkey.png")
	if err != nil {
		panic(err)
	}

	bg := image.NewRGBA(backgroundImg.Bounds())
	draw.Draw(bg, bg.Bounds(), backgroundImg, image.Point{}, draw.Src)

	addLabel(bg, 100, 100, "Hello World!")

	drawBackgroundImage(&fb, backgroundImg)

	offsetX, offsetY := 10, 10
	drawImage(&fb, offsetX, offsetY)

	for {
		key, err := fb.ReadKey()
		if err != nil {
			panic(err)
		}

		if key[0] == 27 && key[1] == 91 {
			switch key[2] {
			case 65:
				// up
				offsetY -= 10
				drawBackgroundImage(&fb, backgroundImg)
				drawImage(&fb, offsetX, offsetY)
			case 66:
				// down
				offsetY += 10
				drawBackgroundImage(&fb, backgroundImg)
				drawImage(&fb, offsetX, offsetY)
			case 67:
				// right
				offsetX += 10
				drawBackgroundImage(&fb, backgroundImg)
				drawImage(&fb, offsetX, offsetY)
			case 68:
				// left
				offsetX -= 10
				drawBackgroundImage(&fb, backgroundImg)
				drawImage(&fb, offsetX, offsetY)
			}
		} else if key[0] == 113 {
			break
		}
	}
}
