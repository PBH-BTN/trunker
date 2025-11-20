package mux_local

import (
	"context"
	"errors"
	"math/big"
	"runtime"
	"strconv"
	"sync"
	"sync/atomic"

	"github.com/PBH-BTN/trunker/biz/config"
	"github.com/PBH-BTN/trunker/biz/model"
	"github.com/PBH-BTN/trunker/biz/services/peer/common"
	"github.com/PBH-BTN/trunker/biz/services/peer/local"
	"github.com/PBH-BTN/trunker/biz/services/peer/mux_local/ban"
	"github.com/PBH-BTN/trunker/utils/conv"
	"github.com/cloudwego/hertz/pkg/common/hlog"
	"github.com/xxjwxc/gowp/workpool"
)

type MuxLocalManager struct {
	localList []*local.Manager
	ban       *ban.Manager
}

func NewMuxLocalManager(num int) *MuxLocalManager {
	hlog.Info("running as memory mode, shard: ", num)
	list := make([]*local.Manager, 0, num)
	for i := 0; i < num; i++ {
		list = append(list, local.NewLocalManger())
	}

	// Initialize ban manager with XDP support if configured
	var banManager *ban.Manager
	if config.AppConfig.Tracker.XDP.Enable && config.AppConfig.Tracker.XDP.Interface != "" {
		hlog.Infof("Initializing ban manager with XDP on interface: %s", config.AppConfig.Tracker.XDP.Interface)
		banManager = ban.NewBanManagerWithIP(config.AppConfig.Tracker.XDP.Interface)
	} else {
		banManager = ban.NewBanManager()
	}

	return &MuxLocalManager{
		localList: list,
		ban:       banManager,
	}
}

func (m *MuxLocalManager) pickWorker(hashBytes []byte) *local.Manager {
	hashInt := new(big.Int).SetBytes(hashBytes)

	// 取模运算以获得服务索引
	return m.localList[hashInt.Uint64()%uint64(len(m.localList))]

}

func (m *MuxLocalManager) BanInfoHash(ctx context.Context, infoHash string) error {
	err := m.ban.AddBan(ban.BanTypeInfoHash, infoHash)
	if err != nil {
		return err
	}
	worker := m.pickWorker(conv.UnsafeStringToBytes(infoHash))
	return worker.BanInfoHash(ctx, infoHash)
}

func (m *MuxLocalManager) BanPeer(_ context.Context, peerID string) error {
	return m.ban.AddBan(ban.BanTypePeerId, peerID)
}

func (m *MuxLocalManager) ClearBanInfoHash() {
	m.ban.Clear(ban.BanTypeInfoHash)
}

func (m *MuxLocalManager) ClearBanPeer() {
	m.ban.Clear(ban.BanTypePeerId)
}

func (m *MuxLocalManager) HandleAnnouncePeer(ctx context.Context, req *model.AnnounceRequest) ([]*common.Peer, error) {
	// process block list
	if m.ban.Test(ban.BanTypeInfoHash, req.InfoHash) {
		hlog.CtxInfof(ctx, "info hash %s is banned", req.InfoHash)
		return nil, errors.New("banned info_hash")
	}
	if m.ban.Test(ban.BanTypePeerId, req.PeerID) {
		hlog.CtxInfof(ctx, "peer id %s is banned", req.PeerID)
		return nil, errors.New("banned peer_id")
	}

	// Check IP ban (application layer filtering)
	// This ensures IP filtering works even when XDP is unavailable or falls back
	if req.ClientIP != nil && m.ban.TestIPAddr(req.ClientIP) {
		hlog.CtxInfof(ctx, "client IP %s is banned", req.ClientIP.String())
		return nil, errors.New("banned ip")
	}

	worker := m.pickWorker(conv.UnsafeStringToBytes(req.InfoHash))
	return worker.HandleAnnouncePeer(ctx, req)
}

func (m *MuxLocalManager) Scrape(ctx context.Context, infoHash string) (*model.ScrapeFile, error) {
	worker := m.pickWorker(conv.UnsafeStringToBytes(infoHash))
	return worker.Scrape(ctx, infoHash)
}
func (m *MuxLocalManager) Clean() int64 {
	wp := workpool.New(max(runtime.NumCPU()-1, 1))
	total := atomic.Int64{}
	for _, manager := range m.localList {
		wp.Do(func() error {
			total.Add(manager.Clean())
			return nil
		})
	}
	_ = wp.Wait()
	return total.Load()
}

func (m *MuxLocalManager) GetStatistic(ctx context.Context) (*common.StatisticInfo, error) {
	peerCount := uint64(0)
	torrentCount := uint64(0)
	extra := make(map[string]*common.StatisticInfo)
	mu := sync.Mutex{}
	wp := workpool.New(max(runtime.NumCPU()-1, 1))
	for i, manager := range m.localList {
		wp.Do(func() error {
			info := manager.GetStatistic(ctx)
			mu.Lock()
			peerCount += info.TotalPeers
			torrentCount += info.TotalTorrents
			extra[strconv.Itoa(i)] = info
			mu.Unlock()
			return nil
		})
	}
	_ = wp.Wait()
	return &common.StatisticInfo{
		TotalPeers:    peerCount,
		TotalTorrents: torrentCount,
		Shards:        extra,
	}, nil
}

func (m *MuxLocalManager) GetPeers(ctx context.Context, infoHash string) ([]*common.Peer, error) {
	worker := m.pickWorker(conv.UnsafeStringToBytes(infoHash))
	return worker.GetPeers(ctx, infoHash)
}

func (m *MuxLocalManager) DeleteInfoHash(ctx context.Context, infoHash string) error {
	worker := m.pickWorker(conv.UnsafeStringToBytes(infoHash))
	return worker.DeleteInfoHash(ctx, infoHash)
}

// IP Ban Management

func (m *MuxLocalManager) BanIP(_ context.Context, ip string) error {
	return m.ban.AddIPBan(ip)
}

func (m *MuxLocalManager) UnbanIP(_ context.Context, ip string) error {
	return m.ban.RemoveIPBan(ip)
}

func (m *MuxLocalManager) ClearBanIP() error {
	return m.ban.ClearIPBans()
}

func (m *MuxLocalManager) GetIPFilterStats() *common.IPFilterStats {
	stats := m.ban.GetIPFilterStats()
	if stats == nil {
		return nil
	}
	return &common.IPFilterStats{
		TotalIPs:       stats.TotalIPs,
		PacketsDropped: stats.PacketsDropped,
		PacketsAllowed: stats.PacketsAllowed,
		IsXDP:          stats.IsXDP,
		InterfaceName:  stats.InterfaceName,
	}
}
