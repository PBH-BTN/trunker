package handler

import (
	"context"

	"github.com/PBH-BTN/trunker/biz/model"
	"github.com/PBH-BTN/trunker/biz/services/peer"
	"github.com/PBH-BTN/trunker/utils/bencode"
	"github.com/cloudwego/hertz/pkg/app"
)

func Announce(ctx context.Context, c *app.RequestContext) {
	req := &model.AnnounceRequest{}
	if c.BindAndValidate(req) != nil {
		c.JSON(400, "Bad Request")
		return
	}
	req.IP = c.ClientIP()
	res := peer.HandleAnnouncePeer(req)
	bencode.ResponseOk(c, model.AnnounceBasicResponse{
		Interval: 60,
		Peers:    res,
	})
}
