#!/bin/bash

VERSION=0.1.3
OS=linux
ARCH=amd64

echo $@
exit 0;

wget https://github.com/JeiKeiLim/ducker/releases/download/v$VERSION/ducker-$VERSION-$OS-$ARCH.tar.gz
tar xzvf ducker-$VERSION-$OS-$ARCH.tar.gz
sudo mv ducker /usr/local/bin/
rm ducker-$VERSION-$OS-$ARCH.tar.gz

