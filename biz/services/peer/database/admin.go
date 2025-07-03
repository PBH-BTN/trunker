package database

import (
	"context"
	"encoding/hex"
	"time"

	"github.com/PBH-BTN/trunker/biz/services/peer/common"
	"github.com/PBH-BTN/trunker/biz/services/peer/database/entity"
	"github.com/bytedance/gg/gslice"
	"github.com/bytedance/gopkg/util/logger"
	"github.com/cloudwego/hertz/pkg/common/hlog"
)

func (m *DBManager) BanInfoHash(ctx context.Context, infoHash string) error {
	m.banInfoHashLock.Lock()
	m.banInfoHash.AddString(infoHash)
	m.banInfoHashLock.Unlock()
	err := m.blockListRepo.AddToBlockList(ctx, infoHash, entity.BlockTypeInfoHash)
	if err != nil {
		hlog.CtxErrorf(ctx, "failed to add info_hash to block list: %s", err.Error())
		return err
	}
	return nil
}

func (m *DBManager) BanPeer(ctx context.Context, peerID string) error {
	m.banPeerLock.Lock()
	m.banPeerId.AddString(peerID)
	m.banPeerLock.Unlock()
	err := m.blockListRepo.AddToBlockList(ctx, peerID, entity.BlockTypePeerID)
	if err != nil {
		hlog.CtxErrorf(ctx, "failed to add peer_id to block list: %s", err.Error())
		return err
	}
	return nil
}

func (m *DBManager) ClearBanInfoHash() {
	m.banPeerLock.Lock()
	m.banPeerId.ClearAll()
	m.banPeerLock.Unlock()
}

func (m *DBManager) ClearBanPeer() {
	m.banPeerLock.Lock()
	m.banPeerId.ClearAll()
	m.banPeerLock.Unlock()
}

func (m *DBManager) GetPeers(ctx context.Context, infoHash string) ([]*common.Peer, error) {
	peers, err := m.peerRepo.GetPeers(ctx, infoHash)
	if err != nil {
		hlog.CtxErrorf(ctx, "failed to get peers from db: %s", err.Error())
		return nil, err
	}
	return gslice.Map(peers, DBToCommon), nil

}

func (m *DBManager) DeleteInfoHash(ctx context.Context, infoHash string) error {
	err := m.peerRepo.DeleteInfoHash(ctx, infoHash)
	if err != nil {
		hlog.CtxErrorf(ctx, "failed to delete info hash from db: %s", err.Error())
		return err
	}
	return nil
}

func (m *DBManager) restoreBlockList() {
	logger.Infof("start to restore block list from db")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	infoHashList, err := m.blockListRepo.GetBlockList(ctx, entity.BlockTypeInfoHash)
	if err != nil {
		logger.Errorf("failed to get info hash block list from db: %s", err.Error())
	} else {
		for _, item := range infoHashList {
			target, _ := hex.DecodeString(item.Target)
			m.banInfoHashLock.Lock()
			m.banInfoHash.Add(target)
			m.banInfoHashLock.Unlock()
		}
	}
	peerList, err := m.blockListRepo.GetBlockList(ctx, entity.BlockTypePeerID)
	if err != nil {
		logger.Errorf("failed to get peer id block list from db: %s", err.Error())
	} else {
		for _, item := range peerList {
			target, _ := hex.DecodeString(item.Target)
			m.banPeerLock.Lock()
			m.banPeerId.Add(target)
			m.banPeerLock.Unlock()
		}
	}
	logger.Infof("restore block list from db done, %d info_hash(es), %d peer_id(s) restored", len(infoHashList), len(peerList))
}
