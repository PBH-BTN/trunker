package entity

import (
	"time"

	"github.com/PBH-BTN/trunker/biz/model"
	"gorm.io/datatypes"
)

type Peers struct {
	Ip         []byte                            `json:"ip" `                  // -
	Ipv4       []byte                            `json:"ipv4" `                // -
	Ipv6       []byte                            `json:"ipv6" `                // -
	ClientIp   []byte                            `json:"client_ip" `           // The ip from the heep client
	LastSeen   time.Time                         `json:"last_seen" `           // Last seen of this peer
	Offers     datatypes.JSONSlice[*model.Offer] `json:"offers" `              // -
	UpdatedAt  time.Time                         `json:"updated_at" `          // -
	InfoHash   string                            `json:"info_hash" `           // Info hash
	PeerID     string                            `json:"peer_id" `             // -
	UserAgent  string                            `json:"user_agent" `          // -
	ID         uint64                            `json:"id" gorm:"primaryKey"` // id
	Port       int                               `json:"port" `                // -
	Left       uint64                            `json:"left" `                // Left size of the peer
	Uploaded   uint64                            `json:"uploaded" `            // Uploaded size
	Downloaded uint64                            `json:"downloaded" `          // Downloaded size
	Type       model.PeerType                    `json:"type"`                 // peer type, 0 - bittorrent， 1- webtorrent
	Source     model.Source                      `json:"source" `              // peer source, 0 - http， 1 - udp, 2 - ws
	Event      int8                              `json:"event" `               // Peer event reported
}

func (m *Peers) TableName() string {
	return "peers"
}
