package data

import (
	"context"
	"encoding/hex"
	"time"

	"github.com/PBH-BTN/trunker/biz/config"
	"github.com/PBH-BTN/trunker/biz/services/peer/common"
	"github.com/PBH-BTN/trunker/biz/services/peer/database/entity"
	"github.com/PBH-BTN/trunker/utils/conv"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type PeerRepository struct {
	db *gorm.DB
}

func NewPeerRepository(db *gorm.DB) *PeerRepository {
	return &PeerRepository{db: db}
}

// PickPeers randomly picks numWant peers from the database
func (r *PeerRepository) PickPeers(ctx context.Context, infoHashRaw string, numWant int, peerType common.PeerType) ([]*entity.Peers, error) {
	var peers []*entity.Peers
	infoHash := hex.EncodeToString(conv.UnsafeStringToBytes(infoHashRaw))
	validTime := time.Now().Add(time.Duration(-1*config.AppConfig.Tracker.TTL) * time.Second)
	err := r.db.WithContext(ctx).Where("info_hash = ? AND last_seen > ? AND event != ? AND type = ?", infoHash, validTime, common.PeerEvent_Stopped, peerType).Order("RAND()").Limit(numWant).Find(&peers).Error
	if err != nil {
		return nil, err
	}
	return peers, nil
}

func (r *PeerRepository) SavePeer(ctx context.Context, peer *entity.Peers) error {
	return r.db.WithContext(ctx).Clauses(clause.OnConflict{
		DoUpdates: clause.AssignmentColumns([]string{
			`ip`,
			`ipv4`,
			`ipv6`,
			`client_ip`,
			`port`,
			"`left`",
			`uploaded`,
			`downloaded`,
			`last_seen`,
			`user_agent`,
			`event`,
			`offers`,
			`updated_at`,
		}),
	}).Save(peer).Error
}
