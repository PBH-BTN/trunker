package database

import (
	"context"
	"encoding/hex"
	"errors"
	"net"
	"sync"
	"time"

	"github.com/PBH-BTN/trunker/biz/model"
	"github.com/PBH-BTN/trunker/biz/services/peer/common"
	"github.com/PBH-BTN/trunker/biz/services/peer/database/data"
	"github.com/PBH-BTN/trunker/service/database"
	"github.com/PBH-BTN/trunker/utils"
	"github.com/PBH-BTN/trunker/utils/conv"
	"github.com/bits-and-blooms/bloom/v3"
	"github.com/bytedance/gopkg/util/gopool"
	"github.com/cloudwego/hertz/pkg/common/hlog"
)

type DBManager struct {
	peerRepo        *data.PeerRepository
	blockListRepo   *data.BlockListRepository
	banInfoHashLock sync.RWMutex
	banPeerLock     sync.RWMutex
	banInfoHash     *bloom.BloomFilter
	banPeerId       *bloom.BloomFilter
}

func (m *DBManager) HandleAnnouncePeer(ctx context.Context, req *model.AnnounceRequest) ([]*common.Peer, error) {
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
	peer := &common.Peer{
		ID:         req.PeerID,
		IP:         net.ParseIP(req.IP),
		IPv4:       net.ParseIP(req.IPv4),
		IPv6:       net.ParseIP(req.IPv6),
		ClientIP:   req.ClientIP,
		Uploaded:   req.Uploaded,
		Left:       req.Left,
		Port:       req.Port,
		Downloaded: req.Downloaded,
		LastSeen:   time.Now(),
		Event:      common.ParsePeerEvent(req.Event),
		UserAgent:  req.UserAgent,
	}
	if peer.IPv4 != nil && peer.IPv4.To4() == nil {
		hlog.CtxWarnf(ctx, "invalid ipv4 address,actual: %s", peer.IPv4.String())
		return nil, errors.New("invalid address")
	}
	if peer.IPv6 != nil && peer.IPv6.To4() != nil {
		hlog.CtxWarnf(ctx, "invalid ipv6 address,actual: %s", peer.IPv4.String())
		return nil, errors.New("invalid address")
	}

	gopool.CtxGo(ctx, func() {
		if common.IsPeerConnectable(peer) { // only connectable peer will be saved
			dbPeer := CommonToDB(req.InfoHash, peer)
			err := m.peerRepo.SavePeer(ctx, dbPeer)
			if err != nil {
				hlog.CtxErrorf(ctx, "failed to save peer: %s", err.Error())
			}
		}
	})
	peers, err := m.peerRepo.PickPeers(ctx, req.InfoHash, req.NumWant)
	if err != nil {
		hlog.CtxErrorf(ctx, "failed to get peers: %s", err.Error())
		return nil, err
	}
	return utils.Map(peers, DBToCommon), nil
}

func (m *DBManager) Scrape(ctx context.Context, infoHashRaw string) (*model.ScrapeFile, error) {
	infoHash := hex.EncodeToString(conv.UnsafeStringToBytes(infoHashRaw))
	complete, err := m.peerRepo.GetCompleteCount(ctx, infoHash)
	if err != nil {
		hlog.CtxErrorf(ctx, "failed to get complete count: %s", err.Error())
		return nil, err
	}
	downloaded, err := m.peerRepo.GetDownloadedCount(ctx, infoHash)
	if err != nil {
		hlog.CtxErrorf(ctx, "failed to get downloaded count: %s", err.Error())
		return nil, err
	}
	incomplete, err := m.peerRepo.GetInCompleteCount(ctx, infoHash)
	if err != nil {
		hlog.CtxErrorf(ctx, "failed to get incomplete count: %s", err.Error())
		return nil, err
	}
	return &model.ScrapeFile{
		Complete:   int(complete),
		Downloaded: int(downloaded),
		Incomplete: int(incomplete),
	}, nil
}

func (m *DBManager) GetStatistic(ctx context.Context) (*common.StatisticInfo, error) {
	peersCount, err := m.peerRepo.GetPeersCount(ctx)
	if err != nil {
		hlog.CtxErrorf(ctx, "failed to get torrent count: %s", err.Error())
		return nil, err
	}
	infoHashCount, err := m.peerRepo.GetInfoHashCount(ctx)
	if err != nil {
		hlog.CtxErrorf(ctx, "failed to get info hash count: %s", err.Error())
		return nil, err
	}
	return &common.StatisticInfo{
		TotalPeers:    uint64(peersCount),
		TotalTorrents: uint64(infoHashCount),
		Shards: map[string]*common.StatisticInfo{
			"all": {
				TotalPeers:    uint64(peersCount),
				TotalTorrents: uint64(infoHashCount),
			},
		},
	}, nil

}

func NewDBManager() *DBManager {
	db := database.InitDB()
	m := &DBManager{
		peerRepo:        data.NewPeerRepository(db),
		blockListRepo:   data.NewBlockListRepository(db),
		banPeerLock:     sync.RWMutex{},
		banInfoHashLock: sync.RWMutex{},
		banInfoHash:     bloom.NewWithEstimates(uint(10000), 0.01),
		banPeerId:       bloom.NewWithEstimates(uint(10000), 0.01),
	}
	m.restoreBlockList()
	return m
}
