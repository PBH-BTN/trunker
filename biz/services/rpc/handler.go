package rpc

import (
	"context"
	"errors"

	"github.com/PBH-BTN/trunker/biz/services/peer"
	"github.com/PBH-BTN/trunker/kitex_gen/pbh/btn/trunker"
	"github.com/PBH-BTN/trunker/utils"
)

// TrunkerServiceImpl implements the last service interface defined in the IDL.
type TrunkerServiceImpl struct {
	manager peer.PeerManager
}

var _ trunker.TrunkerService = (*TrunkerServiceImpl)(nil)

func newTrunkerServiceImpl() *TrunkerServiceImpl {
	return &TrunkerServiceImpl{manager: peer.GetPeerManager()}
}

// Announce implements the TrunkerServiceImpl interface.
func (s *TrunkerServiceImpl) Announce(ctx context.Context, request *trunker.AnnounceRequest) (*trunker.AnnounceResponse, error) {
	resp, err := s.manager.HandleAnnouncePeer(ctx, announceRequestIDLToCommon(request))
	if err != nil {
		return nil, err
	}
	return &trunker.AnnounceResponse{Peers: utils.Map(resp, peerCommonToIDL)}, nil
}

func (s *TrunkerServiceImpl) Scrape(ctx context.Context, request *trunker.ScrapeRequest) (*trunker.ScrapeResponse, error) {
	resp := make(map[string]*trunker.ScrapeFile)
	for _, hash := range request.InfoHashes {
		v, err := s.manager.Scrape(ctx, hash)
		if err != nil {
			return nil, err
		}
		resp[hash] = &trunker.ScrapeFile{
			Complete:   int64(v.Complete),
			Downloaded: int64(v.Downloaded),
			Incomplete: int64(v.Incomplete),
			Seeder:     int64(v.Seeder),
		}
	}
	return &trunker.ScrapeResponse{Res: resp}, nil
}

func (s *TrunkerServiceImpl) GetStatistic(ctx context.Context, request *trunker.GetStatisticRequest) (r *trunker.GetStatisticResponse, err error) {
	resp, err := s.manager.GetStatistic(ctx)
	if err != nil {
		return nil, err
	}
	ret := &trunker.GetStatisticResponse{Info: &trunker.StatisticInfo{
		TotalPeers:    int64(resp.TotalPeers),
		TotalTorrents: int64(resp.TotalTorrents),
		Shards:        make(map[string]*trunker.StatisticInfo),
	}}
	for k, v := range resp.Shards {
		ret.Info.Shards[k] = &trunker.StatisticInfo{
			TotalPeers:    int64(v.TotalPeers),
			TotalTorrents: int64(v.TotalTorrents),
		}
	}
	return ret, nil
}

func (s *TrunkerServiceImpl) Ban(ctx context.Context, request *trunker.BanRequest) (*trunker.BanResponse, error) {
	var err error
	switch request.Type {
	case trunker.BanType_InfoHash:
		err = s.manager.BanInfoHash(ctx, request.Target)
	case trunker.BanType_PeerID:
		err = s.manager.BanPeer(ctx, request.Target)
	default:
		err = errors.New("unknown type")
	}
	return &trunker.BanResponse{}, err
}

func (s *TrunkerServiceImpl) DeleteInfoHash(ctx context.Context, request *trunker.DeleteInfoHashRequest) (r *trunker.DeleteInfoHashResponse, err error) {
	return &trunker.DeleteInfoHashResponse{}, s.manager.DeleteInfoHash(ctx, request.Target)
}

func (s *TrunkerServiceImpl) GetPeer(ctx context.Context, request *trunker.GetPeerRequest) (r *trunker.GetPeerResponse, err error) {
	res, err := s.manager.GetPeers(ctx, request.InfoHash)
	if err != nil {
		return nil, err
	}
	return &trunker.GetPeerResponse{
		Peers: utils.Map(res, peerCommonToIDL),
	}, nil
}
