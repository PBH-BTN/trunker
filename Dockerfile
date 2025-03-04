FROM golang:alpine AS build
ARG COMMIT_SHA
ARG VERSION
ARG TARGETPLATFORM
WORKDIR /build
COPY . .
RUN apk add build-base pkgconfig re2-dev bash
RUN export GOARCH=${TARGETPLATFORM#*/} && bash build.sh $VERSION $COMMIT_SHA

FROM alpine
WORKDIR /app
COPY --from=build /build/output .
RUN apk add --no-cache re2
ENTRYPOINT ["./bootstrap.sh"]
