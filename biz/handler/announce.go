package handler

import (
	"context"
	"errors"
	"strings"

	"github.com/PBH-BTN/trunker/biz/config"
	"github.com/PBH-BTN/trunker/biz/model"
	"github.com/PBH-BTN/trunker/biz/services/peer"
	"github.com/PBH-BTN/trunker/biz/services/peer/common"
	"github.com/PBH-BTN/trunker/service/metrics"
	"github.com/PBH-BTN/trunker/utils/bencode"
	"github.com/PBH-BTN/trunker/utils/bittorrent"
	"github.com/PBH-BTN/trunker/utils/conv"
	"github.com/PBH-BTN/trunker/utils/http"
	"github.com/bytedance/gg/gslice"
	"github.com/bytedance/gopkg/lang/fastrand"
	"github.com/bytedance/gopkg/util/gopool"
	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/protocol/consts"
	"github.com/thinkeridea/go-extend/exstrings"
)

func Announce(ctx context.Context, c *app.RequestContext) {
	if config.AppConfig.Tracker.Mode == config.RunningModeMemory && config.AppConfig.Tracker.WSServer.Enable {
		if strings.EqualFold(c.Request.Header.Get(consts.HeaderConnection), "Upgrade") && strings.EqualFold(c.Request.Header.Get("Upgrade"), "websocket") {
			HandleWebTorrent(ctx, c)
			return
		}
	}
	req := model.HttpAnnounceRequest{}
	if err := c.Bind(&req); err != nil {
		metrics.EmitCounter(metrics.CounterInvalidRequest, 1, map[string]string{
			metrics.LabelReason: "bind error",
		})
		bencode.ResponseErr(c, errors.New("bad request"))
		return
	}
	if !validAnnounceReq(&req) {
		bencode.ResponseErr(c, errors.New("bad request"))
		return
	}
	gopool.CtxGo(ctx, func() {
		metrics.EmitCounter(metrics.CounterAnnounce, 1, map[string]string{
			metrics.LabelSource: "http",
			metrics.LabelClient: bittorrent.ParsePeerID(req.PeerID),
		})
	})
	req.ClientIP = http.GetClientIP(ctx, c)
	if v4 := req.ClientIP.To4(); v4 != nil {
		req.ClientIP = v4
	}
	// workaround for memory issue
	req.UserAgent = ""
	if req.NumWant > config.AppConfig.Tracker.Memory.MaxPeersPerTorrent {
		req.NumWant = config.AppConfig.Tracker.Memory.MaxPeersPerTorrent
	}
	if req.NumWant <= 0 || req.NumWant > 500 {
		req.NumWant = 50
	}
	req.Type = model.PeerTypeBittorrent
	scrape, err := peer.GetPeerManager().Scrape(ctx, req.InfoHash)
	if err != nil {
		bencode.ResponseErr(c, err)
		return
	}
	res, err := peer.GetPeerManager().HandleAnnouncePeer(ctx, &model.AnnounceRequest{HttpAnnounceRequest: req, Source: model.SourceHTTP})
	if err != nil {
		bencode.ResponseErr(c, err)
		return
	}
	if req.Compact == 0 {
		bencode.ResponseOk(c, model.AnnounceBasicResponse{
			Interval: config.AppConfig.Tracker.TTL + int64(fastrand.Intn(201)-100),
			Peers: gslice.Map(res, func(p *common.Peer) *model.Peer {
				return p.ToModel()
			}),
			ExternalIp: conv.UnsafeBytesToString(req.ClientIP),
			Incomplete: scrape.Incomplete,
			TrackerId:  config.AppConfig.Tracker.TrackerId,
			Complete:   scrape.Complete,
		})
	} else {
		peers, peers6 := common.PeersToCompact(res)
		bencode.ResponseOk(c, bencode.FastBencode(&model.AnnounceCompactResponse{
			Interval:   config.AppConfig.Tracker.TTL + int64(fastrand.Intn(201)-100),
			Peers:      peers,
			Peers6:     peers6,
			ExternalIp: conv.UnsafeBytesToString(req.ClientIP),
			TrackerId:  config.AppConfig.Tracker.TrackerId,
			Complete:   scrape.Complete,
			Incomplete: scrape.Incomplete,
		}))
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
		var err error
		ret[infoHash], err = manager.Scrape(ctx, infoHash)
		if err != nil {
			bencode.ResponseErr(c, err)
			return
		}
	}
	bencode.ResponseOk(c, model.ScrapeResponse{Files: ret})
}

func Statistic(ctx context.Context, c *app.RequestContext) {
	info, err := peer.GetPeerManager().GetStatistic(ctx)
	if err != nil {
		http.ResponseErr(c, err)
		return
	}
	http.ResponseOK(c, info)
}

func validAnnounceReq(req *model.HttpAnnounceRequest) bool {
	if req == nil {
		return false
	}
	if len(req.InfoHash) != 20 || len(req.PeerID) != 20 {
		metrics.EmitCounter(metrics.CounterInvalidRequest, 1, map[string]string{
			metrics.LabelReason: "info hash or peer id length error",
		})
		return false
	}
	if !(req.Port >= 0 && req.Port < 65535) {
		metrics.EmitCounter(metrics.CounterInvalidRequest, 1, map[string]string{
			metrics.LabelReason: "invalid port",
		})
		return false
	}
	return true
}
