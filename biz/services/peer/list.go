package peer

import (
	"github.com/PBH-BTN/trunker/biz/model"
	"github.com/bytedance/gopkg/util/logger"
	"github.com/zhangyunhao116/skipmap"
	"net"
	"time"
)

var infoHashMap *skipmap.OrderedMap[string, *skipmap.OrderedMap[string, *Peer]]

func init() {
	infoHashMap = skipmap.New[string, *skipmap.OrderedMap[string, *Peer]]()
}

func HandleAnnouncePeer(req *model.AnnounceRequest) []*Peer {
	peer := &Peer{
		ID:         req.PeerID,
		IP:         net.ParseIP(req.IP),
		IPv4:       net.ParseIP(req.IPv4),
		IPv6:       net.ParseIP(req.IPv6),
		ClientIP:   net.ParseIP(req.ClientIP),
		Uploaded:   req.Uploaded,
		Downloaded: req.Downloaded,
		LastSeen:   time.Now(),
		Event:      ParsePeerEvent(req.Event),
	}
	peerMap, ok := infoHashMap.Load(req.InfoHash)
	if !ok { // first seen torrent
		peerMap = skipmap.New[string, *Peer]()
		peerMap.Store(peer.GetKey(), peer)
		infoHashMap.Store(req.InfoHash, peerMap)
		return nil
	}
	if knownPeer, ok := peerMap.Load(peer.GetKey()); ok {
		// update current record
		knownPeer.Uploaded = peer.Uploaded
		knownPeer.Downloaded = peer.Downloaded
		knownPeer.LastSeen = peer.LastSeen
		knownPeer.Event = peer.Event
		defer func() {
			logger.Info("do cleaning for info hash: %s", req.InfoHash)
			go CleanUp(peerMap)
		}()
	} else {
		// new peer!
		peerMap.Store(peer.GetKey(), peer)
	}
	resp := make([]*Peer, 0, peerMap.Len())
	peerMap.Range(func(_ string, value *Peer) bool {
		if value.Port == 0 { // not accept incoming connections
			return true
		}
		if len(resp) >= req.NumWant {
			return false
		}
		resp = append(resp, value)
		return true
	})
	return resp
}
