package rpc

import (
	"github.com/PBH-BTN/trunker/biz/model"
	"github.com/PBH-BTN/trunker/biz/services/peer/common"
	"github.com/PBH-BTN/trunker/kitex_gen/pbh/btn/trunker"
	"github.com/bytedance/gg/gcond"
	"github.com/bytedance/gg/gslice"
)

func announceRequestIDLToCommon(req *trunker.AnnounceRequest) *model.AnnounceRequest {
	if req == nil {
		return nil
	}
	return &model.AnnounceRequest{
		HttpAnnounceRequest: model.HttpAnnounceRequest{
			InfoHash:   req.InfoHash,
			PeerID:     req.PeerId,
			IP:         req.Ip,
			Port:       int(req.Port),
			IPv4:       req.Ipv4,
			IPv6:       req.Ipv6,
			ClientIP:   req.ClientIp,
			Uploaded:   req.Uploaded,
			Downloaded: req.Downloaded,
			Left:       req.Left,
			NumWant:    int(req.NumWant),
			Type:       model.PeerType(req.Type),
			Compact:    int8(gcond.If(req.Compact, 1, 0)),
			Event:      req.Event.String(),
		},
		Source: model.Source(req.Source),
		Offers: gslice.Map(req.Offers, func(o *trunker.Offer) *model.Offer {
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
	if p == nil {
		return nil
	}
	return &trunker.Peer{
		Ip:       p.IP,
		Ipv4:     p.IPv4,
		Ipv6:     p.IPv6,
		ClientIp: p.ClientIP,
		LastSeen: p.LastSeen.Unix(),
		Offers: gslice.Map(p.Offers, func(o *model.Offer) *trunker.Offer {
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

func sourceToMetrics(s trunker.Source) string {
	source := "http"
	switch s {
	case trunker.Source_HTTP:
		source = "http"
	case trunker.Source_UDP:
		source = "udp"
	case trunker.Source_WS:
		source = "ws"
	}
	return source
}
