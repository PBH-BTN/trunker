package handler

import (
	"context"
	"fmt"

	"github.com/PBH-BTN/trunker/biz/services/peer"
	"github.com/PBH-BTN/trunker/utils/http"
	"github.com/cloudwego/hertz/pkg/app"
)

type banInfoHashRequest struct {
	Hash []string `json:"hash"`
}

func HandleBanInfoHash(ctx context.Context, c *app.RequestContext) {
	req := &banInfoHashRequest{}
	if c.Bind(req) != nil {
		http.ResponseBadRequest(c)
		return
	}
	if len(req.Hash) == 0 {
		http.ResponseBadRequest(c)
		return
	}
	manager := peer.GetPeerManager()
	for _, infoHash := range req.Hash {
		manager.BanInfoHash(ctx, infoHash)
	}
	http.ResponseOK(c, fmt.Sprintf("%d info hash banned", len(req.Hash)))
	return
}

func HandleClearBanInfoHash(_ context.Context, c *app.RequestContext) {
	manager := peer.GetPeerManager()
	manager.ClearBanInfoHash()
	http.ResponseOK(c, "all info hash bans cleared")
	return
}

func HandleClearBanPeer(_ context.Context, c *app.RequestContext) {
	manager := peer.GetPeerManager()
	manager.ClearBanPeer()
	http.ResponseOK(c, "all peer bans cleared")
	return
}

type banPeerRequest struct {
	PeerId []string `json:"peer_id"`
}

func HandleBanPeer(ctx context.Context, c *app.RequestContext) {
	req := &banPeerRequest{}
	if c.Bind(req) != nil {
		http.ResponseBadRequest(c)
		return
	}
	if len(req.PeerId) == 0 {
		http.ResponseBadRequest(c)
		return
	}
	manager := peer.GetPeerManager()
	for _, infoHash := range req.PeerId {
		manager.BanPeer(ctx, infoHash)
	}
	http.ResponseOK(c, fmt.Sprintf("%d peer banned", len(req.PeerId)))
	return
}

type getInfoHashPeersReq struct {
	InfoHash string `path:"infoHash" vd:"len($) >0"`
}

func GetInfoHashPeers(ctx context.Context, c *app.RequestContext) {
	req := &getInfoHashPeersReq{}
	if c.BindAndValidate(req) != nil {
		http.ResponseBadRequest(c)
		return
	}
	manager := peer.GetPeerManager()
	peers, err := manager.GetPeers(ctx, req.InfoHash)
	if err != nil {
		http.ResponseErr(c, err)
		return
	}
	http.ResponseOK(c, peers)
}

func DeleteInfoHash(ctx context.Context, c *app.RequestContext) {
	req := &getInfoHashPeersReq{}
	if c.BindAndValidate(req) != nil {
		http.ResponseBadRequest(c)
		return
	}
	manager := peer.GetPeerManager()
	err := manager.DeleteInfoHash(ctx, req.InfoHash)
	if err != nil {
		http.ResponseErr(c, err)
		return
	}
	http.ResponseOK(c, nil)
}
