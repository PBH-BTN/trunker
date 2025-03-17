package websocket

import (
	"context"
	"errors"
	"math/big"

	"github.com/PBH-BTN/trunker/biz/config"
	"github.com/PBH-BTN/trunker/biz/model"
	"github.com/PBH-BTN/trunker/biz/services/peer/common"
	"github.com/PBH-BTN/trunker/biz/services/peer/mux_local/ban"
	"github.com/PBH-BTN/trunker/utils/conv"
	"github.com/cloudwego/hertz/pkg/common/hlog"
)

type WebsocketMuxManager struct {
	localList []*manager
	ban       *ban.Manager
}

func NewWSMuxLocalManager(num int) *WebsocketMuxManager {
	hlog.Info("websocket server enable, shard:", num)
	list := make([]*manager, 0, num)
	for i := 0; i < num; i++ {
		list = append(list, newManager())
	}
	return &WebsocketMuxManager{
		localList: list,
		ban:       ban.NewBanManager(),
	}
}

func (m *WebsocketMuxManager) pickWorker(hashBytes []byte) *manager {
	hashInt := new(big.Int).SetBytes(hashBytes)

	// 取模运算以获得服务索引
	return m.localList[hashInt.Uint64()%uint64(len(m.localList))]

}

func (m *WebsocketMuxManager) HandleAnnouncePeer(ctx context.Context, req *model.AnnounceRequest) ([]*common.Peer, error) {
	// process block list
	if m.ban.Test(ban.BanTypeInfoHash, req.InfoHash) {
		hlog.CtxInfof(ctx, "info hash %s is banned", req.InfoHash)
		return nil, errors.New("banned info_hash")
	}
	if m.ban.Test(ban.BanTypePeerId, req.PeerID) {
		hlog.CtxInfof(ctx, "peer id %s is banned", req.PeerID)
		return nil, errors.New("banned peer_id")
	}

	worker := m.pickWorker(conv.UnsafeStringToBytes(req.InfoHash))
	return worker.HandleAnnouncePeer(ctx, req)
}

func (m *WebsocketMuxManager) Scrape(ctx context.Context, infoHash string) (*model.ScrapeFile, error) {
	worker := m.pickWorker(conv.UnsafeStringToBytes(infoHash))
	return worker.Scrape(ctx, infoHash)
}

var wsManager *WebsocketMuxManager

func InitWSMuxLocalManager() {
	wsManager = NewWSMuxLocalManager(config.AppConfig.Tracker.WSServer.Shard)
}

func (m *WebsocketMuxManager) AnswerToPeer(ctx context.Context, infoHash string, peerID string, answerBody []byte) error {
	worker := m.pickWorker(conv.UnsafeStringToBytes(infoHash))
	return worker.AnswerToPeer(ctx, infoHash, peerID, answerBody)
}

func GetWSManager() *WebsocketMuxManager {
	return wsManager
}
