package common

import "errors"

var (
	ErrNotSupportedInRPCMode = errors.New("this operation is not supported in RPC mode, configure on remote server")
)

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
