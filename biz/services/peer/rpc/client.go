package rpc

import (
	"context"

	"github.com/PBH-BTN/trunker/biz/model"
	"github.com/PBH-BTN/trunker/biz/services/peer/common"
	trunker "github.com/PBH-BTN/trunker/kitex_gen/pbh/btn/trunker/trunkerservice"
	"github.com/cloudwego/kitex/client"
	"github.com/cloudwego/kitex/pkg/remote/codec/thrift"
	"github.com/cloudwego/kitex/transport"
)

type Manager struct {
	c trunker.Client
}

func (m Manager) HandleAnnouncePeer(ctx context.Context, req *model.AnnounceRequest) ([]*common.Peer, error) {
	//TODO implement me
	panic("implement me")
}

func (m Manager) Scrape(ctx context.Context, infoHash string) (*model.ScrapeFile, error) {
	//TODO implement me
	panic("implement me")
}

func (m Manager) GetStatistic(ctx context.Context) (*common.StatisticInfo, error) {
	//TODO implement me
	panic("implement me")
}

func (m Manager) BanInfoHash(ctx context.Context, infoHash string) error {
	//TODO implement me
	panic("implement me")
}

func (m Manager) BanPeer(ctx context.Context, peerID string) error {
	//TODO implement me
	panic("implement me")
}

func (m Manager) ClearBanInfoHash() {
	//TODO implement me
	panic("implement me")
}

func (m Manager) ClearBanPeer() {
	//TODO implement me
	panic("implement me")
}

func (m Manager) GetPeers(ctx context.Context, infoHash string) ([]*common.Peer, error) {
	//TODO implement me
	panic("implement me")
}

func (m Manager) DeleteInfoHash(ctx context.Context, infoHash string) error {
	//TODO implement me
	panic("implement me")
}

func NewManager(target string) *Manager {
	c := trunker.MustNewClient("pbh.btn.trunker",
		client.WithHostPorts(target),
		client.WithPayloadCodec(thrift.NewThriftCodecWithConfig(thrift.FrugalRead|thrift.FrugalWrite)),
		client.WithTransportProtocol(transport.Framed),
	)
	return &Manager{c: c}
}
