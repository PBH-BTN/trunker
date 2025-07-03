package rpc

import (
	"time"

	"github.com/PBH-BTN/trunker/biz/model"
	"github.com/PBH-BTN/trunker/biz/services/peer/common"
	"github.com/PBH-BTN/trunker/kitex_gen/pbh/btn/trunker"
	"github.com/bytedance/gg/gslice"
)

func announceRequestCommonToIDL(req *model.AnnounceRequest) *trunker.AnnounceRequest {
	if req == nil {
		return nil
	}
	return &trunker.AnnounceRequest{
		InfoHash:   req.InfoHash,
		PeerId:     req.PeerID,
		Ip:         req.IP,
		Port:       int32(req.Port),
		Ipv4:       req.IPv4,
		Ipv6:       req.IPv6,
		ClientIp:   req.ClientIP,
		Uploaded:   req.Uploaded,
		Downloaded: req.Downloaded,
		Left:       req.Left,
		NumWant:    int64(req.NumWant),
		Type:       trunker.PeerType(req.Type),
		Compact:    req.Compact == 1,
		Source:     trunker.Source(req.Source),
		Event:      trunker.PeerEvent(common.ParsePeerEvent(req.Event)),
		Offers: gslice.Map(req.Offers, func(o *model.Offer) *trunker.Offer {
			return &trunker.Offer{
				OfferId: o.OfferID,
				Offer: &trunker.OfferDetail{
					Type: o.Offer.Type,
					Sdp:  o.Offer.SDP,
				},
			}
		}),
	}
}

func PeerIDLToCommon(p *trunker.Peer) *common.Peer {
	if p == nil {
		return nil
	}
	return &common.Peer{
		IP:       p.Ip,
		IPv4:     p.Ipv4,
		IPv6:     p.Ipv6,
		ClientIP: p.ClientIp,
		LastSeen: time.Unix(p.LastSeen, 0),
		Offers: gslice.Map(p.Offers, func(o *trunker.Offer) *model.Offer {
			return &model.Offer{
				OfferID: o.OfferId,
				Offer: model.OfferDetail{
					Type: o.Offer.Type,
					SDP:  o.Offer.Sdp,
				},
			}
		}),
		ID:         p.Id,
		UserAgent:  p.UserAgent,
		Port:       int(p.Port),
		Uploaded:   uint64(p.Uploaded),
		Downloaded: uint64(p.Downloaded),
		Left:       uint64(p.Left),
		Type:       model.PeerType(p.Type),
		Event:      common.PeerEvent(p.Event),
		Source:     model.Source(p.Source),
	}
}

func PeerCommonToIDL(p *common.Peer) *trunker.Peer {
	if p == nil {
		return nil
	}
	return &trunker.Peer{
		Ip:       p.IP,
		Ipv4:     p.IPv4,
		Ipv6:     p.IPv6,
		ClientIp: p.ClientIP,
		LastSeen: p.LastSeen.Unix(),
		Offers: gslice.Map(p.Offers, func(o *common.Offer) *trunker.Offer {
			return &trunker.Offer{
				OfferId: o.OfferID,
				Offer: &trunker.OfferDetail{
					Type: o.Offer.Type,
					Sdp:  o.Offer.SDP,
				},
			}
		}),
		Id:         p.ID,
		UserAgent:  p.UserAgent,
		Port:       int32(p.Port),
		Uploaded:   int64(p.Uploaded),
		Downloaded: int64(p.Downloaded),
		Left:       int64(p.Left),
		Type:       trunker.PeerType(p.Type),
		Event:      trunker.PeerEvent(p.Event),
		Source:     trunker.Source(p.Source),
	}
}
