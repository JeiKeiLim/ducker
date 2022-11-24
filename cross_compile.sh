#!/bin/bash

OSS=("darwin" "linux")
ARCHS=("amd64" "arm64" "arm")
VERSION=$(cat VERSION)

for os in ${OSS[@]}; do
    for arch in ${ARCHS[@]}; do
        echo -n "Compile $os-$arch ... "
        if env GOOS=$os GOARCH=$arch go build -ldflags "-X main.version=$VERSION" ./cmd/ducker; then
            tar -czf ducker-$VERSION-$os-$arch.tar.gz ./ducker
            echo "ducker-$VERSION-$os-$arch.tar.gz has been created."
        fi
    done
done
