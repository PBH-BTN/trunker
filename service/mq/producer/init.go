package producer

import (
	"github.com/PBH-BTN/trunker/biz/config"
	"github.com/apache/rocketmq-client-go/v2"
	"github.com/apache/rocketmq-client-go/v2/producer"
)

var Producer rocketmq.Producer

func Init() {
	var err error
	Producer, err = rocketmq.NewProducer(
		producer.WithNameServer([]string{config.AppConfig.RocketMq.Endpoint}),
		producer.WithRetry(2),
	)
	if err != nil {
		panic(err)
	}
	err = Producer.Start()
	if err != nil {
		panic(err)
	}
}
