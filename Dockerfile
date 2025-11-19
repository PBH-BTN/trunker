FROM --platform=$BUILDPLATFORM golang:1.25 AS build
ARG COMMIT_SHA
ARG VERSION
ARG TARGETPLATFORM
WORKDIR /build

COPY . .

# Install bpf2go for generating eBPF code
RUN go install github.com/cilium/ebpf/cmd/bpf2go@latest

# Generate eBPF code (will be no-op on non-Linux platforms due to build tags)
RUN cd biz/services/peer/mux_local/ban/xdp && go generate || true

# Build the application
RUN bash build_docker.sh


FROM debian:stable-slim
WORKDIR /app

RUN apt-get update && \
    apt-get install -y libre2-11 libelf1 && \
    rm -rf /var/lib/apt/lists/*

COPY --from=build /build/output .

ENTRYPOINT ["./bootstrap.sh"]
