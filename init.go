package main

import (
	"github.com/PBH-BTN/trunker/biz/config"
	"github.com/PBH-BTN/trunker/biz/services/peer"
	"github.com/PBH-BTN/trunker/service/database"
	"github.com/PBH-BTN/trunker/service/mq/producer"
)

func Init() {
	config.Init()
	if config.AppConfig.Tracker.UseDB {
		database.Init()
	}
	if config.AppConfig.Tracker.EnableEventProducer {
		producer.Init()
	}
	//cache.Init()
	peer.InitPeerManager()
	peer.GetPeerManager().LoadFromPersist()
}
