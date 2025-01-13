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

func HandleBanInfoHash(_ context.Context, c *app.RequestContext) {
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
		manager.BanInfoHash(infoHash)
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

func HandleBanPeer(_ context.Context, c *app.RequestContext) {
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
		manager.BanPeer(infoHash)
	}
	http.ResponseOK(c, fmt.Sprintf("%d peer banned", len(req.PeerId)))
	return
}

type getInfoHashPeersReq struct {
	InfoHash string `path:"infoHash" vd:"len($) >0"`
}

func GetInfoHashPeers(_ context.Context, c *app.RequestContext) {
	req := &getInfoHashPeersReq{}
	if c.BindAndValidate(req) != nil {
		http.ResponseBadRequest(c)
		return
	}
	manager := peer.GetPeerManager()
	peers := manager.GetPeers(req.InfoHash)
	http.ResponseOK(c, peers)
}

func DeleteInfoHash(_ context.Context, c *app.RequestContext) {
	req := &getInfoHashPeersReq{}
	if c.BindAndValidate(req) != nil {
		http.ResponseBadRequest(c)
		return
	}
	manager := peer.GetPeerManager()
	manager.DeleteInfoHash(req.InfoHash)
	http.ResponseOK(c, nil)
}
