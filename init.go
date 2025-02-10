package main

import (
	"github.com/PBH-BTN/trunker/biz/config"
	"github.com/PBH-BTN/trunker/biz/services/peer"
	"github.com/PBH-BTN/trunker/service/metrics"
	"github.com/PBH-BTN/trunker/service/mq/producer"
)

func Init() {
	config.Init()
	if config.AppConfig.Tracker.EnableEventProducer {
		producer.Init()
	}
	metrics.Init()
	//cache.Init()
	peer.InitPeerManager()
}
