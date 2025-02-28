package udp_server

import (
	"bytes"
	"encoding/binary"
	"net"

	"github.com/PBH-BTN/trunker/biz/model"
	"github.com/PBH-BTN/trunker/biz/services/peer/common"
)

const ProtocolID = uint64(0x41727101980)

const (
	ActionConnect  = uint32(0)
	ActionAnnounce = uint32(1)
	ActionScrape   = uint32(2)
	ActionError    = uint32(3)
)

/*
announce request:

Offset  Size    Name    Value
0       64-bit integer  connection_id
8       32-bit integer  action          1 // announce
12      32-bit integer  transaction_id
16      20-byte string  info_hash
36      20-byte string  peer_id
56      64-bit integer  downloaded
64      64-bit integer  left
72      64-bit integer  uploaded
80      32-bit integer  event           0 // 0: none; 1: completed; 2: started; 3: stopped
84      32-bit integer  IP address      0 // default
88      32-bit integer  key
92      32-bit integer  num_want        -1 // default
96      16-bit integer  port
*/
func parseAnnounceRequestV4(buf []byte) *model.AnnounceRequest {
	req := &model.AnnounceRequest{}
	req.InfoHash = string(buf[0:20])
	req.PeerID = string(buf[20:40])
	req.Downloaded = binary.BigEndian.Uint64(buf[40:48])
	req.Left = binary.BigEndian.Uint64(buf[48:56])
	req.Uploaded = binary.BigEndian.Uint64(buf[56:64])
	event := common.PeerEvent(binary.BigEndian.Uint32(buf[64:68]))
	if event == common.PeerEvent_Unknown { // some client not implemented event
		event = common.PeerEvent_Started
	}
	req.Event = event.String()
	req.IPv4 = net.IP(buf[68:72]).String()
	// key ignored 72:76
	req.NumWant = int(binary.BigEndian.Uint32(buf[76:80]))
	req.Port = int(binary.BigEndian.Uint16(buf[80:82]))
	return req
}

func parseAnnounceRequestV6(buf []byte) *model.AnnounceRequest {
	req := &model.AnnounceRequest{}
	req.InfoHash = string(buf[0:20])
	req.PeerID = string(buf[20:40])
	req.Downloaded = binary.BigEndian.Uint64(buf[40:48])
	req.Left = binary.BigEndian.Uint64(buf[48:56])
	req.Uploaded = binary.BigEndian.Uint64(buf[56:64])
	event := common.PeerEvent(binary.BigEndian.Uint32(buf[64:68]))
	if event == common.PeerEvent_Unknown { // some client not implemented event
		event = common.PeerEvent_Started
	}
	req.Event = event.String()
	req.IPv6 = net.IP(buf[68:84]).String()
	// key ignored 84:88
	req.NumWant = int(binary.BigEndian.Uint32(buf[88:92]))
	req.Port = int(binary.BigEndian.Uint16(buf[92:94]))
	return req
}

type ExtensionType int8
type Extension struct {
	Type   ExtensionType
	Length uint8
	Data   []byte
}

func writeHeader(buf *bytes.Buffer, action uint32, tid uint32) {
	_ = binary.Write(buf, binary.BigEndian, action)
	_ = binary.Write(buf, binary.BigEndian, tid)
}
