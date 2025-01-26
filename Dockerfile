FROM golang:alpine AS build
ARG COMMIT_SHA
WORKDIR /build
COPY . .
RUN apk add --no-cache git
RUN sh build.sh $COMMIT_SHA

FROM alpine
WORKDIR /app
COPY --from=build /build/output .
ENTRYPOINT ["./bootstrap.sh"]
