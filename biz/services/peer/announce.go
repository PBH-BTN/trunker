package peer

import (
	"context"

	"github.com/PBH-BTN/trunker/biz/config"
	"github.com/PBH-BTN/trunker/biz/model"
	"github.com/PBH-BTN/trunker/biz/services/peer/common"
	muxlocal "github.com/PBH-BTN/trunker/biz/services/peer/mux_local"
)

type PeerManager interface {
	// HandleAnnouncePeer 处理Announce请求
	HandleAnnouncePeer(ctx context.Context, req *model.AnnounceRequest) []*common.Peer
	// Scrape 处理Scrape请求
	Scrape(infoHash string) *model.ScrapeFile
	// Clean 清理不活跃Peer
	Clean()
	// LoadFromPersist 从持久化存储加载数据
	LoadFromPersist()
	// StoreToPersist 保存数据到持久化存储
	StoreToPersist()
	GetStatistic() *common.StatisticInfo
}

var manager PeerManager

func InitPeerManager() {
	manager = muxlocal.NewMuxLocalManager(config.AppConfig.Tracker.Shard)
}
func GetPeerManager() PeerManager {
	return manager
}
