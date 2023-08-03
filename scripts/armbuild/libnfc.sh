#!/bin/bash

if [ ! -d libnfc ]; then
    git clone --depth 1 https://github.com/sam1902/libnfc
fi

cd libnfc || exit

autoreconf -vis
./configure --prefix=/build/build --enable-static --with-drivers=all
make -j "$(nproc)"
make install
