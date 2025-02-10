package peer

import (
	"context"

	"github.com/PBH-BTN/trunker/biz/config"
	"github.com/PBH-BTN/trunker/biz/model"
	"github.com/PBH-BTN/trunker/biz/services/peer/common"
	"github.com/PBH-BTN/trunker/biz/services/peer/database"
	muxlocal "github.com/PBH-BTN/trunker/biz/services/peer/mux_local"
)

type PeerManager interface {
	// HandleAnnouncePeer 处理Announce请求
	HandleAnnouncePeer(ctx context.Context, req *model.AnnounceRequest) ([]*common.Peer, error)
	// Scrape 处理Scrape请求
	Scrape(ctx context.Context, infoHash string) (*model.ScrapeFile, error)
	// Clean 清理不活跃Peer
	Clean()
	// LoadFromPersist 从持久化存储加载数据
	LoadFromPersist()
	// StoreToPersist 保存数据到持久化存储
	StoreToPersist()

	/*	admin interface	*/
	GetStatistic(ctx context.Context) (*common.StatisticInfo, error)
	BanInfoHash(ctx context.Context, infoHash string) error
	BanPeer(ctx context.Context, peerID string) error
	ClearBanInfoHash()
	ClearBanPeer()
	GetPeers(ctx context.Context, infoHash string) ([]*common.Peer, error)
	DeleteInfoHash(ctx context.Context, infoHash string) error
}

var manager PeerManager

func InitPeerManager() {
	switch config.AppConfig.Tracker.Mode {
	case config.RunningModeMemory:
		manager = muxlocal.NewMuxLocalManager(config.AppConfig.Tracker.Memory.Shard)
		manager.LoadFromPersist()
	case config.RunningModeDB:
		manager = database.NewDBManager()
	default:
		panic("unknown tracker mode")
	}
}
func GetPeerManager() PeerManager {
	return manager
}
