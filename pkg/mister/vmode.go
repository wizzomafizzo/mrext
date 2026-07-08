// mrext
// Copyright (c) 2026 mrext contributors.
// SPDX-License-Identifier: GPL-3.0-or-later
//
// This file is part of mrext.
//
// mrext is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// mrext is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with mrext. If not, see <http://www.gnu.org/licenses/>.

package mister

import (
	"fmt"
	"os"
	"syscall"
	"unsafe"
)

const (
	FrameBufferDevice         = "/dev/fb0"
	framebufferGetVScreenInfo = 0x4600
	framebufferInfoWords      = 64
)

// GetScreenResolution returns MiSTer's current framebuffer output resolution.
func GetScreenResolution() (int, int, error) {
	// #nosec G304 -- framebuffer path is a fixed MiSTer device.
	framebuffer, err := os.Open(FrameBufferDevice)
	if err != nil {
		return 0, 0, fmt.Errorf("open framebuffer: %w", err)
	}
	defer func() { _ = framebuffer.Close() }()

	// Linux fb_var_screeninfo contains 40 uint32 fields beginning with xres and
	// yres. Extra capacity ensures ioctl cannot overrun this buffer.
	var info [framebufferInfoWords]uint32
	_, _, errno := syscall.Syscall(
		syscall.SYS_IOCTL,
		framebuffer.Fd(),
		framebufferGetVScreenInfo,
		uintptr(unsafe.Pointer(&info[0])), // #nosec G103 -- ioctl requires kernel ABI pointer.
	)
	if errno != 0 {
		return 0, 0, fmt.Errorf("read framebuffer mode: %w", errno)
	}

	return int(info[0]), int(info[1]), nil
}
