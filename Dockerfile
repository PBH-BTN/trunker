FROM golang:alpine AS build
ARG COMMIT_SHA
ARG TARGETPLATFORM
WORKDIR /build
COPY . .
RUN apk add build-base pkgconfig re2-dev
RUN export GOARCH=${TARGETPLATFORM#*/} && sh build.sh $COMMIT_SHA

FROM alpine
WORKDIR /app
COPY --from=build /build/output .
RUN apk add --no-cache re2
ENTRYPOINT ["./bootstrap.sh"]
