#!/bin/bash

if [ ! -d Linux-Kernel_MiSTer ]; then
    git clone --depth 1 https://github.com/MiSTer-devel/Linux-Kernel_MiSTer.git
fi

cd Linux-Kernel_MiSTer || exit

export ARCH=arm
export LOCALVERSION=-MiSTer

make mrproper
make headers_install INSTALL_HDR_PATH=build/usr
