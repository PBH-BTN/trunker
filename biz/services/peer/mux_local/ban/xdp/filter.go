// Package xdp provides XDP-based IP filtering for Linux
package xdp

//go:generate go run github.com/cilium/ebpf/cmd/bpf2go -go-package xdp  -target amd64,arm64 bpf filter.c -- -I/usr/include/bpf -I/usr/include

import (
	"encoding/binary"
	"errors"
	"fmt"
	"net"
	"sync"

	"github.com/cilium/ebpf/link"
	"github.com/cilium/ebpf/rlimit"
)

var (
	ErrNotLinux     = errors.New("XDP is only supported on Linux")
	ErrNoInterface  = errors.New("network interface not found")
	ErrLoadFailed   = errors.New("failed to load XDP program")
	ErrAttachFailed = errors.New("failed to attach XDP program")
)

// XDPFilter manages XDP-based IP filtering
type XDPFilter struct {
	mu sync.RWMutex

	objs      *bpfObjects
	link      link.Link
	iface     string
	ipv4Count uint64
	ipv6Count uint64
}

func NewXDPFilter(ifaceName string, filterPorts []uint16) (*XDPFilter, error) {
	if err := rlimit.RemoveMemlock(); err != nil {
		return nil, fmt.Errorf("failed to remove memlock limit: %w", err)
	}

	// Load the compiled eBPF objects
	objs := &bpfObjects{}
	if err := loadBpfObjects(objs, nil); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrLoadFailed, err)
	}

	// Populate the filter_ports map with the provided ports
	var dummy uint8 = 1
	for _, port := range filterPorts {
		if err := objs.FilterPorts.Put(&port, &dummy); err != nil {
			objs.Close()
			return nil, fmt.Errorf("failed to add port %d to filter map: %w", port, err)
		}
		fmt.Printf("[XDP] Added port %d to filter list\n", port)
	}

	// Find the network interface
	iface, err := net.InterfaceByName(ifaceName)
	if err != nil {
		objs.Close()
		return nil, fmt.Errorf("%w: %s: %v", ErrNoInterface, ifaceName, err)
	}

	// Attach the XDP program to the interface
	// XDP_MODE_NATIVE: attaches to the NIC driver (highest performance)
	// Falls back to XDP_MODE_SKB if driver doesn't support native XDP
	l, err := link.AttachXDP(link.XDPOptions{
		Program:   objs.XdpIpFilter,
		Interface: iface.Index,
	})
	if err != nil {
		objs.Close()
		return nil, fmt.Errorf("%w: %v", ErrAttachFailed, err)
	}

	fmt.Printf("[XDP] Successfully attached to interface %s with %d filter ports\n", ifaceName, len(filterPorts))

	return &XDPFilter{
		objs:  objs,
		link:  l,
		iface: ifaceName,
	}, nil
}

// AddIPv4 adds an IPv4 address to the blocklist
func (x *XDPFilter) AddIPv4(ip net.IP) error {
	x.mu.Lock()
	defer x.mu.Unlock()

	// Convert to 4-byte representation
	ip4 := ip.To4()
	if ip4 == nil {
		return errors.New("not a valid IPv4 address")
	}

	ipBytes := []byte(ip4)
	ipUint32 := binary.LittleEndian.Uint32(ipBytes)

	// Add to BPF map
	var dummy uint8 = 1
	if err := x.objs.BlockedIpv4.Put(&ipUint32, &dummy); err != nil {
		return fmt.Errorf("failed to add IPv4 to BPF map: %w", err)
	}

	x.ipv4Count++

	// Log the addition for debugging - show both representations
	fmt.Printf("[XDP] Added IPv4 to blocklist: %s (0x%08x) (total IPv4: %d)\n", ip.String(), ipUint32, x.ipv4Count)

	return nil
}

// AddIPv6 adds an IPv6 address to the blocklist
func (x *XDPFilter) AddIPv6(ip net.IP) error {
	x.mu.Lock()
	defer x.mu.Unlock()

	// Convert to 16-byte representation
	ip6 := ip.To16()
	if ip6 == nil {
		return errors.New("not a valid IPv6 address")
	}

	// Add to BPF map
	var dummy uint8 = 1
	if err := x.objs.BlockedIpv6.Put(ip6, &dummy); err != nil {
		return fmt.Errorf("failed to add IPv6 to BPF map: %w", err)
	}

	x.ipv6Count++

	// Log the addition for debugging
	fmt.Printf("[XDP] Added IPv6 to blocklist: %s (total IPv6: %d)\n", ip.String(), x.ipv6Count)

	return nil
}

