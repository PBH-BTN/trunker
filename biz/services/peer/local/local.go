package local

import (
	"context"
	"encoding/hex"
	"errors"
	"net"
	"runtime"
	"strings"
	"sync/atomic"
	"time"

	"github.com/PBH-BTN/trunker/biz/config"
	"github.com/PBH-BTN/trunker/biz/model"
	"github.com/PBH-BTN/trunker/biz/services/peer/common"
	"github.com/PBH-BTN/trunker/biz/services/producer"
	"github.com/PBH-BTN/trunker/utils"
	"github.com/PBH-BTN/trunker/utils/collections/mapx"
	"github.com/PBH-BTN/trunker/utils/conv"
	"github.com/bytedance/gopkg/util/gopool"
	json "github.com/bytedance/sonic"
	"github.com/cloudwego/hertz/pkg/common/hlog"
	"github.com/xxjwxc/gowp/workpool"
)

type InfoHashRoot struct {
	peerMap       [3]mapx.SyncStringMap[*common.Peer] // always keep 3 map, one for write, one for readonly and one keeps empty
	lastClean     time.Time
	infoHash      string
	currentActive uint32
}

func NewInfoHashRoot(infoHash string) *InfoHashRoot {
	return &InfoHashRoot{
		currentActive: 0,
		peerMap: [3]mapx.SyncStringMap[*common.Peer]{
			mapx.NewSkipMap[*common.Peer](),
			mapx.NewSkipMap[*common.Peer](),
			mapx.NewSkipMap[*common.Peer](),
		},
		lastClean: time.Now(),
		infoHash:  infoHash,
	}
}

func (i *InfoHashRoot) Load(key string) (*common.Peer, bool) {
	for _, peerMap := range i.peerMap {
		if v, ok := peerMap.Load(key); ok {
			return v, ok
		}
	}
	return nil, false
}

func (i *InfoHashRoot) LoadAndDelete(key string) (*common.Peer, bool) {
	for _, peerMap := range i.peerMap {
		if v, ok := peerMap.LoadAndDelete(key); ok {
			return v, ok
		}
	}
	return nil, false
}

func (i *InfoHashRoot) LoadOrStore(key string, peer *common.Peer) (*common.Peer, bool) {
	return i.peerMap[i.currentActive].LoadOrStore(key, peer)
}

func (i *InfoHashRoot) Store(key string, peer *common.Peer) {
	i.peerMap[i.currentActive].Store(key, peer)
}

func (i *InfoHashRoot) Len() int {
	count := 0
	for _, s := range i.peerMap {
		count += s.Len()
	}
	return count
}

type Manager struct {
	infoHashMap mapx.SyncStringMap[*InfoHashRoot]
}

func NewLocalManger() *Manager {
	return &Manager{
		infoHashMap: mapx.NewSkipMap[*InfoHashRoot](),
	}
}

func (m *Manager) HandleAnnouncePeer(ctx context.Context, req *model.AnnounceRequest) ([]*common.Peer, error) {
	peer := &common.Peer{
		ID:         req.PeerID,
		IP:         net.ParseIP(req.IP),
		IPv4:       net.ParseIP(req.IPv4),
		IPv6:       net.ParseIP(req.IPv6),
		ClientIP:   req.ClientIP,
		Uploaded:   req.Uploaded,
		Left:       req.Left,
		Port:       req.Port,
		Type:       req.Type,
		Downloaded: req.Downloaded,
		Offers:     req.Offers,
		LastSeen:   time.Now(),
		Event:      common.ParsePeerEvent(req.Event),
		UserAgent:  req.UserAgent,
		Conn:       req.Conn,
		Source:     req.Source,
	}
	if peer.IPv4 != nil && peer.IPv4.To4() == nil {
		hlog.CtxWarnf(ctx, "invalid ipv4 address,actual: %s", peer.IPv4.String())
		return nil, errors.New("invalid address")
	}
	if peer.IPv6 != nil && peer.IPv6.To4() != nil {
		hlog.CtxWarnf(ctx, "invalid ipv6 address,actual: %s", peer.IPv6.String())
		return nil, errors.New("invalid address")
	}

	root, ok := m.infoHashMap.LoadOrStoreLazy(req.InfoHash, func() *InfoHashRoot {
		return NewInfoHashRoot(req.InfoHash)
	})
	if peer.Type == model.PeerTypeWebtorrent && req.Conn != nil {
		peer.Conn.CloseCallback = func() {
			hlog.CtxDebugf(ctx, "delete peer %s from %s due to connect close", peer.ID, hex.EncodeToString(conv.UnsafeStringToBytes(req.InfoHash)))
			if v, ok := root.LoadAndDelete(req.PeerID); ok {
				v.Conn = nil
			}
		}
	}
	if !ok { // first seen torrent
		if common.IsPeerConnectable(peer) {
			root.LoadOrStore(peer.GetKey(), peer)
		}
		go producer.SendPeerEvent(ctx, req.InfoHash, peer)
		return nil, nil
	}
	if peer.Event == common.PeerEvent_Stopped { // stopped peer must remove and return nothing
		root.LoadAndDelete(peer.GetKey())
		return nil, nil
	}
	// add to peer list
	gopool.CtxGo(ctx, func() {
		if knownPeer, ok := root.LoadAndDelete(peer.GetKey()); ok {
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
			root.Store(knownPeer.GetKey(), knownPeer)
		} else {
			// new peer!
			if common.IsPeerConnectable(peer) { // skip private ip
				root.LoadOrStore(peer.GetKey(), peer)
				go producer.SendPeerEvent(ctx, req.InfoHash, peer)
			}
		}
	})
	// get return
	resp := make([]*common.Peer, 0, utils.Positive(min(root.Len(), req.NumWant)))
	root.Range(func(_ string, value *common.Peer) bool {
		if value.Type != peer.Type { // same type peer only
			return true
		}
		if value.Type == model.PeerTypeWebtorrent {
			if value.Conn == nil {
				return true
			}
		}
		if value.Event == common.PeerEvent_Stopped { // stopped peer should not return
			return true
		}

		if len(resp) >= req.NumWant {
			return false
		}
		if value.ID == peer.ID {
			return true
		}
		resp = append(resp, value)
		return true
	})
	if root.peerMap[root.currentActive].Len() > config.AppConfig.Tracker.Memory.MaxPeersPerTorrent/2 { // reach max, start to eject
		hlog.CtxDebugf(ctx, "[info_hash %s]active set full, currentActive %d, size: 1: %d 2:%d 3:%d", hex.EncodeToString(conv.UnsafeStringToBytes(req.InfoHash)), root.currentActive, root.peerMap[0].Len(), root.peerMap[1].Len(), root.peerMap[2].Len())
		current := root.currentActive
		if atomic.CompareAndSwapUint32(&root.currentActive, current, (current+1)%3) { // write head switch to next
			hlog.CtxDebugf(ctx, "[info_hash %s] active set swapped! current:%d", hex.EncodeToString(conv.UnsafeStringToBytes(req.InfoHash)), root.currentActive)
			// empty the oldest map
			hlog.CtxDebugf(ctx, "[info_hash %s] clean oldest set %d, len:%d", hex.EncodeToString(conv.UnsafeStringToBytes(req.InfoHash)), (current+2)%3, root.peerMap[(current+2)%3].Len())
			root.peerMap[(current+2)%3] = mapx.NewSkipMap[*common.Peer]()
			go runtime.GC()
		}
	}

	return resp, nil
}

