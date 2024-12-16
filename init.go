package main

import (
	"github.com/PBH-BTN/trunker/biz/config"
	"github.com/PBH-BTN/trunker/biz/services/peer"
	"github.com/PBH-BTN/trunker/service/database"
)

func Init() {
	config.Init()
	if config.AppConfig.Tracker.UseDB {
		database.Init()
	}
	//cache.Init()
	peer.InitPeerManager()
	peer.GetPeerManager().LoadFromPersist()
}
