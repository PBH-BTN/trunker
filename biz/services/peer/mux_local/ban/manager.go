package ban

import (
	"github.com/PBH-BTN/trunker/utils/conv"
)

type banType int8

const (
	BanTypePeerId banType = iota
	BanTypeInfoHash
)

type Manager struct {
	infoHash *banItem
	peerId   *banItem
}

func NewBanManager() *Manager {
	infoHash, err := newBanItem("banInfoHash.dat")
	if err != nil {
		panic(err)
	}
	peerId, err := newBanItem("banPeerId.dat")
	if err != nil {
		panic(err)
	}
	return &Manager{
		infoHash: infoHash,
		peerId:   peerId,
	}
}

func (b *Manager) getItem(targetType banType) *banItem {
	switch targetType {
	case BanTypePeerId:
		return b.peerId
	case BanTypeInfoHash:
		return b.infoHash
	default:
		panic("invalid ban type")
	}
}

func (b *Manager) AddBan(targetType banType, target string) error {
	return b.getItem(targetType).add(target)
}

func (b *Manager) Test(targetType banType, target string) bool {
	return b.getItem(targetType).test(conv.UnsafeStringToBytes(target))
}

func (b *Manager) Clear(targetType banType) {
	b.getItem(targetType).clear()
}
