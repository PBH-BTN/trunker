package main

import (
	"context"
	"os"
	"time"

	appConfig "github.com/PBH-BTN/trunker/biz/config"
	"github.com/PBH-BTN/trunker/biz/middleware"
	"github.com/PBH-BTN/trunker/biz/services/interval"
	"github.com/PBH-BTN/trunker/biz/services/peer"
	"github.com/PBH-BTN/trunker/service/metrics"
	"github.com/cloudwego/hertz/pkg/app/server"
	"github.com/cloudwego/hertz/pkg/common/config"
	"github.com/cloudwego/hertz/pkg/common/hlog"
	prometheus "github.com/hertz-contrib/monitor-prometheus"
	"github.com/hertz-contrib/pprof"
)

func main() {
	if os.Getenv("RUN_ENV") == "prod" {
		hlog.SetLevel(hlog.LevelInfo)
	}
	Init()
	options := []config.Option{
		server.WithTracer(
			prometheus.NewServerTracer(":9091", "/metrics",
				prometheus.WithDefaultServerMux(true),
				prometheus.WithEnableGoCollector(true),
				prometheus.WithRegistry(metrics.GetRegistry()),
			),
		),
		server.WithHostPorts(appConfig.AppConfig.Tracker.HostPorts),
		server.WithExitWaitTime(time.Minute),
	}
	if appConfig.AppConfig.Tracker.UseUnixSocket {
		options = append(options, server.WithNetwork("unix"))
	}
	h := server.Default(options...)
	pprof.Register(h)
	register(h)
	h.Use(middleware.LogSlowQuery)

	h.Engine.OnShutdown = append(h.Engine.OnShutdown, func(_ context.Context) {
		// here save current data
		peer.GetPeerManager().StoreToPersist()
	})
	interval.StartIntervalTask()
	h.Spin()
}
