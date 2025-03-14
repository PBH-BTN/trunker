# Trunker

> A Tracker which will "chuang" you

![image](https://github.com/user-attachments/assets/6f3676a8-4b51-4f14-9107-d08a35868238)

## Introduction

A high-performance BitTorrent Tracker implemented in Go. Using [Hertz](https://github.com/cloudwego/hertz) from cloudwego, with [observability](./docs/metrics.adoc).

For benchmark, please refer to the [Benchmark](#benchmark) section.

### Official Instance 
HTTPS `https://tracker.ghostchu-services.top/announce`  
HTTP `http://tracker.ghostchu-services.top/announce`  
WebSocket (for WebTorrent Protocol) `wss://tracker.ghostchu-services.top/announce`  
UDP `udp://utracker.ghostchu-services.top:6969`

## How to run

```bash
make install_tool
make update_idl
./build.sh
cd output
./bootstrap.sh
```
or
```
docker run -d --name trunker -e ADMIN_KEY=aabbcc -p 8888:8888 gaojianli2333/trunker:latest
```
## Features

- [x] [BEP-0003](https://www.bittorrent.org/beps/bep_0003.html)
- [x] [BEP-0007](https://www.bittorrent.org/beps/bep_0007.html) (IPv6 Tracker Extension)
- [x] [BEP-0023](https://www.bittorrent.org/beps/bep_0023.html) (Compact Peer Lists)
- [x] [BEP-0024](https://www.bittorrent.org/beps/bep_0024.html) (External IP)
- [x] [BEP-0031](https://www.bittorrent.org/beps/bep_0031.html) (Failure Retry Extension)
- [x] [BEP-0048](https://www.bittorrent.org/beps/bep_0048.html) (Scrape)
- [x] [BEP-0015](https://www.bittorrent.org/beps/bep_0015.html) (UDP Tracker Protocol)
- [X] LT-Extension (aka. complete,incomplete)
- [x] Switchable Mode (Memory or MySQL)
- [x] Load and store persist from disk
- [x] Blacklist for info_hash and peer_id
- [x] Eventbus support
- [x] Websocket support
- [x] Prometheus based [metrics](./docs/metrics.adoc)
- [x] RPC mode to support cluster. Powered by [Kitex](https://github.com/cloudwego/kitex)

## Wiki
Trunker provides much config and observability capabilities, if you want to run trunker in production, please see the [Wiki](./docs/toc.adoc).

## Benchmark

Trunker has very strong performance. Here's a record of a real peak.

- CPU: `4 Cores AMD EPYC-Milan`
- Average response time: `100us` when `1267703` torrents and `2406393` peers are online.
- QPS: `2736` (can be higher, but we don't have such many peers connect to our tracker)
- Memory Cost: `2635MB`.

![image](https://github.com/user-attachments/assets/920bd461-9389-4aa4-b0bd-1dee26dc9325)

![image](https://github.com/user-attachments/assets/25bf44e5-8a10-4aec-84cc-02e39b3b9bbc)
