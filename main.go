package main

import (
	"context"
	"os"

	"github.com/PBH-BTN/trunker/biz/middleware"
	"github.com/PBH-BTN/trunker/biz/services/interval"
	"github.com/PBH-BTN/trunker/biz/services/peer"
	"github.com/cloudwego/hertz/pkg/app/server"
	"github.com/cloudwego/hertz/pkg/common/hlog"
	"github.com/hertz-contrib/pprof"
)

func main() {
	if os.Getenv("RUN_ENV") == "prod" {
		hlog.SetLevel(hlog.LevelInfo)
	}
	Init()
	h := server.Default(getServerOption()...)
	pprof.Register(h)
	register(h)
	h.Use(middleware.LogSlowQuery)
	h.Use(middleware.AdaptiveLimit())
	h.Engine.OnShutdown = append(h.Engine.OnShutdown, func(_ context.Context) {
		// here save current data
		peer.GetPeerManager().StoreToPersist()
	})
	interval.StartIntervalTask()
	h.Spin()
}
