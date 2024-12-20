package handler

import (
	"context"
	"math/rand"
	"net"

	"errors"

	"github.com/PBH-BTN/trunker/biz/config"
	"github.com/PBH-BTN/trunker/biz/model"
	"github.com/PBH-BTN/trunker/biz/services/peer"
	"github.com/PBH-BTN/trunker/biz/services/peer/common"
	"github.com/PBH-BTN/trunker/utils"
	"github.com/PBH-BTN/trunker/utils/bencode"
	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/common/hlog"
	"github.com/thinkeridea/go-extend/exstrings"
)

var ipHeader = []string{
	"X-Forwarded-For",
	"X-Real-IP",
}

func getClientIP(ctx context.Context, c *app.RequestContext) net.IP {
	// There is a bug in c.ClientIP() while using Unix Domain Socket, unfortunately hertz don't want to fix this.
	// So we have to use this workaround.
	for _, header := range ipHeader {
		ip := c.Request.Header.Get(header)
		if ip != "" {
			res := net.ParseIP(ip)
			if res != nil {
				return res
			}
			hlog.CtxWarnf(ctx, "invalid ip from header %s:%s", header, ip)
		}
	}
	ip := c.ClientIP()
	if ip == "" {
		hlog.CtxWarnf(ctx, "failed to get client ip,header:%s", c.Request.Header.Header())
		return nil
	}
	return net.ParseIP(ip)
}

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
	req.ClientIP = getClientIP(ctx, c)
	req.UserAgent = exstrings.SubString(string(c.UserAgent()), 0, 256)
	if req.NumWant == 0 {
		req.NumWant = 50
	}
	res := peer.GetPeerManager().HandleAnnouncePeer(ctx, req)
	scrape := peer.GetPeerManager().Scrape(req.InfoHash)
	if req.Compact == 0 {
		bencode.ResponseOk(c, model.AnnounceBasicResponse{
			Interval: config.AppConfig.Tracker.TTL + int64(rand.Intn(201)-100),
			Peers: utils.Map(res, func(p *common.Peer) *model.Peer {
				return p.ToModel()
			}),
			ExternalIp: req.ClientIP.String(),
			Incomplete: scrape.Incomplete,
			Complete:   scrape.Complete,
		})
	} else {
		resp := map[string]any{
			"interval":    config.AppConfig.Tracker.TTL + int64(rand.Intn(201)-100),
			"external ip": req.ClientIP.String(),
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

func Scrape(_ context.Context, c *app.RequestContext) {
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

func Statistic(_ context.Context, c *app.RequestContext) {
	c.JSON(200, peer.GetPeerManager().GetStatistic())
}

func validAnnounceReq(req *model.AnnounceRequest) bool {
	if req == nil {
		return false
	}
	if len(req.InfoHash) != 20 || len(req.PeerID) != 20 {
		return false
	}
	return req.Port >= 0 && req.Port < 65535
}
