#!/bin/bash

if [ ! -d ffmpeg ]; then
    git clone --depth 1 git://source.ffmpeg.org/ffmpeg.git
fi

cd ffmpeg || exit

# TODO: these options need a proper check through, just stole them somewhere
./configure --prefix=./build/ --enable-gpl --enable-static --enable-pic --extra-libs="-ldl"
make -j16