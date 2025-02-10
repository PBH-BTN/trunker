package data

import (
	"context"
	"encoding/hex"
	"time"

	"github.com/PBH-BTN/trunker/biz/config"
	"github.com/PBH-BTN/trunker/biz/services/peer/database/entity"
	"github.com/PBH-BTN/trunker/utils/conv"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type BlockListRepository struct {
	db *gorm.DB
}

func NewBlockListRepository(db *gorm.DB) *BlockListRepository {
	return &BlockListRepository{db: db}
}

func (r *BlockListRepository) AddToBlockList(ctx context.Context, targetRaw string, blockType entity.BlockType) error {
	target := hex.EncodeToString(conv.UnsafeStringToBytes(targetRaw))
	block := &entity.BlockList{
		Target:    target,
		Type:      blockType,
		CreatedAt: time.Now(),
	}
	return r.db.WithContext(ctx).Clauses(clause.OnConflict{DoNothing: true}).Create(block).Error
}

func (r *BlockListRepository) GetBlockList(ctx context.Context, blockType entity.BlockType) ([]*entity.BlockList, error) {
	list := make([]*entity.BlockList, 0)
	err := r.db.WithContext(ctx).Where("type = ?", blockType).Find(&list).Error
	if err != nil {
		return nil, err
	}
	return list, nil
}

// GetPeers returns all peers for a given info hash, this is only used for admin
func (r *PeerRepository) GetPeers(ctx context.Context, infoHashRaw string) ([]*entity.Peers, error) {
	var peers []*entity.Peers
	infoHash := hex.EncodeToString(conv.UnsafeStringToBytes(infoHashRaw))
	validTime := time.Now().Add(time.Duration(-1*config.AppConfig.Tracker.TTL) * time.Second)
	err := r.db.WithContext(ctx).Where("info_hash = ? AND last_seen > ? ", infoHash, validTime).Find(&peers).Error
	if err != nil {
		return nil, err
	}
	return peers, nil
}

func (r *PeerRepository) DeleteInfoHash(ctx context.Context, infoHashRaw string) error {
	infoHash := hex.EncodeToString(conv.UnsafeStringToBytes(infoHashRaw))
	return r.db.WithContext(ctx).Where("info_hash = ?", infoHash).Delete(&entity.Peers{}).Error
}
