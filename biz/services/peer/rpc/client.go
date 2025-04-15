package rpc

import (
	"context"
	"time"
	"unsafe"

	"github.com/PBH-BTN/trunker/biz/model"
	"github.com/PBH-BTN/trunker/biz/services/peer/common"
	"github.com/PBH-BTN/trunker/kitex_gen/pbh/btn/trunker"
	"github.com/PBH-BTN/trunker/kitex_gen/pbh/btn/trunker/trunkerservice"
	"github.com/PBH-BTN/trunker/utils"
	"github.com/cloudwego/hertz/pkg/common/hlog"
	"github.com/cloudwego/kitex/client"
	"github.com/cloudwego/kitex/pkg/remote/codec/thrift"
	"github.com/cloudwego/kitex/transport"
)

type Manager struct {
	c trunkerservice.Client
}

func (m Manager) HandleAnnouncePeer(ctx context.Context, req *model.AnnounceRequest) ([]*common.Peer, error) {
	resp, err := m.c.Announce(ctx, announceRequestCommonToIDL(req))
	if err != nil {
		hlog.CtxErrorf(ctx, "remote call announce error:%s", err.Error())
		return nil, err
	}
	return utils.Map(resp.Peers, PeerIDLToCommon), nil
}

func (m Manager) Scrape(ctx context.Context, infoHash string) (*model.ScrapeFile, error) {
	resp, err := m.c.Scrape(ctx, &trunker.ScrapeRequest{InfoHashes: unsafe.Slice(&infoHash, 1)})
	if err != nil {
		hlog.CtxErrorf(ctx, "remote call scrape error:%s", err.Error())
		return nil, err
	}
	return &model.ScrapeFile{
		Seeder:     int(resp.Res[infoHash].Seeder),
		Complete:   int(resp.Res[infoHash].Complete),
		Incomplete: int(resp.Res[infoHash].Incomplete),
		Downloaded: int(resp.Res[infoHash].Downloaded),
	}, nil
}

func (m Manager) GetStatistic(ctx context.Context) (*common.StatisticInfo, error) {
	resp, err := m.c.GetStatistic(ctx, &trunker.GetStatisticRequest{})
	if err != nil {
		hlog.CtxErrorf(ctx, "remote call GetStatistic error:%s", err.Error())
		return nil, err
	}
	ret := &common.StatisticInfo{
		TotalPeers:    uint64(resp.Info.TotalPeers),
		TotalTorrents: uint64(resp.Info.TotalTorrents),
		Shards:        make(map[string]*common.StatisticInfo),
	}
	for k, shard := range resp.Info.Shards {
		ret.Shards[k] = &common.StatisticInfo{
			TotalPeers:    uint64(shard.TotalPeers),
			TotalTorrents: uint64(shard.TotalTorrents),
		}
	}
	return ret, nil
}

func (m Manager) BanInfoHash(ctx context.Context, infoHash string) error {
	_, err := m.c.Ban(ctx, &trunker.BanRequest{
		Type:   trunker.BanType_InfoHash,
		Target: infoHash,
	})
	if err != nil {
		hlog.CtxErrorf(ctx, "remote call ban error:%s", err.Error())
		return err
	}
	return nil
}

func (m Manager) BanPeer(ctx context.Context, peerID string) error {
	_, err := m.c.Ban(ctx, &trunker.BanRequest{
		Type:   trunker.BanType_PeerID,
		Target: peerID,
	})
	if err != nil {
		hlog.CtxErrorf(ctx, "remote call ban error:%s", err.Error())
		return err
	}
	return nil
}

func (m Manager) GetPeers(ctx context.Context, infoHash string) ([]*common.Peer, error) {
	res, err := m.c.GetPeer(ctx, &trunker.GetPeerRequest{InfoHash: infoHash})
	if err != nil {
		hlog.CtxErrorf(ctx, "remote call GetPeer error:%s", err.Error())
		return nil, err
	}
	return utils.Map(res.Peers, PeerIDLToCommon), nil
}

func (m Manager) DeleteInfoHash(ctx context.Context, infoHash string) error {
	_, err := m.c.DeleteInfoHash(ctx, &trunker.DeleteInfoHashRequest{Target: infoHash})
	if err != nil {
		hlog.CtxErrorf(ctx, "remote call DeleteInfoHash error:%s", err.Error())
		return err
	}
	return nil
}

func NewManager(target string) *Manager {
	c := trunkerservice.MustNewClient("pbh.btn.trunker",
		client.WithHostPorts(target),
		client.WithRPCTimeout(5*time.Second),
		client.WithConnectTimeout(3*time.Second),
		client.WithPayloadCodec(thrift.NewThriftCodecWithConfig(thrift.FrugalRead|thrift.FrugalWrite)),
		client.WithTransportProtocol(transport.Framed),
	)
	return &Manager{c: c}
}
