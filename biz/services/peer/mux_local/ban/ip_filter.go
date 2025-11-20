package ban

import (
	"net"
	"sync"
)

// IPFilter defines the interface for IP-based filtering
// This abstraction allows platform-specific implementations:
// - Linux: XDP/eBPF kernel-level filtering (high performance)
// - macOS/Others: Application-level filtering (fallback)
type IPFilter interface {
	// AddIP adds an IP address to the blocklist
	// Returns error if the operation fails
	AddIP(ip net.IP) error

	// RemoveIP removes an IP address from the blocklist
	RemoveIP(ip net.IP) error

	// TestIP checks if an IP address is blocked
	// Returns true if the IP should be blocked
	TestIP(ip net.IP) bool

	// Clear removes all IPs from the blocklist
	Clear() error

	// Close releases resources (detaches XDP programs, etc.)
	Close() error

	// Stats returns statistics about the filter
	Stats() *IPFilterStats
}

// IPFilterStats contains statistics about IP filtering
type IPFilterStats struct {
	// TotalIPs is the total number of blocked IPs
	TotalIPs uint64

	// PacketsDropped is the number of packets dropped (XDP only)
	PacketsDropped uint64

	// PacketsAllowed is the number of packets allowed (XDP only)
	PacketsAllowed uint64

	// IsXDP indicates if XDP filtering is active
	IsXDP bool

	// InterfaceName is the network interface name (XDP only)
	InterfaceName string
}

// fallbackIPFilter implements IPFilter using in-memory map
// Used on platforms that don't support XDP (macOS, Windows, etc.)
type fallbackIPFilter struct {
	mu      sync.RWMutex
	blocked map[string]struct{} // IP string -> empty struct
}

func newFallbackIPFilter() *fallbackIPFilter {
	return &fallbackIPFilter{
		blocked: make(map[string]struct{}),
	}
}

func (f *fallbackIPFilter) AddIP(ip net.IP) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.blocked[ip.String()] = struct{}{}
	return nil
}

func (f *fallbackIPFilter) RemoveIP(ip net.IP) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	delete(f.blocked, ip.String())
	return nil
}

func (f *fallbackIPFilter) TestIP(ip net.IP) bool {
	f.mu.RLock()
	defer f.mu.RUnlock()
	_, blocked := f.blocked[ip.String()]
	return blocked
}

func (f *fallbackIPFilter) Clear() error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.blocked = make(map[string]struct{})
	return nil
}

func (f *fallbackIPFilter) Close() error {
	// No resources to release for fallback implementation
	return nil
}

func (f *fallbackIPFilter) Stats() *IPFilterStats {
	f.mu.RLock()
	defer f.mu.RUnlock()
	return &IPFilterStats{
		TotalIPs: uint64(len(f.blocked)),
		IsXDP:    false,
	}
}