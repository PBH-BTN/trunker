# Trunker

> A Tracker which will "chuang" you

![image](https://github.com/user-attachments/assets/6f3676a8-4b51-4f14-9107-d08a35868238)


## Introduction

A BitTorrent Tracker implemented in Go. Using [Hertz](https://github.com/cloudwego/hertz) from cloudwego.

This tracker is hosted as https://btn-prod.ghostchu-services.top/announce

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

## Configuration

| Name                | Description                                                                                    | Default        |
|:--------------------|:-----------------------------------------------------------------------------------------------|:---------------|
| ttl                 | The time of a peer announce next time.                                                         | 3600s          |
| invervalTask        | The interval of the task to clean the expired peers.                                           | 600s           |
| useDB               | Use the database to store the data.  (currently no usable)                                     | false          |
| enablePersist       | Save the peers to the disk and load them while launching.                                      | true           |
| maxPeersPerTorrent  | The max number of peers per torrent.                                                           | 100            |
| shard               | The number of shards. More shards will improve response time, by cost more memory.             | 16             |
| useUnixSocket       | Use Unix Socket instead tcp                                                                    | false          |
| hostPorts           | The host and port of the server, For unix socket mode, this is the file name of the sock file. | 127.0.0.1:8888 |
| useAnnounceIP       | Allow peer to announce its IP in the query string.                                             | true           |
| enableEventProducer | Send peer event to the mq. Caution: This will produce tons of message.                         | false          |

JSON config is also supported with env `TRUNKER_CONFIG`.
## Benchmark

CPU: 4 Cores ARM64 Oracle Cloud

Average response time: 600μs when 30K torrents and 27K peers are online.

QPS: 700~ (can be higher but we don't have such many peers connect to our tracker)

Memory Cost: 348MB.

![image](https://github.com/user-attachments/assets/746babae-1eb3-4944-afb4-f629f78a007d)
