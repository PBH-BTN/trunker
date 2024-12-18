package mux_local

import (
	"context"
	"math/big"
	"runtime"
	"strconv"

	"github.com/PBH-BTN/trunker/biz/model"
	"github.com/PBH-BTN/trunker/biz/services/peer/common"
	"github.com/PBH-BTN/trunker/biz/services/peer/local"
	"github.com/PBH-BTN/trunker/utils/conv"
	"github.com/bytedance/gopkg/util/logger"
	"github.com/xxjwxc/gowp/workpool"
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
	worker := m.pickWorker(conv.UnsafeStringToBytes(req.InfoHash))
	return worker.HandleAnnouncePeer(ctx, req)
}

func (m *MuxLocalManager) Scrape(infoHash string) *model.ScrapeFile {
	worker := m.pickWorker(conv.UnsafeStringToBytes(infoHash))
	return worker.Scrape(infoHash)
}
func (m *MuxLocalManager) Clean() {
	wp := workpool.New(max(runtime.NumCPU()-1, 1))
	for i, manager := range m.localList {
		wp.Do(func() error {
			logger.Info("clean shard %d", i)
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
		logger.Infof("shard %d, peer:%d, torrent:%d", i, info.TotalPeers, info.TotalTorrents)
	}
	return &common.StatisticInfo{
		TotalPeers:    peerCount,
		TotalTorrents: torrentCount,
		Extra:         extra,
	}
}
