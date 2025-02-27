package local

import (
	"context"
	"encoding/hex"
	"errors"
	"net"
	"strings"
	"time"

	"github.com/PBH-BTN/trunker/biz/config"
	"github.com/PBH-BTN/trunker/biz/model"
	"github.com/PBH-BTN/trunker/biz/services/peer/common"
	"github.com/PBH-BTN/trunker/biz/services/producer"
	"github.com/PBH-BTN/trunker/utils"
	"github.com/PBH-BTN/trunker/utils/conv"
	"github.com/bytedance/gopkg/util/gopool"
	json "github.com/bytedance/sonic"
	"github.com/cloudwego/hertz/pkg/common/hlog"
	hertz "github.com/cloudwego/hertz/pkg/common/utils"
	"github.com/lestrrat-go/choose"
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
}

func NewLocalManger() *Manager {
	return &Manager{
		infoHashMap: skipmap.New[string, *InfoHashRoot](),
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
			hlog.CtxDebugf(ctx, "delete peer %s from %s due to connect close", peer.ID, req.InfoHash)
			if v, ok := root.peerMap.LoadAndDelete(req.PeerID); ok {
				v.Conn = nil
			}
		}
	}
	if !ok { // first seen torrent
		if common.IsPeerConnectable(peer) {
			root.peerMap.LoadOrStore(peer.GetKey(), peer)
		}
		go producer.SendPeerEvent(ctx, req.InfoHash, peer)
		return nil, nil
	}
	// add to peer list
	gopool.CtxGo(ctx, func() {
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
			if common.IsPeerConnectable(peer) { // skip private ip
				// there is a data race, but it's impossible for concurrent access to one peer
				if peer.Type == model.PeerTypeWebtorrent && len(peer.Offers) > 0 {
					go m.sendOffers(ctx, root.infoHash, root.peerMap, peer, req.NumWant)
				}
				root.peerMap.LoadOrStore(peer.GetKey(), peer)
				go producer.SendPeerEvent(ctx, req.InfoHash, peer)
			}
		}
	})
	// get return
	resp := make([]*common.Peer, 0, min(root.peerMap.Len(), req.NumWant))
	timeoutPeer := make([]*common.Peer, 0)
	var oldestTime *time.Time
	var oldestPeer *common.Peer
	shouldEject := root.peerMap.Len() > config.AppConfig.Tracker.Memory.MaxPeersPerTorrent
	root.peerMap.Range(func(_ string, value *common.Peer) bool {
		if value.ID == peer.ID {
			return true
		}
		if time.Now().Add(time.Duration(-1*config.AppConfig.Tracker.TTL) * time.Second).After(value.LastSeen) {
			// timeout!
			timeoutPeer = append(timeoutPeer, value)
			return true
		}
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
		if len(resp) >= req.NumWant {
			return false
		}
		resp = append(resp, value)
		return true
	})
	if len(timeoutPeer) > 0 {
		gopool.CtxGo(ctx, func() {
			for _, toClean := range timeoutPeer {
				root.peerMap.Delete(toClean.GetKey())
			}
		})
	}
	if shouldEject && oldestPeer != nil {
		gopool.CtxGo(ctx, func() {
			hlog.CtxDebugf(ctx, "info hash %s eject %s:%d(%s) %s, last seen:%s", hex.EncodeToString(conv.UnsafeStringToBytes(root.infoHash)), oldestPeer.GetIP().String(), oldestPeer.Port, oldestPeer.ID, oldestPeer.UserAgent, oldestTime.Format(time.DateTime))
			root.peerMap.Delete(oldestPeer.GetKey())
		})
	}
	return resp, nil
}

func (m *Manager) Scrape(_ context.Context, infoHash string) (*model.ScrapeFile, error) {
	root, ok := m.infoHashMap.Load(infoHash)
	if !ok {
		return &model.ScrapeFile{
			Complete:   0,
			Incomplete: 0,
			Downloaded: 0,
		}, nil
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
	}, nil
}

func (m *Manager) GetStatistic(_ context.Context) *common.StatisticInfo {
	peerCount := 0
	m.infoHashMap.Range(func(_ string, value *InfoHashRoot) bool {
		peerCount += value.peerMap.Len()
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
	return utils.SkipMapToSlice(peerMap.peerMap), nil
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
	peer, ok := root.peerMap.Load(peerID)
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

func (m *Manager) sendOffers(ctx context.Context, infoHash string, peerMap *skipmap.OrderedMap[string, *common.Peer], peer *common.Peer, numWant int) {
	toClean := make([]string, 0)
	candidates := make([]*common.Peer, 0)
	peerMap.Range(func(key string, value *common.Peer) bool {
		if value.ID == peer.ID {
			return true
		}
		if value.Type == model.PeerTypeWebtorrent && value.Conn != nil {
			candidates = append(candidates, value)
		}
		return true
	})
	picker := choose.Slice(peer.Offers)
	for _, value := range choose.Slice(candidates).N(numWant) {
		o := picker.One()
		offer := hertz.H{
			"action":    "announce",
			"info_hash": conv.UnsafeBytesToString(conv.Trans9959_1ToUTF8(conv.UnsafeStringToBytes(infoHash))),
			"offer_id":  o.OfferID,
			"peer_id":   peer.ID,
			"offer":     o.Offer,
		}
		if err := value.Conn.WriteJSON(offer); err != nil {
			hlog.CtxErrorf(ctx, "write response error: %s", err.Error())
			toClean = append(toClean, value.GetKey())
			break
		}

	}
	for _, disconnectedPeer := range toClean {
		peerMap.Delete(disconnectedPeer)
	}
}
