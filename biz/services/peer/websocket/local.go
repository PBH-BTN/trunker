package websocket

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
	"github.com/PBH-BTN/trunker/service/cache"
	"github.com/PBH-BTN/trunker/utils"
	"github.com/PBH-BTN/trunker/utils/collections/mapx"
	"github.com/PBH-BTN/trunker/utils/conv"
	"github.com/bytedance/gopkg/util/gopool"
	json "github.com/bytedance/sonic"
	"github.com/cloudwego/hertz/pkg/common/hlog"
	"github.com/xxjwxc/gowp/workpool"
)

type manager struct {
	infoHashMap mapx.SyncStringMap[*infoHashRoot]
}

func newManager() *manager {
	return &manager{
		infoHashMap: mapx.NewSkipMap[*infoHashRoot](),
	}
}

type infoHashRoot struct {
	infoHash string
	peerMap  mapx.SyncStringMap[*common.Peer]
}

func (m *manager) HandleAnnouncePeer(ctx context.Context, req *model.AnnounceRequest) ([]*common.Peer, error) {
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
	hlog.CtxDebugf(ctx, "handle peer announce %s from %s", peer.ID, hex.EncodeToString(conv.UnsafeStringToBytes(req.InfoHash)))

	root, ok := m.infoHashMap.LoadOrStoreLazy(req.InfoHash, func() *infoHashRoot {
		return &infoHashRoot{
			infoHash: req.InfoHash,
			peerMap:  mapx.NewSkipMap[*common.Peer](),
		}
	})
	peer.Conn.CloseCallback = func() {
		hlog.CtxDebugf(ctx, "delete peer %s from %s due to connect close, ip: %s", conv.Trans8859_1ToUTF8([]byte(peer.ID)), hex.EncodeToString(conv.UnsafeStringToBytes(req.InfoHash)), peer.GetIP().String())
		if v, ok := root.peerMap.LoadAndDelete(req.PeerID); ok {
			v.Conn = nil
		}
	}
	if !ok { // first seen torrent
		root.peerMap.LoadOrStore(peer.GetKey(), peer)
		go producer.SendPeerEvent(ctx, req.InfoHash, peer)
		return nil, nil
	}
	if peer.Event == common.PeerEvent_Stopped { // stopped peer must remove and return nothing
		_ = peer.Conn.Close()
		root.peerMap.Delete(peer.GetKey())
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
			root.peerMap.LoadOrStore(peer.GetKey(), peer)
		}
	})
	// get return
	resp := make([]*common.Peer, 0, utils.Positive(min(root.peerMap.Len(), req.NumWant)))
	root.peerMap.Range(func(_ string, value *common.Peer) bool {
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

	return resp, nil
}

func (m *manager) Scrape(ctx context.Context, infoHash string) (*model.ScrapeFile, error) {
	if config.AppConfig.Cache.Enable {
		if v, ok := cache.Get[model.ScrapeFile](ctx, "scrape_ws_"+infoHash); ok {
			return v, nil
		}
	}
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
	root.peerMap.Range(func(_ string, value *common.Peer) bool {
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

	_ = p.Wait()
	ret := &model.ScrapeFile{
		Seeder:     int(seeder.Load()),
		Complete:   int(complete.Load()),
		Incomplete: int(incomplete.Load()),
		Downloaded: int(downloaded.Load()), // 这个目前不实现
	}
	if config.AppConfig.Cache.Enable {
		_ = cache.Set(ctx, "scrape_ws_"+infoHash, ret, time.Minute*5)
	}
	return ret, nil
}

func (m *manager) AnswerToPeer(ctx context.Context, infoHash string, peerID string, answerBody []byte) error {
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
