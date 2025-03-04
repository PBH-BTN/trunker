package bittorrent

import (
	"strings"
	"unicode"

	regexp "github.com/wasilibs/go-re2"
)

var (
	commonClients  = regexp.MustCompile(`^-.+-`)                          // common clients, use for qbitorrent, utorrent, vuze, bittorrent, etc
	btSpritClients = regexp.MustCompile(`^-[a-zA-Z]+[0-9]{1,4}`)          // btsprit-like clients, no second -
	aria2Clients   = regexp.MustCompile(`^A2-[0-9]+-[0-9]+-[0-9]+-`)      // Aria2 clients A2-1-35-0-xxxx
	mgClients      = regexp.MustCompile(`^MG-[0-9]+\.[0-9]+\.[0-9]{1,4}`) // MG clients MG-1-35-0-xxxx
	middleDash     = regexp.MustCompile(`^[a-zA-Z]+[0-9]+-`)              // middle dash clients TIX0332-77wfm2rcxovo
)

// ParsePeerID Parse the client name from the peer_id
func ParsePeerID(peerIdRaw string) string {
	peerId := strings.TrimFunc(peerIdRaw, func(r rune) bool {
		return unicode.MaxASCII < r
	})
	if strings.HasPrefix(peerId, "-FD51") { //-FD51]�-FdrWCsIvJAk4
		return "-FD51"
	}
	common := commonClients.FindStringSubmatch(peerId)
	if len(common) > 0 {
		return common[0][:len(common[0])-1]
	}
	btSprit := btSpritClients.FindStringSubmatch(peerId)
	if len(btSprit) > 0 {
		return btSprit[0]
	}
	mg := mgClients.FindStringSubmatch(peerId)
	if len(mg) > 0 {
		return mg[0]
	}
	aria2 := aria2Clients.FindStringSubmatch(peerId)
	if len(aria2) > 0 {
		return aria2[0][:len(aria2[0])-1]
	}
	middle := middleDash.FindStringSubmatch(peerId)
	if len(middle) > 0 {
		return middle[0][:len(middle[0])-1]
	}
	if strings.HasPrefix(peerId, "FD6") { //FD68Ki0o~Jd0mWb(GCY5
		return "FD6"
	}
	if strings.HasPrefix(peerId, "12BS") { //12BS�\u007F]���A%��o\u001F�U\u0006�
		return "12BS"
	}
	return "unknown"
}
