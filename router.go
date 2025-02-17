package main

import (
	"github.com/PBH-BTN/trunker/biz/handler"
	"github.com/PBH-BTN/trunker/biz/router"
	"github.com/cloudwego/hertz/pkg/app/server"
)

// customizeRegister registers customize routers.
func customizedRegister(r *server.Hertz) {
	r.HEAD("/ping", handler.Ping)
	r.GET("/ping", handler.Ping)
	r.GET("/announce", handler.Announce)
	r.GET("/scrape", handler.Scrape)
	router.RegisterAdminRouter(r)
	// your code ...
}
