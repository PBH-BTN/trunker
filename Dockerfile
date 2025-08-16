FROM --platform=$BUILDPLATFORM golang:1.25 AS build
ARG COMMIT_SHA
ARG VERSION
ARG TARGETPLATFORM
WORKDIR /build
COPY . .
RUN bash build_docker.sh


FROM debian:stable-slim
WORKDIR /app
COPY --from=build /build/output .
RUN apt-get update && apt-get install -y 'libre2-[0-9]+' && rm -rf /var/lib/apt/lists/*
ENTRYPOINT ["./bootstrap.sh"]
