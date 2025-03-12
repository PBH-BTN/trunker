package rpc

import (
	"time"

	"github.com/PBH-BTN/trunker/biz/model"
	"github.com/PBH-BTN/trunker/biz/services/peer/common"
	"github.com/PBH-BTN/trunker/kitex_gen/pbh/btn/trunker"
	"github.com/PBH-BTN/trunker/utils"
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
		Uploaded:   int64(req.Uploaded),
		Downloaded: int64(req.Downloaded),
		Left:       int64(req.Left),
		NumWant:    int64(req.NumWant),
		Type:       trunker.PeerType(req.Type),
		Compact:    req.Compact == 1,
		Source:     trunker.Source(req.Source),
		Event:      trunker.PeerEvent(common.ParsePeerEvent(req.Event)),
		Offers: utils.Map(req.Offers, func(o *model.Offer) *trunker.Offer {
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

func peerIDLToCommon(p *trunker.Peer) *common.Peer {
	if p == nil {
		return nil
	}
	return &common.Peer{
		IP:       p.Ip,
		IPv4:     p.Ipv4,
		IPv6:     p.Ipv6,
		ClientIP: p.ClientIp,
		LastSeen: time.Unix(p.LastSeen, 0),
		Offers: utils.Map(p.Offers, func(o *trunker.Offer) *model.Offer {
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
