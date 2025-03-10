package udp_server

import (
	"bytes"
	"context"
	"encoding/binary"
	"errors"
	"net"

	"github.com/PBH-BTN/trunker/biz/config"
	"github.com/PBH-BTN/trunker/biz/model"
	"github.com/PBH-BTN/trunker/biz/services/peer"
	"github.com/PBH-BTN/trunker/service/metrics"
	"github.com/PBH-BTN/trunker/utils/bittorrent"
	"github.com/bytedance/gopkg/lang/fastrand"
	"github.com/cloudwego/hertz/pkg/common/hlog"
	"github.com/panjf2000/gnet/v2"
)

func (s *UDPServer) handleAnnounce(ctx context.Context, remote *net.UDPAddr, tid uint32, c gnet.Conn, b []byte) error {
	var req *model.AnnounceRequest
	isIPv6 := false
	if remote.IP.To4() != nil {
		if len(b) < 82 {
			return errors.New("invalid announce request")
		}
		req = parseAnnounceRequestV4(b)
		if req.IPv4 == net.IPv4zero.String() { // some client may not announce ip
			req.IPv4 = remote.IP.String()
		}
	} else {
		isIPv6 = true
		if len(b) < 94 {
			if len(b) >= 82 { // some client send v4 packet to v6, such as qbittorrent
				req = parseAnnounceRequestV4(b)
				req.IPv4 = ""
				req.IPv6 = remote.IP.String()
			} else {
				return errors.New("invalid announce request")
			}
		} else {
			req = parseAnnounceRequestV6(b)
			if req.IPv6 == net.IPv6zero.String() {
				req.IPv6 = remote.IP.String()
			}
		}
	}
	req.ClientIP = remote.IP
	req.Type = model.PeerTypeBittorrent
	req.Source = model.SourceUDP
	go metrics.EmitCounter(metrics.CounterAnnounce, 1, map[string]string{
		metrics.LabelSource: "udp",
		metrics.LabelClient: bittorrent.ParsePeerID(req.PeerID),
	})
	res, err := peer.GetPeerManager().HandleAnnouncePeer(ctx, req)
	if err != nil {
		return err
	}
	scrape, err := peer.GetPeerManager().Scrape(ctx, req.InfoHash)
	if err != nil {
		return err
	}
	buf := bytes.NewBuffer(make([]byte, 0, 8+12+len(res)*18))
	writeHeader(buf, ActionAnnounce, tid)                                                                       // 8
	_ = binary.Write(buf, binary.BigEndian, uint32(config.AppConfig.Tracker.TTL+int64(fastrand.Intn(201)-100))) // interval 4
	_ = binary.Write(buf, binary.BigEndian, uint32(scrape.Incomplete))                                          // leechers 4
	_ = binary.Write(buf, binary.BigEndian, uint32(scrape.Seeder))                                              // seeders 4
	for _, p := range res {
		var ip net.IP
		if isIPv6 {
			ip = p.GetIP().To16()
		} else {
			ip = p.GetIP().To4()
		}
		if ip != nil {
			_ = binary.Write(buf, binary.BigEndian, ip)
			_ = binary.Write(buf, binary.BigEndian, uint16(p.Port))
		}
	}
	_, err = c.Write(buf.Bytes())
	if err != nil {
		hlog.CtxErrorf(ctx, "failed to write to connection:%s", err.Error())
		return nil
	}
	return nil
}

func (s *UDPServer) handleScrape(ctx context.Context, tid uint32, conn gnet.Conn, buf []byte) error {
	n := len(buf)
	infoHash := make([]string, 0, n/20)
	for i := 20; i <= n; i += 20 {
		infoHash = append(infoHash, string(buf[i-20:i]))
	}
	res := bytes.NewBuffer(make([]byte, 0, 8+len(infoHash)*12))
	writeHeader(res, ActionScrape, tid)
	for _, hash := range infoHash {
		scrape, err := peer.GetPeerManager().Scrape(ctx, hash)
		if err != nil {
			return err
		}
		_ = binary.Write(res, binary.BigEndian, uint32(scrape.Seeder))     // seeders 4
		_ = binary.Write(res, binary.BigEndian, uint32(scrape.Complete))   // completed 4
		_ = binary.Write(res, binary.BigEndian, uint32(scrape.Incomplete)) // leechers 4
	}
	_, err := conn.Write(res.Bytes())
	if err != nil {
		hlog.CtxErrorf(ctx, "failed to write to connection:%s", err.Error())
		return nil
	}
	return nil
}
