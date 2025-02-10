package entity

import "time"

type BlockType int8

const (
	BlockTypeInfoHash BlockType = 0
	BlockTypePeerID   BlockType = 1
)

type BlockList struct {
	ID        uint64    `json:"id" `
	Target    string    `json:"target" ` // banned target
	Type      BlockType `json:"type" `   // banned type, 0 - info_hash, 1 - peer_id
	CreatedAt time.Time `json:"created_at" `
}

func (m *BlockList) TableName() string {
	return "block_list"
}
