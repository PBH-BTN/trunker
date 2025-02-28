package model

import (
	"net"
)

type HttpAnnounceRequest struct {
	InfoHash   string `json:"info_hash" query:"info_hash,required"`
	PeerID     string `json:"peer_id" query:"peer_id,required"`
	Port       int    `json:"port" query:"port,required"`
	Uploaded   uint64 `json:"uploaded" query:"uploaded"`
	Downloaded uint64 `json:"downloaded" query:"downloaded"`
	Event      string `json:"event" query:"event"`
	Left       uint64 `json:"left" query:"left"`
	NumWant    int    `default:"50" json:"numwant" query:"numwant"`
	ClientIP   net.IP
	UserAgent  string
	Type       PeerType `json:"type" query:"type"`
	IP         string   `query:"ip"`
	IPv4       string   `query:"ipv4"`
	IPv6       string   `query:"ipv6"`
	Compact    int8     `default:"1" json:"compact" query:"compact"`
}

// AnnounceRequest Bittorrent Announce Request https://wiki.theory.org/BitTorrent_Tracker_Protocol
type AnnounceRequest struct {
	HttpAnnounceRequest
	Source Source   `json:"source" query:"source"`
	Offers []*Offer `json:"offers" query:"offers"`
	Conn   *Conn    `form:"-" json:"-" query:"-" header:"-"`
}

type Source int8

const (
	SourceHTTP Source = iota
	SourceUDP
	SourceWS
)

type PeerType int8

const (
	PeerTypeBittorrent PeerType = iota
	PeerTypeWebtorrent
)

type Peer struct {
	ID   string `json:"id" bencode:"id"`
	IP   string `json:"ip" bencode:"ip"`
	Port int    `json:"port" bencode:"port"`
}

type AnnounceBasicResponse struct {
	Interval   int64    `json:"interval" bencode:"interval"`
	Peers      []*Peer  `json:"peers" bencode:"peers"`
	ExternalIp []byte   `json:"externalIp" bencode:"external ip"`
	Complete   int      `json:"complete" bencode:"complete"`
	Incomplete int      `json:"incomplete" bencode:"incomplete"`
	Offers     []*Offer `json:"offers,omitempty" bencode:"offers,omitempty"`
}

type Offer struct {
	OfferID string      `json:"offer_id" query:"offer_id"`
	Offer   OfferDetail `json:"offer" query:"offer"`
}
type OfferDetail struct {
	Type string `json:"type" query:"type"`
	SDP  string `json:"sdp" query:"sdp"`
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
	Seeder     int `json:"seeder" bencode:"-"`
	Complete   int `json:"complete" bencode:"complete"`
	Incomplete int `json:"incomplete" bencode:"incomplete"`
	Downloaded int `json:"downloaded" bencode:"downloaded"`
}
