package handler

import (
	"context"
	"fmt"

	"github.com/PBH-BTN/trunker/biz/services/peer"
	"github.com/PBH-BTN/trunker/utils/conv"
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
		err := manager.BanInfoHash(ctx, conv.UnsafeBytesToString(conv.TransUTF8To8859_1(conv.UnsafeStringToBytes(infoHash))))
		if err != nil {
			http.ResponseErr(c, err)
			return
		}
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
	for _, peerId := range req.PeerId {
		err := manager.BanPeer(ctx, conv.UnsafeBytesToString(conv.TransUTF8To8859_1(conv.UnsafeStringToBytes(peerId))))
		if err != nil {
			http.ResponseErr(c, err)
			return
		}
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

// IP Ban Management

type banIPRequest struct {
	IPs []string `json:"ips"`
}

func HandleBanIP(ctx context.Context, c *app.RequestContext) {
	req := &banIPRequest{}
	if c.Bind(req) != nil {
		http.ResponseBadRequest(c)
		return
	}
	if len(req.IPs) == 0 {
		http.ResponseBadRequest(c)
		return
	}
	manager := peer.GetPeerManager()
	for _, ip := range req.IPs {
		err := manager.BanIP(ctx, ip)
		if err != nil {
			http.ResponseErr(c, err)
			return
		}
	}
	http.ResponseOK(c, fmt.Sprintf("%d IP(s) banned", len(req.IPs)))
}

func HandleUnbanIP(ctx context.Context, c *app.RequestContext) {
	req := &banIPRequest{}
	if c.Bind(req) != nil {
		http.ResponseBadRequest(c)
		return
	}
	if len(req.IPs) == 0 {
		http.ResponseBadRequest(c)
		return
	}
	manager := peer.GetPeerManager()
	for _, ip := range req.IPs {
		err := manager.UnbanIP(ctx, ip)
		if err != nil {
			http.ResponseErr(c, err)
			return
		}
	}
	http.ResponseOK(c, fmt.Sprintf("%d IP(s) unbanned", len(req.IPs)))
}

func HandleClearBanIP(_ context.Context, c *app.RequestContext) {
	manager := peer.GetPeerManager()
	err := manager.ClearBanIP()
	if err != nil {
		http.ResponseErr(c, err)
		return
	}
	http.ResponseOK(c, "all IP bans cleared")
}

type ipFilterStatsResponse struct {
	TotalIPs       uint64 `json:"total_ips"`
	PacketsDropped uint64 `json:"packets_dropped"`
	PacketsAllowed uint64 `json:"packets_allowed"`
	IsXDP          bool   `json:"is_xdp"`
	InterfaceName  string `json:"interface_name,omitempty"`
}

func HandleGetIPFilterStats(_ context.Context, c *app.RequestContext) {
	manager := peer.GetPeerManager()
	stats := manager.GetIPFilterStats()
	if stats == nil {
		http.ResponseErr(c, fmt.Errorf("IP filter not available"))
		return
	}
	http.ResponseOK(c, &ipFilterStatsResponse{
		TotalIPs:       stats.TotalIPs,
		PacketsDropped: stats.PacketsDropped,
		PacketsAllowed: stats.PacketsAllowed,
		IsXDP:          stats.IsXDP,
		InterfaceName:  stats.InterfaceName,
	})
}
