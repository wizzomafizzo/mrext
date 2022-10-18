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
	"unsafe"
)

type Framebuffer struct {
	Fd           int
	BitsPerPixel int
	XRes         int
	YRes         int
	Data         []byte
	XOffset      int
	YOffset      int
	LineLength   int
	ScreenSize   int
}

func (fb *Framebuffer) Open() error {
	// hide cursor
	fmt.Print("\033[?25l")

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

	fb.BitsPerPixel = int(vinfo.bits_per_pixel)
	fb.XRes = int(vinfo.xres)
	fb.YRes = int(vinfo.yres)
	fb.XOffset = int(vinfo.xoffset)
	fb.YOffset = int(vinfo.yoffset)
	fb.LineLength = int(finfo.line_length)
	fb.ScreenSize = int(finfo.smem_len)

	addr := uintptr(C.mmap(nil, C.size_t(fb.ScreenSize), C.PROT_READ|C.PROT_WRITE, C.MAP_SHARED, fd, 0))

	var sl = struct {
		addr uintptr
		len  int
		cap  int
	}{addr, fb.ScreenSize, fb.ScreenSize}

	fb.Data = *(*[]byte)(unsafe.Pointer(&sl))

	return nil
}

func (fb *Framebuffer) Close() {
	C.munmap(unsafe.Pointer(&fb.Data[0]), C.size_t(fb.ScreenSize))
	C.close(C.int(fb.Fd))
}

func (fb *Framebuffer) Set(x int, y int, r uint32, g uint32, b uint32, a uint32) error {
	if x < 0 || x > fb.XRes || y < 0 || y > fb.YRes {
		return fmt.Errorf("pixel coords out of bounds: %d, %d", x, y)
	}

	location := (x+fb.XOffset)*(fb.BitsPerPixel/8) + (y+fb.YOffset)*fb.LineLength

	fb.Data[location+3] = byte(a & 0xff)
	fb.Data[location+2] = byte(r & 0xff)
	fb.Data[location+1] = byte(g & 0xff)
	fb.Data[location] = byte(b & 0xff)

	return nil
}

func (fb *Framebuffer) Fill(r, g, b, a uint32) {
	for y := 0; y < fb.YRes; y++ {
		for x := 0; x < fb.XRes; x++ {
			fb.Set(x, y, r, g, b, a)
		}
	}
}

func (fb *Framebuffer) SetImage(xOffset int, yOffset int, image image.Image) {
	bounds := image.Bounds()

	for y := 0; y < bounds.Max.Y; y++ {
		for x := 0; x < bounds.Max.X; x++ {
			r, g, b, a := image.At(x, y).RGBA()
			fb.Set(x+xOffset, y+yOffset, r&0xff, g&0xff, b&0xff, a&0xff)
		}
	}
}
