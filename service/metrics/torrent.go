package metrics

import (
	"context"

	"github.com/PBH-BTN/trunker/biz/services/peer"
	"github.com/PBH-BTN/trunker/biz/services/peer/common"
	"github.com/prometheus/client_golang/prometheus"
)

const (
	GaugeTorrentTotal counterMetrics = "torrents_total"
	GaugeTorrent                     = "shard_torrents"
	GaugePeer         counterMetrics = "shard_peers"
	LabelShards                      = "shards"
)

var info *common.StatisticInfo

func gaugeSet(gaugeVec *prometheus.GaugeVec, value uint64, labels prometheus.Labels) error {
	gauge, err := gaugeVec.GetMetricWith(labels)
	if err != nil {
		return err
	}
	gauge.Set(float64(value))
	return nil
}

func registerGauge(registry *prometheus.Registry) map[counterMetrics]prometheus.Collector {
	m := make(map[counterMetrics]prometheus.Collector)
	m[GaugePeer] =
		prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Name: metricsPrefix + string(GaugePeer),
			Help: "Peer counts per shard",
		}, []string{LabelShards})
	m[GaugeTorrent] =
		prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Name: metricsPrefix + string(GaugeTorrent),
			Help: "Torrent counts per shard",
		}, []string{LabelShards})
	m[GaugeTorrentTotal] = prometheus.NewGaugeFunc(prometheus.GaugeOpts{
		Name: metricsPrefix + string(GaugeTorrentTotal),
		Help: "Total Torrent numbers",
	}, func() float64 {
		var err error
		info, err = peer.GetPeerManager().GetStatistic(context.Background())
		if err != nil {
			return 0
		}
		for s, v := range info.Shards {
			_ = gaugeSet(m[GaugePeer].(*prometheus.GaugeVec), v.TotalPeers, prometheus.Labels{LabelShards: s})
			_ = gaugeSet(m[GaugeTorrent].(*prometheus.GaugeVec), v.TotalTorrents, prometheus.Labels{LabelShards: s})
		}
		return float64(info.TotalTorrents)
	})

	for _, h := range m {
		registry.MustRegister(h)
	}

	return m
}
