package main

import (
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/PBH-BTN/trunker/biz/config"
	"github.com/PBH-BTN/trunker/biz/services/peer"
	"github.com/PBH-BTN/trunker/biz/services/peer/websocket"
	"github.com/PBH-BTN/trunker/biz/services/rpc"
	"github.com/PBH-BTN/trunker/biz/services/udp_server"
	"github.com/PBH-BTN/trunker/service/cache"
	"github.com/PBH-BTN/trunker/service/metrics"
	"github.com/PBH-BTN/trunker/service/mq/producer"
	"github.com/cloudwego/hertz/pkg/app/server"
	hertzConfig "github.com/cloudwego/hertz/pkg/common/config"
	"github.com/cloudwego/hertz/pkg/common/hlog"
	prometheus "github.com/hertz-contrib/monitor-prometheus"
	"github.com/panjf2000/gnet/v2"
)

func Init() {
	config.Init()
	if config.AppConfig.Tracker.EnableEventProducer {
		producer.Init()
	}
	if config.AppConfig.Cache.Enable {
		cache.Init()
	}
	if config.AppConfig.Tracker.WSServer.Enable {
		websocket.InitWSMuxLocalManager()
	}
	initPprof()
	metrics.Init()
	//cache.Init()
	peer.InitPeerManager()
	if config.AppConfig.Tracker.UDPServer.Enable {
		go initUDPServer()
	}
	if config.AppConfig.Tracker.RPC.EnableServer {
		go rpc.StartRPCServer(config.AppConfig.Tracker.RPC.HostPorts)
	}

}

func getServerOption() []hertzConfig.Option {
	options := []hertzConfig.Option{
		server.WithHostPorts(config.AppConfig.Tracker.HostPorts),
		server.WithExitWaitTime(time.Minute),
	}
	if config.AppConfig.Tracker.EnableMetrics {
		options = append(options, server.WithTracer(
			prometheus.NewServerTracer(config.AppConfig.Tracker.MetricsHostPorts, "/metrics",
				prometheus.WithDefaultServerMux(true),
				prometheus.WithEnableGoCollector(true),
				prometheus.WithRegistry(metrics.GetRegistry()),
			),
		))
		hlog.Info("metrics is listen at ", config.AppConfig.Tracker.MetricsHostPorts)
	}
	if config.AppConfig.Tracker.UseUnixSocket {
		options = append(options, server.WithNetwork("unix"))
	}
	return options
}

func initUDPServer() {
	log.Fatal(gnet.Run(udp_server.NewUDPServer(), "udp://"+config.AppConfig.Tracker.UDPServer.HostPorts, gnet.WithMulticore(true), gnet.WithTicker(true), gnet.WithLogger(hlog.DefaultLogger())))
}

func initPprof() {
	if config.AppConfig.Tracker.DebugPort > 0 {
		go func() {
			hlog.Info("debug port listening on address=localhost:", config.AppConfig.Tracker.DebugPort)
			_ = http.ListenAndServe("localhost:"+strconv.Itoa(int(config.AppConfig.Tracker.DebugPort)), nil)
		}()
	}

}
