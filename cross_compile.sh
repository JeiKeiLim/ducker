#!/bin/bash

VERSION="0.1.3"
OSS=("darwin" "linux")
ARCHS=("amd64" "arm64" "arm")

for os in ${OSS[@]}; do
    for arch in ${ARCHS[@]}; do
        echo "Compile $os-$arch ..."
        env GOOS=$os GOARCH=$arch go build ./cmd/ducker
        if [ $? -eq 0 ]; then
            tar -czf ducker-$VERSION-$os-$arch.tar.gz ./ducker
        fi
    done
done
