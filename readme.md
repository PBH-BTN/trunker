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

## Admin

Trunker provides a simple admin interface to check the status of the tracker. Before using the admin interface, you need to set the `ADMIN_KEY` env.

### Authorization

Header: `Authorization: Bearer ${ADMIN_KEY}`

### Endpoints

#### Statistic
| Method | Path                   | Description                       |
|:-------|:-----------------------|:----------------------------------|
| GET    | `/admin/statistic`     | Get the statistic of the tracker. |

#### Block list
You can ban some torrent or peer by using the following endpoints.

| Method | Path                   | Description                       |
|:-------|:-----------------------|:----------------------------------|
| PUT    | `/admin/ban/info_hash` | Ban a torrent by info_hash.       |
| DELETE | `/admin/ban/info_hash` | Clear all info_hash ban list.     |
| PUT    | `/ban/peer`            | Ban a peer by peer_id.            |
| DELETE | `/ban/peer`            | Clear all peer ban list.          |

#### InfoHash
| Method | Path                               | Description                                                      |
|:-------|:-----------------------------------|:-----------------------------------------------------------------|
| GET    | `/admin/info_hash/:infoHash/peers` | Get all peers of a info_hash. **The response may be very large** |
| DELETE | `/admin/info_hash/:infoHash`       | Delete a info_hash                                               |

## Metrics
Currently trunker is serving a prometheus handler at `:9091` by default. With 2 paths:
- `/metrics` Providing metrics about http request, including latency and QPS.
- `/pprof` Providing golang runtime pprof info.

## Benchmark

CPU: 4 Cores ARM64 Oracle Cloud

Average response time: 600μs when 30K torrents and 27K peers are online.

QPS: 700~ (can be higher but we don't have such many peers connect to our tracker)

Memory Cost: 348MB.

![image](https://github.com/user-attachments/assets/746babae-1eb3-4944-afb4-f629f78a007d)
