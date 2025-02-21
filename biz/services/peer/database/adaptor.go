package database

import (
	"encoding/hex"
	"time"

	"github.com/PBH-BTN/trunker/biz/services/peer/common"
	"github.com/PBH-BTN/trunker/biz/services/peer/database/entity"
	"github.com/PBH-BTN/trunker/utils/conv"
	"gorm.io/datatypes"
)

func CommonToDB(infoHash string, peer *common.Peer) *entity.Peers {
	infoHashSafe := hex.EncodeToString(conv.UnsafeStringToBytes(infoHash))
	peerIdSafe := hex.EncodeToString(conv.UnsafeStringToBytes(peer.ID))
	return &entity.Peers{
		InfoHash:   infoHashSafe,
		PeerID:     peerIdSafe,
		Ip:         peer.IP,
		Ipv4:       peer.IPv4,
		Ipv6:       peer.IPv6,
		ClientIp:   peer.ClientIP,
		Port:       peer.Port,
		Type:       peer.Type,
		Left:       peer.Left,
		Uploaded:   peer.Uploaded,
		Downloaded: peer.Downloaded,
		LastSeen:   peer.LastSeen,
		UserAgent:  peer.UserAgent,
		Event:      int8(peer.Event),
		Offers:     datatypes.NewJSONSlice(peer.Offers),
		UpdatedAt:  time.Now(),
	}
}

func DBToCommon(peer *entity.Peers) *common.Peer {
	peerId, _ := hex.DecodeString(peer.PeerID)
	return &common.Peer{
		ID:         conv.UnsafeBytesToString(peerId),
		IP:         peer.Ip,
		IPv4:       peer.Ipv4,
		IPv6:       peer.Ipv6,
		ClientIP:   peer.ClientIp,
		Port:       peer.Port,
		Type:       peer.Type,
		Left:       peer.Left,
		Uploaded:   peer.Uploaded,
		Downloaded: peer.Downloaded,
		LastSeen:   peer.LastSeen,
		UserAgent:  peer.UserAgent,
		Offers:     peer.Offers,
		Event:      common.PeerEvent(peer.Event),
	}
}
