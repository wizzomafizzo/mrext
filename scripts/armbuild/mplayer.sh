#!/bin/bash

# example of running mplayer on framebuffer nicely:
# > chvt 2
# > echo -e '\e[?25l' > /dev/tty2
# > vmode -r 640 480 rgb32
# > LD_LIBRARY_PATH='/tmp' ./mplayer video.mp4
# > echo -e '\e[?25h' > /dev/tty2
# > chvt 1
# > echo 'load_core /media/fat/menu.rbf' > /dev/MiSTer_cmd
# another cursor hiding method:
# > echo 0 > /sys/class/graphics/fbcon/cursor_blink
# or:
# > setterm -cursor off

if [ ! -f MPlayer-1.5.tar.xz ]; then
    curl -o MPlayer-1.5.tar.xz http://www.mplayerhq.hu/MPlayer/releases/MPlayer-1.5.tar.xz
fi

if [ ! -d MPlayer-1.5 ]; then
    tar -xvf MPlayer-1.5.tar.xz
fi

cd MPlayer-1.5 || exit

# for some reason enabling static is unable to link libasound
./configure --prefix=/media/fat/linux --enable-fbdev \
  --enable-alsa --enable-openssl-nondistributable #--enable-static
make -j16

# need libtinfo to launch on mister with dynamic linked build
cp /lib/arm-linux-gnueabihf/libtinfo.so.6 ./libtinfo.so.6