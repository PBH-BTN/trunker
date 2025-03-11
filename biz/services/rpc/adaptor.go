package rpc

import (
	"github.com/PBH-BTN/trunker/biz/model"
	"github.com/PBH-BTN/trunker/biz/services/peer/common"
	"github.com/PBH-BTN/trunker/kitex_gen/pbh/btn/trunker"
	"github.com/PBH-BTN/trunker/utils"
)

func announceRequestCommonToIDL(req *model.AnnounceRequest) *trunker.AnnounceRequest {
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
func announceRequestIDLToCommon(req *trunker.AnnounceRequest) *model.AnnounceRequest {
	return &model.AnnounceRequest{
		HttpAnnounceRequest: model.HttpAnnounceRequest{
			InfoHash:   req.InfoHash,
			PeerID:     req.PeerId,
			IP:         req.Ip,
			Port:       int(req.Port),
			IPv4:       req.Ipv4,
			IPv6:       req.Ipv6,
			ClientIP:   req.ClientIp,
			Uploaded:   uint64(req.Uploaded),
			Downloaded: uint64(req.Downloaded),
			Left:       uint64(req.Left),
			NumWant:    int(req.NumWant),
			Type:       model.PeerType(req.Type),
			Compact:    int8(utils.If(req.Compact, 1, 0)),
			Event:      req.Event.String(),
		},
		Source: model.Source(req.Source),
		Offers: utils.Map(req.Offers, func(o *trunker.Offer) *model.Offer {
			return &model.Offer{
				OfferID: o.OfferId,
				Offer: model.OfferDetail{
					Type: o.Offer.Type,
					SDP:  o.Offer.Sdp,
				},
			}
		}),
	}
}

func peerCommonToIDL(p *common.Peer) *trunker.Peer {
	return &trunker.Peer{
		Ip:       p.IP,
		Ipv4:     p.IPv4,
		Ipv6:     p.IPv6,
		ClientIp: p.ClientIP,
		LastSeen: p.LastSeen.Unix(),
		Offers: utils.Map(p.Offers, func(o *model.Offer) *trunker.Offer {
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
