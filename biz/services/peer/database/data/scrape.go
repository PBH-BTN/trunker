package data

import (
	"context"
	"time"

	"github.com/PBH-BTN/trunker/biz/config"
	"github.com/PBH-BTN/trunker/biz/services/peer/database/entity"
)

func (r *PeerRepository) GetDownloadedCount(ctx context.Context, infoHash string) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&entity.Peers{}).Where("info_hash = ? AND event = ?", infoHash, 3).Count(&count).Error
	if err != nil {
		return 0, err
	}
	return count, nil
}

func (r *PeerRepository) GetCompleteCount(ctx context.Context, infoHash string) (int64, error) {
	var count int64
	validTime := time.Now().Add(time.Duration(-1*config.AppConfig.Tracker.TTL) * time.Second)
	err := r.db.WithContext(ctx).Model(&entity.Peers{}).Where("info_hash = ? AND last_seen > ? AND `left` = ?", infoHash, validTime, 0).Count(&count).Error
	if err != nil {
		return 0, err
	}
	return count, nil
}

func (r *PeerRepository) GetSeederCount(ctx context.Context, infoHash string) (int64, error) {
	var count int64
	validTime := time.Now().Add(time.Duration(-1*config.AppConfig.Tracker.TTL) * time.Second)
	err := r.db.WithContext(ctx).Model(&entity.Peers{}).Where("info_hash = ? AND last_seen > ? AND `left` = ? AND event <> ?", infoHash, validTime, 0, 2).Count(&count).Error
	if err != nil {
		return 0, err
	}
	return count, nil
}

func (r *PeerRepository) GetInCompleteCount(ctx context.Context, infoHash string) (int64, error) {
	var count int64
	validTime := time.Now().Add(time.Duration(-1*config.AppConfig.Tracker.TTL) * time.Second)
	err := r.db.WithContext(ctx).Model(&entity.Peers{}).Where("info_hash = ? AND last_seen > ? AND `left` > 0 ", infoHash, validTime).Count(&count).Error
	if err != nil {
		return 0, err
	}
	return count, nil
}

func (r *PeerRepository) GetPeersCount(ctx context.Context) (int64, error) {
	var count int64
	validTime := time.Now().Add(time.Duration(-1*config.AppConfig.Tracker.TTL) * time.Second)
	err := r.db.WithContext(ctx).Model(&entity.Peers{}).Where("last_seen > ?", validTime).Count(&count).Error
	if err != nil {
		return 0, err
	}
	return count, nil
}

func (r *PeerRepository) GetInfoHashCount(ctx context.Context) (int64, error) {
	var count int64
	validTime := time.Now().Add(time.Duration(-1*config.AppConfig.Tracker.TTL) * time.Second)
	err := r.db.WithContext(ctx).Model(&entity.Peers{}).Distinct("info_hash").Where("last_seen > ?", validTime).Count(&count).Error
	if err != nil {
		return 0, err
	}
	return count, nil
}
