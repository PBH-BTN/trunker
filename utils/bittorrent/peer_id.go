package bittorrent

import (
	"strings"

	"github.com/cloudwego/hertz/pkg/common/hlog"
	regexp "github.com/wasilibs/go-re2"
)

var (
	commonClients  = regexp.MustCompile(`^-.+-`)                     // common clients, use for qbitorrent, utorrent, vuze, bittorrent, etc
	btSpritClients = regexp.MustCompile(`^-[a-zA-Z]+[0-9]{4}`)       // btsprit-like clients, no second -
	aria2Clients   = regexp.MustCompile(`^A2-[0-9]+-[0-9]+-[0-9]+-`) // Aria2 clients A2-1-35-0-xxxx
	middleDash     = regexp.MustCompile(`^[a-zA-Z]+[0-9]+-`)         // middle dash clients TIX0332-77wfm2rcxovo
)

// ParsePeerID Parse the client name from the peer_id
func ParsePeerID(peerIdRaw string) string {
	if len(peerIdRaw) < 8 { // peer_id must have 20 bytes, this will never happen
		return ""
	}
	common := commonClients.FindStringSubmatch(peerIdRaw)
	if len(common) > 0 {
		return common[0][:len(common[0])-1]
	}
	btSprit := btSpritClients.FindStringSubmatch(peerIdRaw)
	if len(btSprit) > 0 {
		return btSprit[0]
	}
	aria2 := aria2Clients.FindStringSubmatch(peerIdRaw)
	if len(aria2) > 0 {
		return aria2[0][:len(aria2[0])-1]
	}
	middle := middleDash.FindStringSubmatch(peerIdRaw)
	if len(middle) > 0 {
		return middle[0][:len(middle[0])-1]
	}
	if strings.HasPrefix(peerIdRaw, "FD6") { //FD68Ki0o~Jd0mWb(GCY5
		return "FD6"
	}
	hlog.Info("unknown peer id: ", peerIdRaw)
	return "unknown"
}
