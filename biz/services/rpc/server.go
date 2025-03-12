package rpc

import (
	"log"
	"net"

	trunker "github.com/PBH-BTN/trunker/kitex_gen/pbh/btn/trunker/trunkerservice"
	"github.com/cloudwego/hertz/pkg/common/hlog"
	"github.com/cloudwego/kitex/pkg/remote/codec/thrift"
	"github.com/cloudwego/kitex/server"
)

func StartRPCServer(hostPort string) {
	addr, err := net.ResolveTCPAddr("tcp", hostPort)
	if err != nil {
		log.Fatal("invalid rpc listen address:", err.Error())
	}
	svr := trunker.NewServer(newTrunkerServiceImpl(),
		server.WithPayloadCodec(thrift.NewThriftCodecWithConfig(thrift.FrugalRead|thrift.FrugalWrite)),
		server.WithServiceAddr(addr),
	)
	hlog.Info("rpc server is listen at ", hostPort)
	err = svr.Run()

	if err != nil {
		hlog.Fatal("run rpc server error:", err.Error())
	}
}