func (m *Manager) Scrape(ctx context.Context, infoHash string) (*model.ScrapeFile, error) {
	root, ok := m.infoHashMap.Load(infoHash)
	if !ok {
		return &model.ScrapeFile{
			Complete:   0,
			Incomplete: 0,
			Downloaded: 0,
			Seeder:     0,
		}, nil
	}
	var complete, incomplete, downloaded, seeder atomic.Int64
	p := workpool.New(runtime.NumCPU() * 2)
	for _, s := range root.peerMap {
		s.Range(func(_ string, value *common.Peer) bool {
			p.Do(func() error {
				if value.Left == 0 {
					downloaded.Add(1)
					complete.Add(1)
					if value.Event != common.PeerEvent_Stopped {
						seeder.Add(1)
					}
					return nil
				}
				if value.Event == common.PeerEvent_Completed {
					complete.Add(1)
				} else {
					incomplete.Add(1)
				}
				return nil
			})
			return true
		})
	}

	_ = p.Wait()
	return &model.ScrapeFile{
		Seeder:     int(seeder.Load()),
		Complete:   int(complete.Load()),
		Incomplete: int(incomplete.Load()),
		Downloaded: int(downloaded.Load()), // 这个目前不实现
	}, nil
}

func (m *Manager) GetStatistic(_ context.Context) *common.StatisticInfo {
	peerCount := 0
	m.infoHashMap.Range(func(_ string, value *InfoHashRoot) bool {
		peerCount += value.Len()
		return true
	})
	return &common.StatisticInfo{
		TotalTorrents: uint64(m.infoHashMap.Len()),
		TotalPeers:    uint64(peerCount),
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
	root.Store(peer.GetKey(), peer)
}

func (i *InfoHashRoot) Range(f func(key string, value *common.Peer) bool) {
	for _, s := range i.peerMap {
		s.Range(f)
	}
}

func (m *Manager) StoreToPersist() {
	panic("please use mux to persist")
}

func (m *Manager) LoadFromPersist() {
	panic("please use mux to persist")
}

func (m *Manager) BanInfoHash(_ context.Context, infoHash string) error {
	// ban process in the mux, we just delete at here
	m.infoHashMap.Delete(infoHash)
	return nil
}

func (m *Manager) BanPeer(_ context.Context, _ string) error {
	// do nothing, let its ttl end
	return nil
}

func (m *Manager) ClearBanInfoHash() {
	// do nothing
}

func (m *Manager) ClearBanPeer() {
	// do nothing
}

func (m *Manager) GetPeers(_ context.Context, infoHash string) ([]*common.Peer, error) {
	peerMap, ok := m.infoHashMap.Load(infoHash)
	if !ok {
		return []*common.Peer{}, nil
	}
	return utils.SkipMapToSlice(peerMap), nil
}

func (m *Manager) DeleteInfoHash(_ context.Context, infoHash string) error {
	m.infoHashMap.Delete(infoHash)
	return nil
}

func (m *Manager) AnswerToPeer(ctx context.Context, infoHash string, peerID string, answerBody []byte) error {
	root, ok := m.infoHashMap.Load(infoHash)
	if !ok {
		return errors.New("info hash not found")
	}
	peer, ok := root.Load(peerID)
	if !ok {
		return errors.New("peer not found")
	}
	if peer.Conn == nil {
		return errors.New("peer not connected")
	}
	resp := map[string]any{}
	_ = json.Unmarshal(answerBody, &resp)
	delete(resp, "to_peer_id")
	err := peer.Conn.WriteJSON(resp)
	if err != nil {
		if strings.Contains(err.Error(), "close") {
			peer.Conn = nil
			return errors.New("remote peer offline")
		}
		return err
	}
	return nil
}
