package interval

import (
	"context"

	"github.com/PBH-BTN/trunker/biz/services/peer"
	"github.com/bytedance/gopkg/util/logger"
)

func cleanInactivePeer() {
	count := peer.GetPeerManager().Clean()
	logger.Infof("[Clean] clean inactive peers: %d", count)
}

func saveDB() {
	peer.GetPeerManager().StoreToPersist()
}

func printStatics() {
	statics, err := peer.GetPeerManager().GetStatistic(context.Background())
	if err != nil {
		logger.Errorf("[Statics] get statistic error: %v", err)
		return
	}
	logger.Infof("[Statics] total peer: %d, total seed: %d", statics.TotalPeers, statics.TotalTorrents)
}
