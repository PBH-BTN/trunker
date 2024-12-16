package model

// AnnounceRequest Bittorrent Announce Request https://wiki.theory.org/BitTorrent_Tracker_Protocol
type AnnounceRequest struct {
	InfoHash   string `json:"info_hash" query:"info_hash,required"`
	PeerID     string `json:"peer_id" query:"peer_id,required"`
	Port       int    `json:"port" query:"port,required"`
	Uploaded   int    `json:"uploaded" query:"uploaded"`
	Downloaded int    `json:"downloaded" query:"downloaded"`
	Event      string `json:"event" query:"event"`
	Left       int    `json:"left" query:"left"`
	NumWant    int    `default:"50" json:"numwant" query:"numwant"`
	ClientIP   string
	UserAgent  string
	IP         string `query:"ip"`
	IPv4       string `query:"ipv4"`
	IPv6       string `query:"ipv6"`
	Compact    int8   `default:"1" json:"compact" query:"compact"`
}
type Peer struct {
	ID   string `json:"id" bencode:"id"`
	IP   string `json:"ip" bencode:"ip"`
	Port int    `json:"port" bencode:"port"`
}

type AnnounceBasicResponse struct {
	Interval   int64   `json:"interval" bencode:"interval"`
	Peers      []*Peer `json:"peers" bencode:"peers"`
	ExternalIp string  `json:"externalIp" bencode:"external ip"`
	Complete   int     `json:"complete" bencode:"complete"`
	Incomplete int     `json:"incomplete" bencode:"incomplete"`
}

type ErrorResponse struct {
	FailureReason string `json:"failureReason" bencode:"failure reason"`
	Retry         string `json:"retry" bencode:"retry in"`
}

type ScrapeRequest struct {
	InfoHashes []string `query:"info_hash"`
}

type ScrapeResponse struct {
	Files map[string]*ScrapeFile `json:"files" bencode:"files"`
}

type ScrapeFile struct {
	Complete   int `json:"complete" bencode:"complete"`
	Incomplete int `json:"incomplete" bencode:"incomplete"`
	Downloaded int `json:"downloaded" bencode:"downloaded"`
}
