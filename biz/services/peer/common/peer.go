package common

import (
	"encoding/binary"
	"net"
	"time"

	"github.com/PBH-BTN/trunker/biz/config"
	"github.com/PBH-BTN/trunker/biz/model"
	"github.com/hertz-contrib/websocket"
)

type Peer struct {
	ID         string
	IP         net.IP
	IPv4       net.IP
	IPv6       net.IP
	ClientIP   net.IP
	Port       int
	Left       uint64
	Type       PeerType
	Uploaded   uint64    `json:"uploaded"`
	Downloaded uint64    `json:"downloaded"`
	LastSeen   time.Time `json:"lastSeen"`
	UserAgent  string
	Event      PeerEvent
	Offers     []*Offer        `json:"offers"`
	Conn       *websocket.Conn `json:"-"`
}
type PeerType = model.PeerType
type Offer = model.Offer
type OfferDetail = model.OfferDetail

func (p *Peer) ToModel() *model.Peer {
	if p == nil {
		return nil
	}
	return &model.Peer{
		ID:   p.ID,
		IP:   p.GetIP().String(),
		Port: p.Port,
	}
}

func PeersToCompact(peerList []*Peer) ([]byte, []byte) {
	var peers []byte
	var peers6 []byte
	for _, peer := range peerList {
		if ip := peer.GetIP().To4(); ip != nil { // IPv4 address
			port := make([]byte, 2)
			binary.BigEndian.PutUint16(port, uint16(peer.Port))
			peers = append(peers, ip...)
			peers = append(peers, port...)
		} else if ip := peer.GetIP().To16(); ip != nil { // IPv6 address
			port := make([]byte, 2)
			binary.BigEndian.PutUint16(port, uint16(peer.Port))
			peers6 = append(peers6, ip...)
			peers6 = append(peers6, port...)
		}
	}
	return peers, peers6
}

func (p *Peer) GetIP() net.IP {
	if config.AppConfig.Tracker.UseAnnounceIP {
		if p.IP != nil {
			return p.IP
		}
		if p.IPv4 != nil {
			return p.IPv4
		}
		if p.IPv6 != nil {
			return p.IPv6
		}
	}
	if p.ClientIP != nil {
		return p.ClientIP
	}
	return net.IP{}
}

func (p *Peer) GetKey() string {
	return p.ID
}

type PeerEvent int8

const (
	PeerEvent_Unknown PeerEvent = iota
	PeerEvent_Started
	PeerEvent_Stopped
	PeerEvent_Completed
)

func ParsePeerEvent(s string) PeerEvent {
	switch s {
	case "started":
		return PeerEvent_Started
	case "stopped":
		return PeerEvent_Stopped
	case "completed":
		return PeerEvent_Completed
	default:
		return PeerEvent_Unknown
	}
}

func (e PeerEvent) String() string {
	switch e {
	case PeerEvent_Started:
		return "started"
	case PeerEvent_Stopped:
		return "stopped"
	case PeerEvent_Completed:
		return "completed"
	default:
		return "unknown"
	}
}

// IsPeerConnectable Check If Peer is connectable
func IsPeerConnectable(peer *Peer) bool {
	if peer.Type == model.PeerTypeBittorrent {
		return !(peer.GetIP().IsPrivate() || peer.GetIP().IsLoopback() || peer.Port == 0 || peer.Port == 1)
	} else if peer.Type == model.PeerTypeWebtorrent {
		return len(peer.Offers) > 0
	}
	return false
}
