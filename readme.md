# Trunker

> A Tracker which will "chuang" you

![image](https://github.com/user-attachments/assets/6f3676a8-4b51-4f14-9107-d08a35868238)


## Introduction

A high-performance BitTorrent Tracker implemented in Go. Using [Hertz](https://github.com/cloudwego/hertz) from cloudwego, with observability.

This tracker is hosted as https://btn-prod.ghostchu-services.top/announce

For benchmark, please refer to the [Benchmark](#benchmark) section.

## How to run

```bash
./build.sh
cd output
./bootstrap.sh
```
or
```
docker pull gaojianli2333/trunker:latest
```
## Features

- [x] [BEP-0003](https://www.bittorrent.org/beps/bep_0003.html)
- [x] [BEP-0007](https://www.bittorrent.org/beps/bep_0007.html) (IPv6 Tracker Extension)
- [x] [BEP-0023](https://www.bittorrent.org/beps/bep_0023.html) (Compact Peer Lists)
- [x] [BEP-0024](https://www.bittorrent.org/beps/bep_0024.html) (External IP)
- [x] [BEP-0031](https://www.bittorrent.org/beps/bep_0031.html) (Failure Retry Extension)
- [x] [BEP-0048](https://www.bittorrent.org/beps/bep_0048.html) (Scrape)
- [X] LT-Extension (complete,incomplete)
- [x] Full-Memory Mode
- [x] Load from Persist
- [x] MySQL Mode

## Wiki
If you want to run trunker by yourself, please see the [Wiki](./docs/toc.adoc).

## Benchmark

CPU: 4 Cores ARM64 Oracle Cloud

Average response time: 600μs when 30K torrents and 27K peers are online.

QPS: 700~ (can be higher but we don't have such many peers connect to our tracker)

Memory Cost: 348MB.

![image](https://github.com/user-attachments/assets/746babae-1eb3-4944-afb4-f629f78a007d)
