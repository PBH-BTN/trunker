package router

import (
	"github.com/PBH-BTN/trunker/biz/handler"
	"github.com/PBH-BTN/trunker/biz/middleware"
	"github.com/cloudwego/hertz/pkg/route"
)

func RegisterAdminRouter(r route.IRouter) {
	g := r.Group("admin")
	g.Use(middleware.GetAdminAuthMiddleware()...)
	g.GET("/statistic", handler.Statistic)
	g.PUT("/ban/info_hash", handler.HandleBanInfoHash)
	g.PUT("/ban/peer", handler.HandleBanPeer)
	g.DELETE("/ban/info_hash", handler.HandleClearBanInfoHash)
	g.DELETE("/ban/peer", handler.HandleClearBanPeer)

	// IP ban management
	g.PUT("/ban/ip", handler.HandleBanIP)
	g.DELETE("/ban/ip", handler.HandleUnbanIP)
	g.DELETE("/clear/ip/ban", handler.HandleClearBanIP)
	g.GET("/ip/filter/stats", handler.HandleGetIPFilterStats)

	g.GET("/info_hash/:infoHash/peers", handler.GetInfoHashPeers)
	g.DELETE("/info_hash/:infoHash", handler.DeleteInfoHash)
}
