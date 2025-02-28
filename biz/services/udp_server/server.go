package udp_server

import (
	"context"
	"encoding/binary"
	"errors"
	"net"
	"runtime/debug"
	"time"

	"github.com/PBH-BTN/trunker/service/metrics"
	"github.com/cloudwego/hertz/pkg/common/hlog"
	"github.com/hitoshi44/go-uid64"
	"github.com/panjf2000/gnet/v2"
	"github.com/zhangyunhao116/skipmap"
)

type connection struct {
	id         uint64
	remoteAddr net.Addr
	time       time.Time
}

type UDPServer struct {
	gnet.BuiltinEventEngine

	eng            gnet.Engine
	id             *uid64.Generator
	connectionList *skipmap.Uint64Map[*connection]
}

func (s *UDPServer) OnBoot(eng gnet.Engine) (action gnet.Action) {
	s.eng = eng
	return gnet.None
}

func (s *UDPServer) OnTick() (delay time.Duration, action gnet.Action) {
	go s.cleanConnection()
	return time.Minute * 2, gnet.None
}

func NewUDPServer() *UDPServer {
	generator, _ := uid64.NewGenerator(0)
	s := &UDPServer{
		id:             generator,
		connectionList: skipmap.NewUint64[*connection](),
	}
	return s
}
func (s *UDPServer) OnTraffic(conn gnet.Conn) gnet.Action {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	start := time.Now()
	defer func() {
		if err := recover(); err != nil {
			hlog.CtxErrorf(ctx, "panic:%s", err)
			println(string(debug.Stack()))
		}
	}()
	err := s.handleRequest(ctx, conn)
	cost := time.Now().Sub(start)
	if err != nil {
		hlog.CtxErrorf(ctx, "failed to handle request:%s", err.Error())
		metrics.EmitCounter(metrics.CounterUDPRequestError, 1, map[string]string{
			metrics.LabelReason: err.Error(),
			metrics.LabelAction: "unknown",
		})
		if conn.Context() == nil {
			conn.SetContext("unknown")
		}
	}

	metrics.EmitCounter(metrics.CounterUDPRequest, 1, map[string]string{
		metrics.LabelAction: conn.Context().(string),
	})
	metrics.ObserveHistogram(metrics.HistogramLatency, float64(cost.Microseconds()), map[string]string{
		metrics.LabelAction: conn.Context().(string),
	})

	return gnet.None
}

func (s *UDPServer) handleRequest(ctx context.Context, conn gnet.Conn) error {
	headerBuf, err := conn.Next(-1) // connection_id or protocol_id
	if err != nil || len(headerBuf) < 16 {
		return errors.New("invalid request")
	}
	connectionID := binary.BigEndian.Uint64(headerBuf[:8])
	action := binary.BigEndian.Uint32(headerBuf[8:12])
	transactionID := binary.BigEndian.Uint32(headerBuf[12:16])
	if connectionID == ProtocolID && action == ActionConnect {
		conn.SetContext("connection")
		if err := s.handleConnection(ctx, conn.RemoteAddr(), transactionID, conn); err != nil {
			hlog.CtxErrorf(ctx, "failed to handle connection:%s", err.Error())
			return err
		}
		return nil
	}

	if c, ok := s.connectionList.Load(connectionID); ok {
		if c.remoteAddr.String() != conn.RemoteAddr().String() {
			responseError(conn, transactionID, errors.New("connection mismatch"))
			return nil
		}
		var err error
		switch action {
		case ActionAnnounce:
			conn.SetContext("announce")
			err = s.handleAnnounce(ctx, conn.RemoteAddr().(*net.UDPAddr), transactionID, conn, headerBuf[16:])
		case ActionScrape:
			conn.SetContext("scrape")
			err = s.handleScrape(ctx, transactionID, conn, headerBuf[16:])
		default:
			conn.SetContext("unknown")
			err = errors.New("invalid action")
		}
		if err != nil {
			responseError(conn, transactionID, err)
		}
		return nil
	} else {
		responseError(conn, transactionID, errors.New("connection expired"))
		return nil
	}

}
