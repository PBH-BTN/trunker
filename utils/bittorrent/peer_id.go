package bittorrent

import regexp "github.com/wasilibs/go-re2"

var commonClients = regexp.MustCompile(`^-.+-`)           // common clients, use for qbitorrent, utorrent, vuze, bittorrent, etc
var btspritClients = regexp.MustCompile(`^-[A-Z]+[0-9]+`) // btsprit-like clients, no second -

// ParsePeerID Parse the client name from the peer_id
func ParsePeerID(peerIdRaw string) string {
	if len(peerIdRaw) < 8 { // peer_id must have 20 bytes, this will never happen
		return ""
	}
	common := commonClients.FindStringSubmatch(peerIdRaw)
	if len(common) > 0 {
		return common[0]
	}
	slash := btspritClients.FindStringSubmatch(peerIdRaw)
	if len(slash) > 0 {
		return slash[0]
	}
	return "unknown"
}
