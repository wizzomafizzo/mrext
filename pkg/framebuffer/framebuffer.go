package framebuffer

/*
#include <stdlib.h>
#include <unistd.h>
#include <fcntl.h>
#include <sys/mman.h>
#include <sys/ioctl.h>
#include <linux/fb.h>

int openFrameBuffer(char *name) {
	return open(name, O_RDWR);
}

static int getFixedScreenInfo(int fd, struct fb_fix_screeninfo *finfo) {
	return ioctl(fd, FBIOGET_FSCREENINFO, finfo);
}

static int getVarScreenInfo(int fd, struct fb_var_screeninfo *vinfo) {
	return ioctl(fd, FBIOGET_VSCREENINFO, vinfo);
}
*/
import "C"
import (
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"os"
	"unsafe"

	"golang.org/x/term"
)

type Framebuffer struct {
	fd           int
	bitsPerPixel int
	xRes         int
	yRes         int
	data         []byte
	xOffset      int
	yOffset      int
	lineLength   int
	screenSize   int
}

func (fb *Framebuffer) Open() error {
	fmt.Print("\033[?25l")
	_, err := term.MakeRaw(int(os.Stdin.Fd()))
	if err != nil {
		return err
	}

	devFile := C.CString("/dev/fb0")
	fd, err := C.openFrameBuffer(devFile)
	C.free(unsafe.Pointer(devFile))

	if err != nil {
		return err
	}

	var fixInfo C.struct_fb_fix_screeninfo
	if _, err := C.getFixedScreenInfo(fd, &fixInfo); err != nil {
		return err
	}

	var varInfo C.struct_fb_var_screeninfo
	if _, err := C.getVarScreenInfo(fd, &varInfo); err != nil {
		return err
	}

	fb.bitsPerPixel = int(varInfo.bits_per_pixel)
	fb.xRes = int(varInfo.xres)
	fb.yRes = int(varInfo.yres)
	fb.xOffset = int(varInfo.xoffset)
	fb.yOffset = int(varInfo.yoffset)
	fb.lineLength = int(fixInfo.line_length)
	fb.screenSize = int(fixInfo.smem_len)

	addr := uintptr(C.mmap(nil, C.size_t(fb.screenSize), C.PROT_READ|C.PROT_WRITE, C.MAP_SHARED, fd, 0))

	var sl = struct {
		addr uintptr
		len  int
		cap  int
	}{addr, fb.screenSize, fb.screenSize}

	fb.data = *(*[]byte)(unsafe.Pointer(&sl))

	return nil
}

func (fb *Framebuffer) Close() {
	fb.Fill(color.Black)
	C.munmap(unsafe.Pointer(&fb.data[0]), C.size_t(fb.screenSize))
	C.close(C.int(fb.fd))
}

func (fb *Framebuffer) ColorModel() color.Model {
	return color.RGBAModel
}

func (fb *Framebuffer) Bounds() image.Rectangle {
	return image.Rect(0, 0, fb.xRes, fb.yRes)
}

func (fb *Framebuffer) addressAt(x, y int) int {
	return (x+fb.xOffset)*(fb.bitsPerPixel/8) + (y+fb.yOffset)*fb.lineLength
}

func (fb *Framebuffer) At(x, y int) color.Color {
	if x < 0 || x >= fb.xRes || y < 0 || y >= fb.yRes {
		return color.Black
	}

	addr := fb.addressAt(x, y)
	b := fb.data[addr]
	g := fb.data[addr+1]
	r := fb.data[addr+2]
	a := fb.data[addr+3]

	return color.RGBA{R: r, G: g, B: b, A: a}
}

func (fb *Framebuffer) Set(x, y int, c color.Color) {
	if x < 0 || x > fb.xRes || y < 0 || y > fb.yRes {
		return
	}

	r, g, b, a := c.RGBA()
	addr := fb.addressAt(x, y)
	fb.data[addr] = byte(b)
	fb.data[addr+1] = byte(g)
	fb.data[addr+2] = byte(r)
	fb.data[addr+3] = byte(a)
}

func (fb *Framebuffer) Fill(c color.Color) {
	draw.Draw(fb, fb.Bounds(), &image.Uniform{C: c}, image.Point{}, draw.Src)
}

func (fb *Framebuffer) ReadKey() ([3]byte, error) {
	var b [3]byte
	_, err := os.Stdin.Read(b[:])
	return b, err
}
