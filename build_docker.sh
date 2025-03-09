#!/usr/bin/env bash
if [ "$TARGETPLATFORM" = "linux/arm64" ] ; then
  dpkg --add-architecture arm64
  apt-get update
  apt-get install -y crossbuild-essential-arm64 libre2-dev:arm64
  export GOARCH=arm64
  export CC=aarch64-linux-gnu-gcc
  export CXX=aarch64-linux-gnu-g++
  export AR=aarch64-linux-gnu-ar
  export PKG_CONFIG_PATH=/usr/lib/aarch64-linux-gnu/pkgconfig
else
  apt-get update
  apt-get install -y build-essential libre2-dev
fi
./build.sh $VERSION $COMMIT_SHA