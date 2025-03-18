package cache

import (
	"github.com/PBH-BTN/trunker/biz/config"
	"github.com/redis/go-redis/v9"
)

var Client *redis.Client

func Init() {
	Client = redis.NewClient(&redis.Options{
		Addr:     config.AppConfig.Cache.Addr,
		Network:  config.AppConfig.Cache.Network,
		Password: "", // no password set
		DB:       0,  // use default DB
	})
}
