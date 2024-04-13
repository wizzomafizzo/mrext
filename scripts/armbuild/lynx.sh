#!/bin/bash

if [ ! -f lynx-cur.tar.gz ]; then
    curl -o lynx-cur.tar.gz http://invisible-island.net/datafiles/release/lynx-cur.tar.gz
fi

if [ ! -d lynx-cur ]; then
    tar -xvf lynx-cur.tar.gz
fi

cd lynx2.9.0 || exit

./configure --with-ssl --enable-static
make -j16
