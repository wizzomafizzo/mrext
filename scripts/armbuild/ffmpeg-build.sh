#!/bin/bash

git clone --depth 1 git://source.ffmpeg.org/ffmpeg.git
cd ffmpeg
./configure --prefix=./build/ --enable-gpl --enable-static --enable-pic --extra-libs="-ldl"
make -j8