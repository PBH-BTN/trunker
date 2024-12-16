package muxlocal

import (
	"context"
	"math/big"
	"strconv"

	"github.com/PBH-BTN/trunker/biz/model"
	"github.com/PBH-BTN/trunker/biz/services/peer/common"
	"github.com/PBH-BTN/trunker/biz/services/peer/local"
	"github.com/bytedance/gopkg/util/logger"
)

type MuxLocalManager struct {
	localList []*local.Manager
}

func NewMuxLocalManager(num int) *MuxLocalManager {
	list := make([]*local.Manager, 0, num)
	for i := 0; i < num; i++ {
		list = append(list, local.NewLocalManger())
	}
	return &MuxLocalManager{
		localList: list,
	}
}

func (m *MuxLocalManager) pickWorker(hashBytes []byte) *local.Manager {
	hashInt := new(big.Int).SetBytes(hashBytes)

	// 取模运算以获得服务索引
	return m.localList[hashInt.Uint64()%uint64(len(m.localList))]

}

func (m *MuxLocalManager) HandleAnnouncePeer(ctx context.Context, req *model.AnnounceRequest) []*common.Peer {
	worker := m.pickWorker([]byte(req.InfoHash))
	return worker.HandleAnnouncePeer(ctx, req)
}

func (m *MuxLocalManager) Scrape(infoHash string) *model.ScrapeFile {
	worker := m.pickWorker([]byte(infoHash))
	return worker.Scrape(infoHash)
}
func (m *MuxLocalManager) Clean() {
	for i, manager := range m.localList {
		logger.Info("clean shard %d", i)
		go manager.Clean()
	}
}

func (m *MuxLocalManager) LoadFromPersist() {
	// todo
}

func (m *MuxLocalManager) StoreToPersist() {
	// todo
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
		logger.Infof("shard %d, peer:%d, torrent:%d", i, info.TotalPeers, info.TotalTorrents)
	}
	return &common.StatisticInfo{
		TotalPeers:    peerCount,
		TotalTorrents: torrentCount,
		Extra:         extra,
	}
}
