package local

import (
	"context"
	"encoding/hex"
	"net"
	"sync/atomic"
	"time"

	"github.com/PBH-BTN/trunker/biz/config"
	"github.com/PBH-BTN/trunker/biz/model"
	"github.com/PBH-BTN/trunker/biz/services/peer/common"
	"github.com/PBH-BTN/trunker/utils/conv"
	"github.com/cloudwego/hertz/pkg/common/hlog"
	"github.com/zhangyunhao116/skipmap"
)

type infoHashRoot struct {
	peerMap   *skipmap.OrderedMap[string, *common.Peer]
	lastClean time.Time
	infoHash  string
}

type Manager struct {
	infoHashMap *skipmap.OrderedMap[string, *infoHashRoot]
	peerCount   atomic.Int64
}

func NewLocalManger() *Manager {
	return &Manager{
		infoHashMap: skipmap.New[string, *infoHashRoot](),
		peerCount:   atomic.Int64{},
	}
}

func (m *Manager) HandleAnnouncePeer(ctx context.Context, req *model.AnnounceRequest) []*common.Peer {
	peer := &common.Peer{
		ID:         req.PeerID,
		IP:         net.ParseIP(req.IP),
		IPv4:       net.ParseIP(req.IPv4),
		IPv6:       net.ParseIP(req.IPv6),
		ClientIP:   net.ParseIP(req.ClientIP),
		Uploaded:   req.Uploaded,
		Left:       req.Left,
		Port:       req.Port,
		Downloaded: req.Downloaded,
		LastSeen:   time.Now(),
		Event:      common.ParsePeerEvent(req.Event),
		UserAgent:  req.UserAgent,
	}
	root, ok := m.infoHashMap.Load(req.InfoHash)
	if !ok { // first seen torrent
		root = &infoHashRoot{
			peerMap:   skipmap.New[string, *common.Peer](),
			lastClean: time.Now(),
			infoHash:  req.InfoHash,
		}
		root.peerMap.Store(peer.GetKey(), peer)
		m.infoHashMap.Store(req.InfoHash, root)
		m.peerCount.Add(1)
		return nil
	}
	shouldEject := false
	if knownPeer, ok := root.peerMap.Load(peer.GetKey()); ok {
		// update current record
		knownPeer.Uploaded = peer.Uploaded
		knownPeer.Downloaded = peer.Downloaded
		knownPeer.LastSeen = peer.LastSeen
		knownPeer.Event = peer.Event
		knownPeer.Left = peer.Left
	} else {
		// new peer!
		if !(peer.GetIP().IsPrivate() || peer.GetIP().IsLoopback() || peer.Port == 0) { // skip private ip
			root.peerMap.Store(peer.GetKey(), peer)
			if root.peerMap.Len() > config.AppConfig.Tracker.MaxPeersPerTorrent {
				shouldEject = true // too much peers, need eject
			}
			m.peerCount.Add(1)
		}
	}
	resp := make([]*common.Peer, 0, min(root.peerMap.Len(), req.NumWant))
	var oldestTime *time.Time
	var oldestPeer *common.Peer
	root.peerMap.Range(func(_ string, value *common.Peer) bool {
		if len(resp) >= req.NumWant {
			return false
		}
		if shouldEject {
			if oldestTime == nil {
				oldestTime = &value.LastSeen
				oldestPeer = value
			} else {
				if value.LastSeen.Before(*oldestTime) {
					oldestTime = &value.LastSeen
					oldestPeer = value
				}
			}
		}
		resp = append(resp, value)
		return true
	})
	if shouldEject {
		hlog.CtxDebugf(ctx, "info hash %s eject %s:%d(%s) %s, last seen:%s", hex.EncodeToString(conv.UnsafeStringToBytes(root.infoHash)), oldestPeer.GetIP().String(), oldestPeer.Port, oldestPeer.ID, oldestPeer.UserAgent, oldestTime.Format(time.DateTime))
		root.peerMap.Delete(oldestPeer.GetKey())
		m.peerCount.Add(-1)
	}
	return resp
}

func (m *Manager) Scrape(infoHash string) *model.ScrapeFile {
	root, ok := m.infoHashMap.Load(infoHash)
	if !ok {
		return &model.ScrapeFile{
			Complete:   0,
			Incomplete: 0,
			Downloaded: 0,
		}
	}
	var complete, incomplete, downloaded int
	root.peerMap.Range(func(_ string, value *common.Peer) bool {
		if value.Event == common.PeerEvent_Completed {
			complete++
			return true
		} else if value.Event == common.PeerEvent_Started {
			incomplete++
			return true
		} else if value.Left == 0 {
			complete++
			return true
		}
		return true
	})
	return &model.ScrapeFile{
		Complete:   complete,
		Incomplete: incomplete,
		Downloaded: downloaded, // 这个目前不实现
	}
}

func (m *Manager) GetStatistic() *common.StatisticInfo {
	return &common.StatisticInfo{
		TotalTorrents: uint64(m.infoHashMap.Len()),
		TotalPeers:    uint64(m.peerCount.Load()),
	}
}
