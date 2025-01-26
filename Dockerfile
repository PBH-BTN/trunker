FROM golang:alpine AS build
COPY . /build
WORKDIR /build
RUN apk add --no-cache git
RUN sh build.sh

FROM alpine
WORKDIR /app
COPY --from=build /build/output .
ENTRYPOINT ["./bootstrap.sh"]
