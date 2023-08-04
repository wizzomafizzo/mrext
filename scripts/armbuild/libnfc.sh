#!/bin/bash

## install libnfc dependencies
#RUN apt-get install -y libusb-dev libtool autoconf automake unzip
## install custom version of libnfc
#COPY patches /patches
#RUN mkdir /internal && cd /internal && \
#    wget https://github.com/nfc-tools/libnfc/releases/download/libnfc-1.8.0/libnfc-1.8.0.tar.bz2 && \
#    tar xvf libnfc-1.8.0.tar.bz2 && \
#    cd libnfc-1.8.0 && \
#    cp /patches/acr122u-fix.patch . && patch -p1 < acr122u-fix.patch && \
#    autoreconf -vis && \
#    ./configure && \
#    make -j "$(nproc)" && \
#    make install

if [ ! -d libnfc ]; then
    git clone --depth 1 https://github.com/sam1902/libnfc
fi

cd libnfc || exit

autoreconf -vis
./configure --prefix=/build/build --enable-static --with-drivers=all
make -j "$(nproc)"
make install
