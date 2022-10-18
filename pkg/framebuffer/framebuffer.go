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
	_, err := term.MakeRaw(int(os.Stdin.Fd()))
	if err != nil {
		return err
	}

	dev_file := C.CString("/dev/fb0")
	fd, err := C.openFrameBuffer(dev_file)
	C.free(unsafe.Pointer(dev_file))

	if err != nil {
		return fmt.Errorf("error opening framebuffer device: %v", err)
	}

	var finfo C.struct_fb_fix_screeninfo
	if _, err := C.getFixedScreenInfo(fd, &finfo); err != nil {
		return err
	}

	var vinfo C.struct_fb_var_screeninfo
	if _, err := C.getVarScreenInfo(fd, &vinfo); err != nil {
		return err
	}

	fb.bitsPerPixel = int(vinfo.bits_per_pixel)
	fb.xRes = int(vinfo.xres)
	fb.yRes = int(vinfo.yres)
	fb.xOffset = int(vinfo.xoffset)
	fb.yOffset = int(vinfo.yoffset)
	fb.lineLength = int(finfo.line_length)
	fb.screenSize = int(finfo.smem_len)

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
	C.munmap(unsafe.Pointer(&fb.data[0]), C.size_t(fb.screenSize))
	C.close(C.int(fb.fd))
}

func (fb *Framebuffer) ColorModel() color.Model {
	return color.RGBAModel
}

func (fb *Framebuffer) Bounds() image.Rectangle {
	return image.Rect(0, 0, fb.xRes, fb.yRes)
}

func (fb *Framebuffer) At(x, y int) color.Color {
	if x < 0 || x >= fb.xRes || y < 0 || y >= fb.yRes {
		return color.Black
	}

	addr := (fb.xOffset + x + (fb.yOffset+y)*fb.lineLength) * fb.bitsPerPixel / 8

	b := fb.data[addr]
	g := fb.data[addr+1]
	r := fb.data[addr+2]
	a := fb.data[addr+3]

	return color.RGBA{r, g, b, a}
}

func (fb *Framebuffer) Set(x, y int, c color.Color) {
	if x < 0 || x > fb.xRes || y < 0 || y > fb.yRes {
		return
	}

	r, g, b, a := c.RGBA()
	addr := (x+fb.xOffset)*(fb.bitsPerPixel/8) + (y+fb.yOffset)*fb.lineLength

	fb.data[addr] = byte(b)
	fb.data[addr+1] = byte(g)
	fb.data[addr+2] = byte(r)
	fb.data[addr+3] = byte(a)
}

func (fb *Framebuffer) Fill(c color.Color) {
	// draw.Draw(&fb, fb.Bounds(), &image.Uniform{color.Black}, image.Point{}, draw.Src)
	for y := 0; y < fb.yRes; y++ {
		for x := 0; x < fb.xRes; x++ {
			fb.Set(x, y, c)
		}
	}
}

func (fb *Framebuffer) ReadKey() (byte, error) {
	var b [1]byte
	_, err := os.Stdin.Read(b[:])
	return b[0], err
}
