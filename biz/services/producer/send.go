package producer

import (
	"context"
	"net"

	"github.com/PBH-BTN/trunker/biz/config"
	"github.com/PBH-BTN/trunker/biz/services/peer/common"
	"github.com/PBH-BTN/trunker/service/mq/producer"
	"github.com/apache/rocketmq-client-go/v2/primitive"
	"github.com/cloudwego/hertz/pkg/common/hlog"
	"github.com/shamaton/msgpack/v2"
)

type ip struct {
	ClientIP net.IP `msgpack:"client_ip"`
	ReportV4 net.IP `msgpack:"report_v4"`
	ReportV6 net.IP `msgpack:"report_v6"`
	ReportIP net.IP `msgpack:"report_ip"`
}
type eventBody struct {
	InfoHash   string `msgpack:"info_hash"`
	IP         ip     `msgpack:"ip"`
	Port       int    `msgpack:"port"`
	Left       int    `msgpack:"left"`
	Uploaded   int    `msgpack:"uploaded"`
	Downloaded int    `msgpack:"downloaded"`
	LastSeen   int64  `msgpack:"last_seen"`
	UserAgent  string `msgpack:"user_agent"`
	Event      string `msgpack:"event"`
}

func SendPeerEvent(ctx context.Context, infoHash string, peer *common.Peer) {
	if producer.Producer == nil {
		return
	}
	body, _ := msgpack.Marshal(eventBody{
		InfoHash: infoHash,
		IP: ip{
			ClientIP: peer.ClientIP,
			ReportV4: peer.IPv4,
			ReportV6: peer.IPv6,
			ReportIP: peer.IP,
		},
		Port:       peer.Port,
		Left:       peer.Left,
		Uploaded:   peer.Uploaded,
		Downloaded: peer.Downloaded,
		LastSeen:   peer.LastSeen.Unix(),
		UserAgent:  peer.UserAgent,
		Event:      peer.Event.String(),
	})
	err := producer.Producer.SendOneWay(ctx, &primitive.Message{
		Topic: config.AppConfig.RocketMq.Topic,
		Body:  body,
	})
	if err != nil {
		hlog.CtxErrorf(ctx, "failed to send to mq:%s", err.Error())
	}
}
