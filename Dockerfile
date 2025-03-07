FROM --platform=$BUILDPLATFORM golang:1.23 AS build
ARG COMMIT_SHA
ARG VERSION
ARG TARGETPLATFORM
WORKDIR /build
COPY . .
RUN apt-get update && apt-get install -y libre2-dev && if [ "$TARGETPLATFORM" -eq "linux/arm64" ] ; then apt-get install -y crossbuild-essential-aarch64;else apt-get install -y build-essential; fi
RUN export GOARCH=${TARGETPLATFORM#*/} && if [ "$GOARCH" -ne $(uname -m) ]; then export CC=aarch64-linux-gnu-gcc;fi && bash build.sh $VERSION $COMMIT_SHA

FROM debian:stable-slim
WORKDIR /app
COPY --from=build /build/output .
RUN apt-get update && apt-get install -y re2 && rm -rf /var/lib/apt/lists/*
ENTRYPOINT ["./bootstrap.sh"]
