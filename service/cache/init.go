package cache

import (
	"context"
	"net"
	"time"

	"github.com/PBH-BTN/trunker/biz/config"
	"github.com/cloudwego/netpoll"
	"github.com/redis/go-redis/v9"
)

var Client *redis.Client

func Init() {
	dialer := netpoll.NewDialer()
	Client = redis.NewClient(&redis.Options{
		Addr:    config.AppConfig.Cache.Addr,
		Network: config.AppConfig.Cache.Network,
		Dialer: func(ctx context.Context, network, addr string) (net.Conn, error) {
			return dialer.DialTimeout(network, addr, time.Millisecond*100)
		},
		WriteTimeout: -2,
		ReadTimeout:  -2,
		Password:     "", // no password set
		DB:           0,  // use default DB
	})
}
