package interval

import (
	"github.com/PBH-BTN/trunker/biz/services/peer"
	"github.com/bytedance/gopkg/util/logger"
)

func cleanInactivePeer() {
	peer.GetPeerManager().Clean()
}

func saveDB() {
	peer.GetPeerManager().StoreToPersist()
}

func printStatics() {
	statics := peer.GetPeerManager().GetStatistic()
	logger.Infof("[Statics] total peer: %d, total seed: %d", statics.TotalPeers, statics.TotalTorrents)
}
