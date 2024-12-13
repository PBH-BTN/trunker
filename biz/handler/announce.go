package handler

import (
	"context"

	"github.com/PBH-BTN/trunker/biz/model"
	"github.com/PBH-BTN/trunker/biz/services/peer"
	"github.com/PBH-BTN/trunker/utils"
	"github.com/PBH-BTN/trunker/utils/bencode"
	"github.com/PBH-BTN/trunker/utils/conv"
	"github.com/cloudwego/hertz/pkg/app"
)

func Announce(ctx context.Context, c *app.RequestContext) {
	req := &model.AnnounceRequest{}
	if c.BindAndValidate(req) != nil {
		c.JSON(400, "Bad Request")
		return
	}
	req.ClientIP = c.ClientIP()
	req.UserAgent = conv.UnsafeBytesToString(c.UserAgent())
	res := peer.HandleAnnouncePeer(req)
	if req.Compact == 0 {
		bencode.ResponseOk(c, model.AnnounceBasicResponse{
			Interval: 60,
			Peers: utils.Map(res, func(p *peer.Peer) *model.Peer {
				return p.ToModel()
			}),
			ExternalIp: req.ClientIP,
		})
	} else {
		resp := map[string]any{
			"interval":    60,
			"external ip": req.ClientIP,
		}
		peers, peers6 := peer.PeersToCompact(res)
		if len(peers) > 0 {
			resp["peers"] = peers
		}
		if len(peers6) > 0 {
			resp["peers6"] = peers6
		}
		bencode.ResponseOk(c, resp)
	}
}
