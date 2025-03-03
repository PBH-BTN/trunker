#!/usr/bin/env bash
RUN_NAME="pbh.btn.trunker"
mkdir -p output

mkdir -p output/bin output/conf
cp script/* output/
chmod +x output/bootstrap.sh
cp conf/* output/conf/
if [ "$BUILD_TYPE" != "test" ]; then
    CGO_ENABLED=1 go build -trimpath -ldflags="-w -s -X 'main.Commit=$1'" -tags="gc_opt poll_opt re2_cgo" -o output/bin/${RUN_NAME}
else
    go build -trimpath -gcflags="all=-N -l -X 'main.Commit=$1'" -o output/bin/${RUN_NAME}
fi