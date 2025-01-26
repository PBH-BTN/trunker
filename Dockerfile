FROM golang:alpine AS build
WORKDIR /build
ADD . .
RUN apk add --no-cache git
RUN sh build.sh

FROM alpine
WORKDIR /app
COPY --from=build /build/output .
ENTRYPOINT ["./bootstrap.sh"]
