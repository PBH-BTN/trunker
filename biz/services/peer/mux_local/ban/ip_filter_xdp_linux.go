//go:build linux

package ban

import (
	"fmt"
	"net"
	"strconv"
	"strings"

	"github.com/PBH-BTN/trunker/biz/config"
	"github.com/PBH-BTN/trunker/biz/services/peer/mux_local/ban/xdp"
	"github.com/cloudwego/hertz/pkg/common/hlog"
)

// xdpIPFilter implements IPFilter using XDP/eBPF on Linux
type xdpIPFilter struct {
	filter *xdp.XDPFilter
	iface  string
}

// parsePort extracts the port number from a "host:port" string
func parsePort(hostPort string) (uint16, error) {
	parts := strings.Split(hostPort, ":")
	if len(parts) < 2 {
		return 0, fmt.Errorf("invalid host:port format: %s", hostPort)
	}
	port, err := strconv.ParseUint(parts[len(parts)-1], 10, 16)
	if err != nil {
		return 0, fmt.Errorf("invalid port number: %s", parts[len(parts)-1])
	}
	return uint16(port), nil
}

// getFilterPorts extracts the ports to filter from the configuration
func getFilterPorts() []uint16 {
	var ports []uint16

	if config.AppConfig == nil {
		return ports
	}

	// Add HTTP tracker port
	if config.AppConfig.Tracker.HostPorts != "" {
		port, err := parsePort(config.AppConfig.Tracker.HostPorts)
		if err == nil {
			ports = append(ports, port)
			hlog.Infof("XDP will filter HTTP port: %d", port)
		} else {
			hlog.Warnf("Failed to parse HTTP port from %s: %v", config.AppConfig.Tracker.HostPorts, err)
		}
	}

	// Add UDP tracker port if enabled
	if config.AppConfig.Tracker.UDPServer.Enable && config.AppConfig.Tracker.UDPServer.HostPorts != "" {
		port, err := parsePort(config.AppConfig.Tracker.UDPServer.HostPorts)
		if err == nil {
			ports = append(ports, port)
			hlog.Infof("XDP will filter UDP port: %d", port)
		} else {
			hlog.Warnf("Failed to parse UDP port from %s: %v", config.AppConfig.Tracker.UDPServer.HostPorts, err)
		}
	}

	return ports
}

// NewIPFilter creates a new IP filter
// On Linux, it attempts to use XDP for kernel-level filtering based on config
func NewIPFilter(ifaceName string) (IPFilter, error) {
	// Check if XDP is enabled in config
	if config.AppConfig == nil || !config.AppConfig.Tracker.XDP.Enable {
		hlog.Info("XDP is disabled in config, using fallback IP filter (application layer)")
		return newFallbackIPFilter(), nil
	}

	if ifaceName == "" {
		hlog.Warn("XDP enabled but no interface specified, using fallback IP filter (application layer)")
		return newFallbackIPFilter(), nil
	}

	// Get the ports to filter from configuration
	filterPorts := getFilterPorts()
	if len(filterPorts) == 0 {
		hlog.Warn("No filter ports configured, XDP will not filter any traffic")
	}

	// Try to initialize XDP filter
	filter, err := xdp.NewXDPFilter(ifaceName, filterPorts)
	if err != nil {
		hlog.Warnf("Failed to initialize XDP filter on %s: %v, falling back to application layer", ifaceName, err)
		return newFallbackIPFilter(), nil
	}

	hlog.Infof("XDP IP filter successfully attached to interface: %s (kernel-level filtering enabled)", ifaceName)
	return &xdpIPFilter{
		filter: filter,
		iface:  ifaceName,
	}, nil
}

func (x *xdpIPFilter) AddIP(ip net.IP) error {
	if ip.To4() != nil {
		return x.filter.AddIPv4(ip)
	}
	return x.filter.AddIPv6(ip)
}

func (x *xdpIPFilter) RemoveIP(ip net.IP) error {
	if ip.To4() != nil {
		return x.filter.RemoveIPv4(ip)
	}
	return x.filter.RemoveIPv6(ip)
}

func (x *xdpIPFilter) TestIP(ip net.IP) bool {
	if ip.To4() != nil {
		return x.filter.TestIPv4(ip)
	}
	return x.filter.TestIPv6(ip)
}

func (x *xdpIPFilter) Clear() error {
	return x.filter.Clear()
}

func (x *xdpIPFilter) Close() error {
	if x.filter != nil {
		return x.filter.Close()
	}
	return nil
}

func (x *xdpIPFilter) Stats() *IPFilterStats {
	dropped, allowed, _ := x.filter.GetStats()
	ipv4Count, ipv6Count := x.filter.GetIPCount()

	return &IPFilterStats{
		TotalIPs:       ipv4Count + ipv6Count,
		PacketsDropped: dropped,
		PacketsAllowed: allowed,
		IsXDP:          true,
		InterfaceName:  x.iface,
	}
}

func (x *xdpIPFilter) String() string {
	stats := x.Stats()
	return fmt.Sprintf("XDP IP Filter on %s: %d IPs blocked, %d pkts dropped, %d pkts allowed",
		stats.InterfaceName, stats.TotalIPs, stats.PacketsDropped, stats.PacketsAllowed)
}
