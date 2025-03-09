#!/usr/bin/env bash
RUN_NAME="pbh.btn.trunker"
mkdir -p output

mkdir -p output/bin output/conf
cp script/* output/
chmod +x output/bootstrap.sh
cp conf/* output/conf/

if [ -z $VERSION];then
  if [ -z $1 ];then
    VERSION="$(git describe --tags --always 2> /dev/null)"
  else
    VERSION=$1
  fi
fi
if [ -z $COMMIT_HASH];then
  if [ -z $2 ];then
      COMMIT_HASH="$(git rev-parse --short HEAD)"
  else
      COMMIT_HASH=$2
  fi
fi
BUILD_TIMESTAMP=$(date '+%Y-%m-%d %H:%M:%S')
LDFLAGS=(
  "-X 'main.Version=${VERSION}'"
  "-X 'main.Commit=${COMMIT_HASH}'"
  "-X 'main.BuildTimestamp=${BUILD_TIMESTAMP}'"
)
if [ "$BUILD_TYPE" != "test" ]; then
    CGO_ENABLED=1 go build -trimpath -ldflags="-w -s ${LDFLAGS[*]}" -tags="gc_opt poll_opt re2_cgo" -o output/bin/${RUN_NAME}
else
    go build -trimpath -gcflags="all=-N -l ${LDFLAGS[*]}" -o output/bin/${RUN_NAME}
fi