package mux_local

import (
	"context"
	"errors"
	"math/big"
	"runtime"
	"strconv"
	"sync"

	"github.com/PBH-BTN/trunker/biz/config"
	"github.com/PBH-BTN/trunker/biz/model"
	"github.com/PBH-BTN/trunker/biz/services/peer/common"
	"github.com/PBH-BTN/trunker/biz/services/peer/local"
	"github.com/PBH-BTN/trunker/utils/conv"
	"github.com/bits-and-blooms/bloom/v3"
	"github.com/cloudwego/hertz/pkg/common/hlog"
	"github.com/xxjwxc/gowp/workpool"
)

type MuxLocalManager struct {
	localList       []*local.Manager
	banInfoHashLock sync.RWMutex
	banPeerLock     sync.RWMutex
	banInfoHash     *bloom.BloomFilter
	banPeerId       *bloom.BloomFilter
}

func NewMuxLocalManager(num int) *MuxLocalManager {
	list := make([]*local.Manager, 0, num)
	for i := 0; i < num; i++ {
		list = append(list, local.NewLocalManger())
	}
	return &MuxLocalManager{
		localList:       list,
		banPeerLock:     sync.RWMutex{},
		banInfoHashLock: sync.RWMutex{},
		banInfoHash:     bloom.NewWithEstimates(uint(10000*num), 0.01),
		banPeerId:       bloom.NewWithEstimates(uint(10000*num*config.AppConfig.Tracker.MaxPeersPerTorrent), 0.01),
	}
}

func (m *MuxLocalManager) pickWorker(hashBytes []byte) *local.Manager {
	hashInt := new(big.Int).SetBytes(hashBytes)

	// 取模运算以获得服务索引
	return m.localList[hashInt.Uint64()%uint64(len(m.localList))]

}

func (m *MuxLocalManager) HandleAnnouncePeer(ctx context.Context, req *model.AnnounceRequest) ([]*common.Peer, error) {
	// process block list
	m.banInfoHashLock.RLock()
	banned := m.banInfoHash.Test(conv.UnsafeStringToBytes(req.InfoHash))
	m.banInfoHashLock.RUnlock()
	if banned {
		hlog.CtxInfof(ctx, "info hash %s is banned", req.InfoHash)
		return nil, errors.New("banned")
	}
	m.banPeerLock.RLock()
	banned = m.banPeerId.Test(conv.UnsafeStringToBytes(req.PeerID))
	m.banPeerLock.RUnlock()
	if banned {
		hlog.CtxInfof(ctx, "peer id %s is banned", req.PeerID)
		return nil, errors.New("banned")
	}

	worker := m.pickWorker(conv.UnsafeStringToBytes(req.InfoHash))
	return worker.HandleAnnouncePeer(ctx, req)
}

func (m *MuxLocalManager) BanInfoHash(infoHash string) {
	m.banInfoHashLock.Lock()
	m.banInfoHash.AddString(infoHash)
	m.banInfoHashLock.Unlock()
	worker := m.pickWorker(conv.UnsafeStringToBytes(infoHash))
	worker.BanInfoHash(infoHash)
}

func (m *MuxLocalManager) BanPeer(peerID string) {
	m.banPeerLock.Lock()
	m.banPeerId.AddString(peerID)
	m.banPeerLock.Unlock()
}

func (m *MuxLocalManager) ClearBanInfoHash() {
	m.banPeerLock.Lock()
	m.banPeerId.ClearAll()
	m.banPeerLock.Unlock()
}

func (m *MuxLocalManager) ClearBanPeer() {
	m.banPeerLock.Lock()
	m.banPeerId.ClearAll()
	m.banPeerLock.Unlock()
}

func (m *MuxLocalManager) Scrape(infoHash string) *model.ScrapeFile {
	worker := m.pickWorker(conv.UnsafeStringToBytes(infoHash))
	return worker.Scrape(infoHash)
}
func (m *MuxLocalManager) Clean() {
	wp := workpool.New(max(runtime.NumCPU()-1, 1))
	for i, manager := range m.localList {
		wp.Do(func() error {
			hlog.Info("clean shard ", i)
			manager.Clean()
			return nil
		})
	}
	_ = wp.Wait()
}

func (m *MuxLocalManager) GetStatistic() *common.StatisticInfo {
	peerCount := uint64(0)
	torrentCount := uint64(0)
	extra := make(map[string]any)
	for i, manager := range m.localList {
		info := manager.GetStatistic()
		peerCount += info.TotalPeers
		torrentCount += info.TotalTorrents
		extra[strconv.Itoa(i)] = info
	}
	return &common.StatisticInfo{
		TotalPeers:    peerCount,
		TotalTorrents: torrentCount,
		Shards:        extra,
	}
}

func (m *MuxLocalManager) GetPeers(infoHash string) []*common.Peer {
	worker := m.pickWorker(conv.UnsafeStringToBytes(infoHash))
	return worker.GetPeers(infoHash)
}

func (m *MuxLocalManager) DeleteInfoHash(infoHash string) {
	worker := m.pickWorker(conv.UnsafeStringToBytes(infoHash))
	worker.DeleteInfoHash(infoHash)
}
