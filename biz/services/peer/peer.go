package peer

import (
	"encoding/binary"
	"net"
	"strconv"
	"time"

	"github.com/PBH-BTN/trunker/biz/model"
)

type Peer struct {
	ID         string
	IP         net.IP
	IPv4       net.IP
	IPv6       net.IP
	ClientIP   net.IP
	Port       int
	Uploaded   int       `json:"uploaded"`
	Downloaded int       `json:"downloaded"`
	LastSeen   time.Time `json:"lastSeen"`
}

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
		if peer.GetIP().To4() != nil {
			ip := peer.GetIP().To4() // IPv4 address
			port := make([]byte, 2)
			binary.BigEndian.PutUint16(port, uint16(peer.Port))
			peers = append(peers, ip...)
			peers = append(peers, port...)
		} else if peer.GetIP().To16() != nil {
			ip := peer.GetIP().To16() // IPv6 address
			port := make([]byte, 2)
			binary.BigEndian.PutUint16(port, uint16(peer.Port))
			peers6 = append(peers6, ip...)
			peers6 = append(peers6, port...)
		}
	}
	return peers, peers6
}

func (p *Peer) GetIP() net.IP {
	if p.IP != nil {
		return p.IP
	}
	if p.IPv4 != nil {
		return p.IPv4
	}
	if p.IPv6 != nil {
		return p.IPv6
	}
	if p.ClientIP != nil {
		return p.ClientIP
	}
	return net.IP{}
}

func (p *Peer) GetKey() string {
	return p.GetIP().String() + strconv.FormatInt(int64(p.Port), 10)
}
