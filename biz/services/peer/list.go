package peer

import (
	"time"

	"github.com/PBH-BTN/trunker/biz/model"
	"github.com/PBH-BTN/trunker/utils"
	"github.com/zhangyunhao116/skipmap"
)

var infoHashMap *skipmap.OrderedMap[string, *skipmap.OrderedMap[string, *Peer]]

func init() {
	infoHashMap = skipmap.New[string, *skipmap.OrderedMap[string, *Peer]]()
}

func HandleAnnouncePeer(req *model.AnnounceRequest) []*model.Peer {
	peer := &Peer{
		Peer: model.Peer{
			ID:   req.PeerID,
			Port: req.Port,
			IP:   req.IP,
		},
		Uploaded:   req.Uploaded,
		Downloaded: req.Downloaded,
		LastSeen:   time.Now(),
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
	} else {
		// new peer!
		peerMap.Store(peer.GetKey(), peer)
	}
	return utils.Map(utils.SkipMapToSlice(peerMap), func(p *Peer) *model.Peer {
		return &p.Peer
	})
}
