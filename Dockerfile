FROM golang:alpine as build
WORKDIR /build
COPY . .
RUN sh build.sh

FROM alpine
WORKDIR /app
COPY --from=build /build/output .
ENTRYPOINT ["./bootstrap.sh"]
