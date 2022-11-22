#!/bin/bash

VERSION="0.1.3"
OSS=("darwin" "linux")
ARCHS=("amd64" "arm64" "arm")

for os in ${OSS[@]}; do
    for arch in ${ARCHS[@]}; do
        echo -n "Compile $os-$arch ... "
        if env GOOS=$os GOARCH=$arch go build ./cmd/ducker; then
            tar -czf ducker-$VERSION-$os-$arch.tar.gz ./ducker
            echo "ducker-$VERSION-$os-$arch.tar.gz has been created."
        fi
    done
done
