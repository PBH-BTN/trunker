#!/usr/bin/env bash
RUN_NAME="pbh.btn.trunker"
mkdir -p output

mkdir -p output/bin output/conf
cp script/* output/
chmod +x output/bootstrap.sh
cp conf/* output/conf/
export GOEXPERIMENT=arenas
if [ "$BUILD_TYPE" != "test" ]; then
    go build -ldflags="-w -s" -o output/bin/${RUN_NAME}
else
    go build -gcflags="all=-N -l" -o output/bin/${RUN_NAME}
fi