package peer

import (
	"net"
	"time"

	"github.com/PBH-BTN/trunker/biz/model"
	"github.com/bytedance/gopkg/util/logger"
	"github.com/zhangyunhao116/skipmap"
)

type InfoHashRoot struct {
	peerMap   *skipmap.OrderedMap[string, *Peer]
	lastClean time.Time
}

var infoHashMap *skipmap.OrderedMap[string, *InfoHashRoot]

func init() {
	infoHashMap = skipmap.New[string, *InfoHashRoot]()
}

func GetAllMap() *skipmap.OrderedMap[string, *InfoHashRoot] {
	return infoHashMap
}

func HandleAnnouncePeer(req *model.AnnounceRequest) []*Peer {
	peer := &Peer{
		ID:         req.PeerID,
		IP:         net.ParseIP(req.IP),
		IPv4:       net.ParseIP(req.IPv4),
		IPv6:       net.ParseIP(req.IPv6),
		ClientIP:   net.ParseIP(req.ClientIP),
		Uploaded:   req.Uploaded,
		Left:       req.Left,
		Downloaded: req.Downloaded,
		LastSeen:   time.Now(),
		Event:      ParsePeerEvent(req.Event),
	}
	root, ok := infoHashMap.Load(req.InfoHash)
	if !ok { // first seen torrent
		root = &InfoHashRoot{peerMap: skipmap.New[string, *Peer]()}
		root.peerMap.Store(peer.GetKey(), peer)
		infoHashMap.Store(req.InfoHash, root)
		return nil
	}
	if knownPeer, ok := root.peerMap.Load(peer.GetKey()); ok {
		// update current record
		knownPeer.Uploaded = peer.Uploaded
		knownPeer.Downloaded = peer.Downloaded
		knownPeer.LastSeen = peer.LastSeen
		knownPeer.Event = peer.Event
		knownPeer.Left = peer.Left
		defer func() {
			logger.Infof("do cleaning for info hash: %s", req.InfoHash)
			go CleanUp(root)
		}()
	} else {
		// new peer!
		root.peerMap.Store(peer.GetKey(), peer)
	}
	resp := make([]*Peer, 0, root.peerMap.Len())
	root.peerMap.Range(func(_ string, value *Peer) bool {
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

func Scrape(infoHash string) *model.ScrapeFile {
	root, ok := infoHashMap.Load(infoHash)
	if !ok {
		return &model.ScrapeFile{
			Complete:   0,
			Incomplete: 0,
			Downloaded: 0,
		}
	}
	var complete, incomplete, downloaded int
	root.peerMap.Range(func(_ string, value *Peer) bool {
		if value.Event == PeerEvent_Completed {
			complete++
			return true
		} else if value.Event == PeerEvent_Started {
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
