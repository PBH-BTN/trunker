package handler

import (
	"context"
	"encoding/json"
	"errors"
	"math/rand"

	"github.com/PBH-BTN/trunker/biz/config"
	"github.com/PBH-BTN/trunker/biz/model"
	"github.com/PBH-BTN/trunker/biz/services/peer"
	"github.com/PBH-BTN/trunker/service/metrics"
	"github.com/PBH-BTN/trunker/utils"
	"github.com/PBH-BTN/trunker/utils/conv"
	"github.com/PBH-BTN/trunker/utils/http"
	"github.com/PBH-BTN/trunker/utils/webtorrent"
	"github.com/bytedance/sonic"
	"github.com/bytedance/sonic/ast"
	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/common/hlog"
	hertz "github.com/cloudwego/hertz/pkg/common/utils"
	"github.com/hertz-contrib/websocket"
	"github.com/lestrrat-go/choose"
	"github.com/thinkeridea/go-extend/exstrings"
)

var u = websocket.HertzUpgrader{
	CheckOrigin: func(ctx *app.RequestContext) bool {
		return true // don't check origin for client
	},
} // use default options
func HandleWebTorrent(ctx context.Context, c *app.RequestContext) {
	err := u.Upgrade(c, func(conn *websocket.Conn) {
		wrapConn := model.NewConn(conn)
		defer wrapConn.Close()
		for {
			_, message, err := conn.ReadMessage()
			if err != nil {
				if websocket.IsCloseError(err, websocket.CloseNoStatusReceived, websocket.CloseNormalClosure, websocket.CloseGoingAway) {
					_ = wrapConn.Close()
					break
				}
				hlog.CtxErrorf(ctx, "failed to read from websocket:%s", err.Error())
				break
			}
			hlog.CtxDebugf(ctx, "ws received: %s", message)
			actionRaw, err := sonic.Get(message, "action")
			if err != nil {
				hlog.CtxErrorf(ctx, "get action error: %s", err.Error())
				_ = webtorrent.ResponseErr(conn, errors.New("get action error"))
				_ = conn.Close()
				return
			}
			action, err := actionRaw.String()
			if err != nil {
				hlog.CtxErrorf(ctx, "get action error: %s", err.Error())
				_ = webtorrent.ResponseErr(conn, errors.New("get action error"))
				_ = conn.Close()
				return
			}
			switch action {
			case "announce":
				var answerRaw ast.Node
				if answerRaw, err = sonic.Get(message, "answer"); err == nil {
					if answerRaw.Valid() {
						err = handleWSAnswer(ctx, message)
						break
					}
				}
				err = handleWSAnnounce(ctx, message, c, wrapConn)
			case "scrape":
				err = handleWSScrape(ctx, message, wrapConn)
			default:
				_ = webtorrent.ResponseErr(conn, errors.New("invalid action"))
				_ = conn.Close()
				return
			}
			if err != nil {
				hlog.CtxErrorf(ctx, "handle action error: %s", err.Error())
				_ = webtorrent.ResponseErr(conn, err)
				_ = conn.Close()
				return
			}
		}
	})
	if err != nil {
		hlog.CtxInfof(ctx, "websocket upgrade error: %s", err.Error())
		http.ResponseErrCustom(c, 406, errors.New("only support websocket protocol"))
		return
	}
}
func handleWSAnswer(ctx context.Context, msg []byte) error {
	infoHashRaw, err := sonic.Get(msg, "info_hash")
	if err != nil {
		return err
	}
	infoHash, err := infoHashRaw.String()
	if err != nil {
		return err
	}
	infoHash = string(conv.TransUTF8To9959_1(conv.UnsafeStringToBytes(infoHash)))
	peerIdRaw, err := sonic.Get(msg, "to_peer_id")
	if err != nil {
		return err
	}
	peerId, err := peerIdRaw.String()
	if err != nil {
		return err
	}
	return peer.GetPeerManager().AnswerToPeer(ctx, infoHash, peerId, msg)
}

