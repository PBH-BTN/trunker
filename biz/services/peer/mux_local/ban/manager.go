package ban

import (
	"errors"
	"net"

	"github.com/PBH-BTN/trunker/utils/conv"
)

var (
	ErrInvalidIP = errors.New("invalid IP address")
)

type banType int8

const (
	BanTypePeerId banType = iota
	BanTypeInfoHash
	BanTypeIP
)

type Manager struct {
	infoHash *banItem
	peerId   *banItem
	ipFilter IPFilter
}

// NewBanManager creates a new ban manager with InfoHash and PeerID filtering
func NewBanManager() *Manager {
	return NewBanManagerWithIP("")
}

// NewBanManagerWithIP creates a new ban manager with IP filtering support
// ifaceName: network interface name for XDP (e.g., "eth0", "ens33")
// If empty, IP filtering will use application-level fallback (or XDP will be disabled)
func NewBanManagerWithIP(ifaceName string) *Manager {
	infoHash, err := newBanItem("banInfoHash.dat")
	if err != nil {
		panic(err)
	}
	peerId, err := newBanItem("banPeerId.dat")
	if err != nil {
		panic(err)
	}

	// Initialize IP filter (XDP on Linux, fallback on other platforms)
	ipFilter, err := NewIPFilter(ifaceName)
	if err != nil {
		panic(err)
	}

	return &Manager{
		infoHash: infoHash,
		peerId:   peerId,
		ipFilter: ipFilter,
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

// AddIPBan adds an IP address to the blocklist
func (b *Manager) AddIPBan(ip string) error {
	parsedIP := net.ParseIP(ip)
	if parsedIP == nil {
		return ErrInvalidIP
	}
	return b.ipFilter.AddIP(parsedIP)
}

// RemoveIPBan removes an IP address from the blocklist
func (b *Manager) RemoveIPBan(ip string) error {
	parsedIP := net.ParseIP(ip)
	if parsedIP == nil {
		return ErrInvalidIP
	}
	return b.ipFilter.RemoveIP(parsedIP)
}

// TestIP checks if an IP address is blocked
// Note: With XDP on Linux, blocked packets never reach application layer
// This method is mainly for testing and compatibility
func (b *Manager) TestIP(ip string) bool {
	parsedIP := net.ParseIP(ip)
	if parsedIP == nil {
		return false
	}
	return b.ipFilter.TestIP(parsedIP)
}

// TestIPAddr checks if a net.IP is blocked
func (b *Manager) TestIPAddr(ip net.IP) bool {
	return b.ipFilter.TestIP(ip)
}

// ClearIPBans removes all IP bans
func (b *Manager) ClearIPBans() error {
	return b.ipFilter.Clear()
}

// GetIPFilterStats returns IP filter statistics
func (b *Manager) GetIPFilterStats() *IPFilterStats {
	return b.ipFilter.Stats()
}

// Close releases resources (important for XDP programs)
func (b *Manager) Close() error {
	if b.ipFilter != nil {
		return b.ipFilter.Close()
	}
	return nil
}