// RemoveIPv4 removes an IPv4 address from the blocklist
func (x *XDPFilter) RemoveIPv4(ip net.IP) error {
	x.mu.Lock()
	defer x.mu.Unlock()

	ip4 := ip.To4()
	if ip4 == nil {
		return errors.New("not a valid IPv4 address")
	}

	ipBytes := []byte(ip4)
	ipUint32 := binary.LittleEndian.Uint32(ipBytes)
	if err := x.objs.BlockedIpv4.Delete(&ipUint32); err != nil {
		return fmt.Errorf("failed to remove IPv4 from BPF map: %w", err)
	}

	if x.ipv4Count > 0 {
		x.ipv4Count--
	}
	return nil
}

// RemoveIPv6 removes an IPv6 address from the blocklist
func (x *XDPFilter) RemoveIPv6(ip net.IP) error {
	x.mu.Lock()
	defer x.mu.Unlock()

	ip6 := ip.To16()
	if ip6 == nil {
		return errors.New("not a valid IPv6 address")
	}

	if err := x.objs.BlockedIpv6.Delete(ip6); err != nil {
		return fmt.Errorf("failed to remove IPv6 from BPF map: %w", err)
	}

	if x.ipv6Count > 0 {
		x.ipv6Count--
	}
	return nil
}

// TestIPv4 checks if an IPv4 address is in the blocklist
// Note: This reads from the BPF map, but actual filtering happens in kernel
func (x *XDPFilter) TestIPv4(ip net.IP) bool {
	x.mu.RLock()
	defer x.mu.RUnlock()

	ip4 := ip.To4()
	if ip4 == nil {
		return false
	}

	ipBytes := []byte(ip4)
	ipUint32 := binary.LittleEndian.Uint32(ipBytes)
	var dummy uint8
	err := x.objs.BlockedIpv4.Lookup(&ipUint32, &dummy)
	return err == nil
}

// TestIPv6 checks if an IPv6 address is in the blocklist
func (x *XDPFilter) TestIPv6(ip net.IP) bool {
	x.mu.RLock()
	defer x.mu.RUnlock()

	ip6 := ip.To16()
	if ip6 == nil {
		return false
	}

	var dummy uint8
	err := x.objs.BlockedIpv6.Lookup(ip6, &dummy)
	return err == nil
}

// Clear removes all IPs from the blocklist
func (x *XDPFilter) Clear() error {
	x.mu.Lock()
	defer x.mu.Unlock()

	// Clear IPv4 map
	var ipv4Key uint32
	iter := x.objs.BlockedIpv4.Iterate()
	for iter.Next(&ipv4Key, nil) {
		if err := x.objs.BlockedIpv4.Delete(&ipv4Key); err != nil {
			return fmt.Errorf("failed to clear IPv4 map: %w", err)
		}
	}

	// Clear IPv6 map
	var ipv6Key [16]byte
	iter6 := x.objs.BlockedIpv6.Iterate()
	for iter6.Next(&ipv6Key, nil) {
		if err := x.objs.BlockedIpv6.Delete(&ipv6Key); err != nil {
			return fmt.Errorf("failed to clear IPv6 map: %w", err)
		}
	}

	x.ipv4Count = 0
	x.ipv6Count = 0
	return nil
}

// GetStats returns statistics from the XDP program
func (x *XDPFilter) GetStats() (dropped, allowed uint64, err error) {
	x.mu.RLock()
	defer x.mu.RUnlock()

	var key uint32
	var value uint64

	// Get dropped packets count
	key = 0 // STAT_DROPPED
	if err := x.objs.Stats.Lookup(&key, &value); err == nil {
		dropped = value
	}

	// Get allowed packets count
	key = 1 // STAT_ALLOWED
	if err := x.objs.Stats.Lookup(&key, &value); err == nil {
		allowed = value
	}

	return dropped, allowed, nil
}

// GetIPCount returns the number of blocked IPs
func (x *XDPFilter) GetIPCount() (ipv4, ipv6 uint64) {
	x.mu.RLock()
	defer x.mu.RUnlock()
	return x.ipv4Count, x.ipv6Count
}

// GetInterface returns the interface name
func (x *XDPFilter) GetInterface() string {
	return x.iface
}

// Close detaches the XDP program and releases resources
func (x *XDPFilter) Close() error {
	x.mu.Lock()
	defer x.mu.Unlock()

	var errs []error
	if x.link != nil {
		if err := x.link.Close(); err != nil {
			errs = append(errs, fmt.Errorf("failed to close link: %w", err))
		}
	}

	if x.objs != nil {
		if err := x.objs.Close(); err != nil {
			errs = append(errs, fmt.Errorf("failed to close objects: %w", err))
		}
	}

	if len(errs) > 0 {
		return fmt.Errorf("errors during close: %v", errs)
	}
	return nil
}
