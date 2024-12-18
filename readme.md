# BTN-Server

## Introduce

A BitTorrent Tracker implemented in Go. Using [Hertz](https://github.com/cloudwego/hertz) from cloudwego.

## How to run

```bash
./build.sh
cd output
./bootstrap.sh
```

## Features
- [x] BEP-0003
- [x] BEP-0007
- [x] BEP-0023
- [x] BEP-0024
- [x] BEP-0031
- [x] BEP-0048
- [X] LT-Extension(complete,incomplete)
- [x] Full-Memory Mode
- [x] Load from Persist
- [ ] MySQL Mode

## Benchmark
CPU: 4 Cores ARM64 Oracle Cloud

Average response time: 600ms when 30K torrents and 27K peers are online. 

QPS: 700~

Memory Cost: 348MB.

![image](https://github.com/user-attachments/assets/00526a7c-1907-4949-a246-0ce6fab6302f)
