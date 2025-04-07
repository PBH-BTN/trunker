package udp_server

import (
	"bytes"
	"context"
	"encoding/binary"
	"net"
	"time"

	"github.com/cloudwego/hertz/pkg/common/hlog"
	"github.com/panjf2000/gnet/v2"
)

/*
connect request:

Offset  Size            Name            Value
0       64-bit integer  protocol_id     0x41727101980 // magic constant
8       32-bit integer  action          0 // connect
12      32-bit integer  transaction_id
16
connect response:

Offset  Size            Name            Value
0       32-bit integer  action          0 // connect
4       32-bit integer  transaction_id
8       64-bit integer  connection_id
*/
func (s *UDPServer) handleConnection(ctx context.Context, remote net.Addr, tid uint32, conn gnet.Conn) error {
	newId, err := s.id.Gen()
	if err != nil {
		hlog.CtxErrorf(ctx, "failed to generate connection id:%s", err.Error())
		return err
	}
	cid := uint64(newId.ToInt())
	s.connectionList.LoadOrStore(cid, &connection{id: cid, remoteAddr: remote, time: time.Now()})
	// send response
	buf := bytes.NewBuffer(make([]byte, 0, 16))
	writeHeader(buf, ActionConnect, tid)
	_ = binary.Write(buf, binary.BigEndian, cid)
	_, err = conn.Write(buf.Bytes())
	if err != nil {
		return err
	}
	return nil
}

const connectionTimeout = 2 * time.Minute

// cleanConnection clean expire connection
func (s *UDPServer) cleanConnection() {
	t := time.Now()
	count := 0
	s.connectionList.Range(func(id uint64, conn *connection) bool {
		if conn.time.Add(connectionTimeout).Before(t) {
			s.connectionList.Delete(id)
			count++
		}
		return true
	})
	hlog.Info(count, " expired connections cleaned.")
}

func responseError(c gnet.Conn, tid uint32, err error) {
	reason := err.Error()
	buf := bytes.NewBuffer(make([]byte, 0, 8+len(reason)))
	writeHeader(buf, ActionError, tid)
	buf.WriteString(reason)
	_, _ = c.Write(buf.Bytes())
}
