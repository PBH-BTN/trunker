#!/usr/bin/env bash
apt-get update
apt-get install -y clang llvm libbpf-dev
ln -s /usr/include/aarch64-linux-gnu/asm /usr/include/asm
go generate biz/services/peer/mux_local/ban/xdp/filter.go
echo "eBPF generated"
apt-get remove -y llvm clang
make install_tool && make update_idl
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
  apt-get install -y build-essential libre2-dev
  export GOAMD64=v4
fi
echo "start compiling"
./build.sh $VERSION $COMMIT_SHA
