#!/bin/bash
#
# Shell script for install and uninstall ducker
#
# - Author: Jongkuk Lim
# - Contact: limjk@jmarple.ai
#
# Usage)
# 1. Install
#   - install.sh install $OS $ARCH $VERSION
#
# 2. Uninstall
#   - install.sh uninstall

VERSION=0.1.3
OS=linux
ARCH=amd64

if [ "$1" = "install" ]; then
    if [ -n "$2" ]; then
        OS=$2
        if [ -n "$3" ]; then
            ARCH=$3
            if [ -n "$4" ]; then
                VERSION=$4
            fi
        fi
    fi
elif [ "$1" = "uninstall" ]; then
    echo "Uninstall ducker ..."
    sudo rm /usr/local/bin/ducker
    if [ $? -ne 0 ]; then
        echo "Failed to uninstall ducker."
        exit 1
    else
        echo "ducker has been removed!"
        exit 0
    fi
fi

echo "Install ducker-$VERSION-$OS-$ARCH ..."

wget https://github.com/JeiKeiLim/ducker/releases/download/v$VERSION/ducker-$VERSION-$OS-$ARCH.tar.gz
tar xzvf ducker-$VERSION-$OS-$ARCH.tar.gz
sudo mv ducker /usr/local/bin/
rm ducker-$VERSION-$OS-$ARCH.tar.gz

if [ -f /usr/local/bin/ducker ]; then
    echo "ducker has been successfully installed!"
    exit 0
else
    echo "ducker has not been installed."
    exit 1
fi
