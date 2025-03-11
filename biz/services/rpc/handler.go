package rpc

import (
	"context"

	"github.com/PBH-BTN/trunker/biz/services/peer"
	"github.com/PBH-BTN/trunker/kitex_gen/pbh/btn/trunker"
	"github.com/PBH-BTN/trunker/utils"
)

// TrunkerServiceImpl implements the last service interface defined in the IDL.
type TrunkerServiceImpl struct {
	manager peer.PeerManager
}

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
