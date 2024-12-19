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
	"github.com/PBH-BTN/trunker/biz/services/producer"
	"github.com/PBH-BTN/trunker/utils/conv"
	"github.com/cloudwego/hertz/pkg/common/hlog"
	"github.com/zhangyunhao116/skipmap"
)

type InfoHashRoot struct {
	peerMap   *skipmap.OrderedMap[string, *common.Peer]
	lastClean time.Time
	infoHash  string
}

func NewInfoHashRoot(infoHash string) *InfoHashRoot {
	return &InfoHashRoot{
		peerMap:   skipmap.New[string, *common.Peer](),
		lastClean: time.Now(),
		infoHash:  infoHash,
	}
}

type Manager struct {
	infoHashMap *skipmap.OrderedMap[string, *InfoHashRoot]
	peerCount   atomic.Int64
}

func NewLocalManger() *Manager {
	return &Manager{
		infoHashMap: skipmap.New[string, *InfoHashRoot](),
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
	root, ok := m.infoHashMap.LoadOrStoreLazy(req.InfoHash, func() *InfoHashRoot {
		return NewInfoHashRoot(req.InfoHash)
	})
	if !ok { // first seen torrent
		root.peerMap.Store(peer.GetKey(), peer)
		m.peerCount.Add(1)
		go producer.SendPeerEvent(ctx, req.InfoHash, peer)
		return nil
	}
	shouldEject := false
	if knownPeer, ok := root.peerMap.Load(peer.GetKey()); ok {
		// update current record
		if (knownPeer.Left != 0 && peer.Left == 0) || knownPeer.Event != peer.Event {
			go producer.SendPeerEvent(ctx, req.InfoHash, peer)
		}
		knownPeer.Uploaded = peer.Uploaded
		knownPeer.Downloaded = peer.Downloaded
		knownPeer.LastSeen = peer.LastSeen
		knownPeer.Event = peer.Event
		knownPeer.Left = peer.Left
		knownPeer.Event = peer.Event
	} else {
		// new peer!
		if !(peer.GetIP().IsPrivate() || peer.GetIP().IsLoopback() || peer.Port == 0) { // skip private ip
			// there is a data race, but it's impossible for concurrent access to one peer
			root.peerMap.Store(peer.GetKey(), peer)
			if root.peerMap.Len() > config.AppConfig.Tracker.MaxPeersPerTorrent {
				shouldEject = true // too much peers, need eject
			}
			m.peerCount.Add(1)
			go producer.SendPeerEvent(ctx, req.InfoHash, peer)
		}
	}
	resp := make([]*common.Peer, 0, min(root.peerMap.Len(), req.NumWant))
	timeoutPeer := make([]*common.Peer, 0)
	var oldestTime *time.Time
	var oldestPeer *common.Peer
	root.peerMap.Range(func(_ string, value *common.Peer) bool {
		if time.Now().Add(time.Duration(-1*config.AppConfig.Tracker.TTL) * time.Second).After(value.LastSeen) {
			// timeout!
			timeoutPeer = append(timeoutPeer, value)
			return true
		}
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
	if len(timeoutPeer) > 0 {
		for _, toClean := range timeoutPeer {
			_, ok := root.peerMap.LoadAndDelete(toClean.GetKey())
			if ok {
				m.peerCount.Add(-1)
			}
		}
	}
	if shouldEject {
		hlog.CtxDebugf(ctx, "info hash %s eject %s:%d(%s) %s, last seen:%s", hex.EncodeToString(conv.UnsafeStringToBytes(root.infoHash)), oldestPeer.GetIP().String(), oldestPeer.Port, oldestPeer.ID, oldestPeer.UserAgent, oldestTime.Format(time.DateTime))
		_, ok := root.peerMap.LoadAndDelete(oldestPeer.GetKey())
		if ok {
			m.peerCount.Add(-1)
		}
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
		if value.Left == 0 {
			complete++
		} else {
			incomplete++
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

func (m *Manager) RangeMap(f func(key string, value *InfoHashRoot) bool) {
	m.infoHashMap.Range(f)
}

// DirectStore Store directly, no check, unsafe
func (m *Manager) DirectStore(infoHash string, peer *common.Peer) {
	root, _ := m.infoHashMap.LoadOrStoreLazy(infoHash, func() *InfoHashRoot {
		return NewInfoHashRoot(infoHash)
	})
	root.peerMap.Store(peer.GetKey(), peer)
}

func (i *InfoHashRoot) Range(f func(key string, value *common.Peer) bool) {
	i.peerMap.Range(f)
}

func (m *Manager) StoreToPersist() {
	panic("please use mux to persist")
}

func (m *Manager) LoadFromPersist() {
	panic("please use mux to persist")
}
