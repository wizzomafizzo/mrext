#!/bin/bash

if [ ! -f aria2-1.37.0.tar.gz ]; then
    curl -o aria2-1.37.0.tar.gz https://github.com/aria2/aria2/releases/download/release-1.37.0/aria2-1.37.0.tar.gz
fi

if [ ! -d aria2-1.37.0 ]; then
    tar -xvf aria2-1.37.0.tar.gz
fi

cd aria2-1.37.0 || exit

./configure --enable-static --without-libcares --without-appletls --without-wintls --without-sqlite3
make -j16
