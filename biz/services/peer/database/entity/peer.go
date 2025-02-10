package entity

import (
	"time"
)

type Peers struct {
	ID         uint64    `json:"id" gorm:"primaryKey"` // id
	InfoHash   string    `json:"info_hash" `           // Info hash
	PeerID     string    `json:"peer_id" `
	Ip         []byte    `json:"ip" `
	Ipv4       []byte    `json:"ipv4" `
	Ipv6       []byte    `json:"ipv6" `
	ClientIp   []byte    `json:"client_ip" ` // The ip from the heep client
	Port       int       `json:"port" `
	Left       uint64    `json:"left" `       // Left size of the peer
	Uploaded   uint64    `json:"uploaded" `   // Uploaded size
	Downloaded uint64    `json:"downloaded" ` // Downloaded size
	LastSeen   time.Time `json:"last_seen" `  // Last seen of this peer
	UserAgent  string    `json:"user_agent" `
	Event      int8      `json:"event" ` // Peer event reported
	UpdatedAt  time.Time `json:"updated_at" `
}

func (m *Peers) TableName() string {
	return "peers"
}
