package model

// AnnounceRequest Bittorrent Announce Request https://wiki.theory.org/BitTorrent_Tracker_Protocol
type AnnounceRequest struct {
	InfoHash   string `json:"info_hash" query:"info_hash,required" vd:"len($)==20"`
	PeerID     string `json:"peer_id" query:"peer_id,required"  vd:"len($)==20"`
	Port       int    `json:"port" query:"port,required" vd:"$>1 && $<65535"`
	Uploaded   int    `json:"uploaded" query:"uploaded"`
	Downloaded int    `json:"downloaded" query:"downloaded"`
	Event      string `json:"event" query:"event"` // not used
	Left       int    `json:"left" query:"left"`
	NumWant    int    `default:"50" json:"numwant" query:"numwant"`
	Compact    int8   `default:"1" json:"compact" query:"compact"`
	ClientIP   string
	IP         string `query:"ip"`
	IPv4       string `query:"ipv4"`
	IPv6       string `query:"ipv6"`
}
type Peer struct {
	ID   string `json:"id" bencode:"id"`
	IP   string `json:"ip" bencode:"ip"`
	Port int    `json:"port" bencode:"port"`
}

type AnnounceBasicResponse struct {
	Interval int     `json:"interval" bencode:"interval"`
	Peers    []*Peer `json:"peers" bencode:"peers"`
}
