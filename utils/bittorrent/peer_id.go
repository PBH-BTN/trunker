package bittorrent

// ParsePeerID Parse the client name from the peer_id
func ParsePeerID(peerIdRaw string) string {
	if len(peerIdRaw) < 8 { // peer_id must have 20 bytes, this will never happen
		return ""
	}
	client := peerIdRaw[:8]
	if client[0] != '-' {
		return "unknown"
	}
	return client
}
