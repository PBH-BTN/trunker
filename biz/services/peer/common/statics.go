package common

type StatisticInfo struct {
	TotalPeers    uint64         `json:"total_peers"`
	TotalTorrents uint64         `json:"total_torrents"`
	Extra         map[string]any `json:"extra,omitempty"`
}
