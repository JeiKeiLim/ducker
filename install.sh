#!/bin/bash
#
# Shell script for install and uninstall ducker
#
# - Author: Jongkuk Lim
# - Contact: lim.jeikei@gmail.com
#
# Usage)
# 1. Install
#   - install.sh install $OS $ARCH $VERSION
#
# 2. Uninstall
#   - install.sh uninstall

VERSION=$(curl -s https://raw.githubusercontent.com/JeiKeiLim/ducker/main/VERSION)
OS=linux
ARCH=amd64

if [ "$1" = "install" ]; then
    # Check Version, OS, Archiecture from arguments
    if [ -n "$2" ]; then
        OS=$2
        if [ -n "$3" ]; then
            ARCH=$3
            if [ -n "$4" ]; then
                VERSION=$4
            fi
        fi
    fi

    # Ducker already installed
    if [ -f "/usr/local/bin/ducker" ]; then
        CURRENT_VERSION=$(ducker -v | tail -n 1 | cut -d " " -f 3)
        echo "ducker has been found on your system."
        echo "Current: $CURRENT_VERSION"
        echo "Install: $VERSION"
        read -p "Do you want to continue to install? [y/n] " -n 1 -r </dev/tty
        echo
        if [[ ! $REPLY =~ ^[Yy]$  ]]
        then
                exit 1
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
else
    echo "Wrong command $0"
    echo "Please use install.sh install or install.sh uninstall"
    exit 1
fi

echo "Install ducker-$VERSION-$OS-$ARCH ..."
wget -q https://github.com/JeiKeiLim/ducker/releases/download/v$VERSION/ducker-$VERSION-$OS-$ARCH.tar.gz
if [ ! -f ducker-$VERSION-$OS-$ARCH.tar.gz ]; then
    echo "ducker-$VERSION-$OS-$ARCH can not be found."
    echo "Please check correct version at https://github.com/JeiKeiLim/ducker/releases"
    exit 1
fi
tar xzf ducker-$VERSION-$OS-$ARCH.tar.gz
sudo mv ducker /usr/local/bin/
rm ducker-$VERSION-$OS-$ARCH.tar.gz

if [ -f /usr/local/bin/ducker ]; then
    echo "ducker has been successfully installed!"
    exit 0
else
    echo "ducker has not been installed."
    exit 1
fi
