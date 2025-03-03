FROM --platform=$BUILDPLATFORM golang:alpine AS build
ARG COMMIT_SHA
ARG TARGETPLATFORM
WORKDIR /build
COPY . .
RUN export GOARCH=${TARGETPLATFORM#*/}  && sh build.sh $COMMIT_SHA

FROM alpine
WORKDIR /app
COPY --from=build /build/output .
ENTRYPOINT ["./bootstrap.sh"]
