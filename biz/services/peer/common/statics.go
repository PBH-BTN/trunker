package common

type StatisticInfo struct {
	TotalPeers    uint64                    `json:"total_peers"`
	TotalTorrents uint64                    `json:"total_torrents"`
	Shards        map[string]*StatisticInfo `json:"shards,omitempty"`
}