func handleWSAnnounce(ctx context.Context, msg []byte, c *app.RequestContext, conn *model.Conn) error {
	req := &model.AnnounceRequest{}
	if err := json.Unmarshal(msg, req); err != nil {
		metrics.EmitCounter(metrics.CounterInvalidRequest, 1, map[string]string{
			metrics.LabelReason: "bind error",
		})
		return err
	}
	// The raw info hash is an utf-8 encoded bytes, which should be converted to iso-8859-1
	req.InfoHash = string(conv.TransUTF8To9959_1(conv.UnsafeStringToBytes(req.InfoHash)))
	if !validAnnounceReq(req) {
		return errors.New("invalid request")
	}
	req.ClientIP = http.GetClientIP(ctx, c)
	req.UserAgent = exstrings.SubString(string(c.UserAgent()), 0, 256)
	if req.NumWant == 0 || req.NumWant > 500 {
		req.NumWant = 50
	}
	req.Conn = conn
	req.Type = model.PeerTypeWebtorrent
	res, err := peer.GetPeerManager().HandleAnnouncePeer(ctx, req)
	if err != nil {
		return err
	}
	scrape, err := peer.GetPeerManager().Scrape(ctx, req.InfoHash)
	if err != nil {
		return err
	}
	resp := hertz.H{
		"action":     "announce",
		"interval":   config.AppConfig.Tracker.TTL + int64(rand.Intn(201)-100),
		"incomplete": scrape.Incomplete,
		"complete":   scrape.Complete,
		"info_hash":  conv.UnsafeBytesToString(conv.Trans9959_1ToUTF8(conv.UnsafeStringToBytes(req.InfoHash))),
	}
	hlog.CtxDebugf(ctx, "send msg:%s", utils.ToJSON(resp))
	if err := conn.WriteJSON(resp); err != nil {
		hlog.CtxErrorf(ctx, "write response error: %s", err.Error())
		return err
	}
	for _, p := range res {
		o := choose.Slice(p.Offers).One()
		offer := hertz.H{
			"action":    "announce",
			"info_hash": conv.UnsafeBytesToString(conv.Trans9959_1ToUTF8(conv.UnsafeStringToBytes(req.InfoHash))),
			"offer_id":  o.OfferID,
			"peer_id":   p.ID,
			"offer":     o.Offer,
		}
		hlog.CtxDebugf(ctx, "send msg:%s", utils.ToJSON(offer))
		if err := conn.WriteJSON(offer); err != nil {
			hlog.CtxErrorf(ctx, "write response error: %s", err.Error())
			return err
		}
	}
	return nil

}

func handleWSScrape(ctx context.Context, req []byte, conn *model.Conn) error {
	infoHashRaw, err := sonic.Get(req, "info_hash")
	if err != nil {
		return err
	}
	infoHashes := make([]string, 0)
	if tryArray, err := infoHashRaw.Array(); err == nil {
		for _, v := range tryArray {
			if s, ok := v.(string); ok {
				infoHashes = append(infoHashes, string(conv.TransUTF8To9959_1(conv.UnsafeStringToBytes(s))))
			}
		}
	} else if tryString, err := infoHashRaw.String(); err == nil {
		infoHashes = append(infoHashes, string(conv.TransUTF8To9959_1(conv.UnsafeStringToBytes(tryString))))
	}
	if len(infoHashes) == 0 {
		return errors.New("info_hash can't be empty")
	}
	ret := make(map[string]*model.ScrapeFile)
	manager := peer.GetPeerManager()
	for _, infoHash := range infoHashes {
		var err error
		ret[infoHash], err = manager.Scrape(ctx, infoHash)
		if err != nil {
			return err
		}
	}
	resp := hertz.H{
		"action": "scrape",
		"files":  ret,
	}
	if err := conn.WriteJSON(resp); err != nil {
		hlog.CtxErrorf(ctx, "write response error: %s", err.Error())
		return err
	}
	return nil
}
