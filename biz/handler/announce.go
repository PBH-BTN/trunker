package handler

import (
	"context"
	"math/rand"

	"errors"

	"github.com/PBH-BTN/trunker/biz/config"
	"github.com/PBH-BTN/trunker/biz/model"
	"github.com/PBH-BTN/trunker/biz/services/peer"
	"github.com/PBH-BTN/trunker/biz/services/peer/common"
	"github.com/PBH-BTN/trunker/utils"
	"github.com/PBH-BTN/trunker/utils/bencode"
	"github.com/cloudwego/hertz/pkg/app"
)

func Announce(ctx context.Context, c *app.RequestContext) {
	req := &model.AnnounceRequest{}
	if c.Bind(req) != nil {
		bencode.ResponseErr(c, errors.New("bad request"))
		return
	}
	if !validAnnounceReq(req) {
		bencode.ResponseErr(c, errors.New("bad request"))
		return
	}
	req.ClientIP = c.ClientIP()
	req.UserAgent = string(c.UserAgent())
	res := peer.GetPeerManager().HandleAnnouncePeer(ctx, req)
	scrape := peer.GetPeerManager().Scrape(req.InfoHash)
	if req.Compact == 0 {
		bencode.ResponseOk(c, model.AnnounceBasicResponse{
			Interval: config.AppConfig.Tracker.TTL + int64(rand.Intn(201)-100),
			Peers: utils.Map(res, func(p *common.Peer) *model.Peer {
				return p.ToModel()
			}),
			ExternalIp: req.ClientIP,
			Incomplete: scrape.Incomplete,
			Complete:   scrape.Complete,
		})
	} else {
		resp := map[string]any{
			"interval":    config.AppConfig.Tracker.TTL + int64(rand.Intn(201)-100),
			"external ip": req.ClientIP,
			"incomplete":  scrape.Incomplete,
			"complete":    scrape.Complete,
		}
		peers, peers6 := common.PeersToCompact(res)
		if len(peers) > 0 {
			resp["peers"] = peers
		}
		if len(peers6) > 0 {
			resp["peers6"] = peers6
		}
		bencode.ResponseOk(c, resp)
	}
}

func Scrape(ctx context.Context, c *app.RequestContext) {
	req := &model.ScrapeRequest{}
	if c.Bind(req) != nil {
		bencode.ResponseErr(c, errors.New("bad request"))
		return
	}
	if len(req.InfoHashes) == 0 {
		bencode.ResponseErr(c, errors.New("info hash can't be empty"))
		return
	}
	ret := make(map[string]*model.ScrapeFile)
	manager := peer.GetPeerManager()
	for _, infoHash := range req.InfoHashes {
		ret[infoHash] = manager.Scrape(infoHash)
	}
	bencode.ResponseOk(c, model.ScrapeResponse{Files: ret})
}

func Statistic(ctx context.Context, c *app.RequestContext) {
	c.JSON(200, peer.GetPeerManager().GetStatistic())
}

func validAnnounceReq(req *model.AnnounceRequest) bool {
	if req == nil {
		return false
	}
	if len(req.PeerID)*len(req.InfoHash) == 0 {
		return false
	}
	return req.Port >= 0 && req.Port < 65535
}
