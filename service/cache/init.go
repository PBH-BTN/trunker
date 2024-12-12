package cache

import (
	"fmt"
	"github.com/PBH-BTN/trunker/biz/config"
	"github.com/redis/go-redis/v9"
)

var Client *redis.Client

func Init() {
	Client = redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%d", config.AppConfig.Cache.Host, config.AppConfig.Cache.Port),
		Password: "", // no password set
		DB:       0,  // use default DB
	})
}
