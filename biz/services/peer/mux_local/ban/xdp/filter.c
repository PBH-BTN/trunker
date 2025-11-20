// +build ignore
// This is the eBPF C program that runs in the kernel via XDP
// It performs ultra-fast IP filtering at the network driver level

#include <linux/bpf.h>
#include <linux/if_ether.h>
#include <linux/ip.h>
#include <linux/ipv6.h>
#include <linux/tcp.h>
#include <linux/udp.h>
#include <linux/in.h>
#include <linux/in6.h>
#include <bpf/bpf_helpers.h>
#include <bpf/bpf_endian.h>

// Maximum number of IPs in the blocklist
// BPF maps have size limits, 100000 should handle most use cases
#define MAX_BLOCKED_IPS 100000
#define MAX_FILTER_PORTS 16

// BPF map to store blocked IPv4 addresses
// Key: __u32 (IPv4 address in network byte order)
// Value: __u8 (dummy value, we only care about key existence)
struct {
    __uint(type, BPF_MAP_TYPE_HASH);
    __uint(max_entries, MAX_BLOCKED_IPS);
    __type(key, __u32);
    __type(value, __u8);
} blocked_ipv4 SEC(".maps");

// BPF map to store blocked IPv6 addresses
// Key: __u8[16] (IPv6 address, 16 bytes)
// Value: __u8 (dummy value)
struct {
    __uint(type, BPF_MAP_TYPE_HASH);
    __uint(max_entries, MAX_BLOCKED_IPS);
    __type(key, __u8[16]);
    __type(value, __u8);
} blocked_ipv6 SEC(".maps");

// BPF map to store filter ports (only filter traffic to these ports)
// Key: __u16 (port number in host byte order)
// Value: __u8 (dummy value)
struct {
    __uint(type, BPF_MAP_TYPE_HASH);
    __uint(max_entries, MAX_FILTER_PORTS);
    __type(key, __u16);
    __type(value, __u8);
} filter_ports SEC(".maps");

// Statistics counters
struct {
    __uint(type, BPF_MAP_TYPE_ARRAY);
    __uint(max_entries, 2);
    __type(key, __u32);
    __type(value, __u64);
} stats SEC(".maps");

#define STAT_DROPPED 0
#define STAT_ALLOWED 1

// XDP program entry point
SEC("xdp")
int xdp_ip_filter(struct xdp_md *ctx)
{
    void *data_end = (void *)(long)ctx->data_end;
    void *data = (void *)(long)ctx->data;

    // Parse Ethernet header
    struct ethhdr *eth = data;
    if ((void *)(eth + 1) > data_end)
        return XDP_PASS;

    // Check if it's an IP packet (IPv4 or IPv6)
    __u16 eth_proto = bpf_ntohs(eth->h_proto);
    __u16 dest_port = 0;
    __u8 protocol = 0;
    void *l4_header = NULL;

    if (eth_proto == ETH_P_IP) {
        // IPv4 packet
        struct iphdr *ip = (void *)(eth + 1);
        if ((void *)(ip + 1) > data_end)
            return XDP_PASS;

        protocol = ip->protocol;
        l4_header = (void *)ip + (ip->ihl * 4);
        __u32 saddr = ip->saddr;

        // Parse TCP/UDP header to get destination port
        if (protocol == IPPROTO_TCP) {
            struct tcphdr *tcp = l4_header;
            if ((void *)(tcp + 1) > data_end)
                return XDP_PASS;
            dest_port = bpf_ntohs(tcp->dest);
        } else if (protocol == IPPROTO_UDP) {
            struct udphdr *udp = l4_header;
            if ((void *)(udp + 1) > data_end)
                return XDP_PASS;
            dest_port = bpf_ntohs(udp->dest);
        } else {
            // Not TCP/UDP, pass through
            return XDP_PASS;
        }

        // Check if this port should be filtered
        __u8 *port_match = bpf_map_lookup_elem(&filter_ports, &dest_port);
        if (!port_match) {
            __u32 key = STAT_ALLOWED;
            __u64 *counter = bpf_map_lookup_elem(&stats, &key);
            if (counter)
                __sync_fetch_and_add(counter, 1);
            return XDP_PASS;
        }

        // Port matches, check IP blocklist
        __u8 *blocked = bpf_map_lookup_elem(&blocked_ipv4, &saddr);
        if (blocked) {
            // IP is blocked, drop the packet
            bpf_printk("XDP: BLOCKING IPv4 %pI4 on port %d", &saddr, dest_port);
            __u32 key = STAT_DROPPED;
            __u64 *counter = bpf_map_lookup_elem(&stats, &key);
            if (counter)
                __sync_fetch_and_add(counter, 1);
            return XDP_DROP;
        }
    }
    else if (eth_proto == ETH_P_IPV6) {
        // IPv6 packet
        struct ipv6hdr *ip6 = (void *)(eth + 1);
        if ((void *)(ip6 + 1) > data_end)
            return XDP_PASS;

        protocol = ip6->nexthdr;
        l4_header = (void *)(ip6 + 1);

        // Parse TCP/UDP header to get destination port
        if (protocol == IPPROTO_TCP) {
            struct tcphdr *tcp = l4_header;
            if ((void *)(tcp + 1) > data_end)
                return XDP_PASS;
            dest_port = bpf_ntohs(tcp->dest);
        } else if (protocol == IPPROTO_UDP) {
            struct udphdr *udp = l4_header;
            if ((void *)(udp + 1) > data_end)
                return XDP_PASS;
            dest_port = bpf_ntohs(udp->dest);
        } else {
            // Not TCP/UDP, pass through
            return XDP_PASS;
        }
        // Check if this port should be filtered
        __u8 *port_match = bpf_map_lookup_elem(&filter_ports, &dest_port);
        if (!port_match) {
            __u32 key = STAT_ALLOWED;
            __u64 *counter = bpf_map_lookup_elem(&stats, &key);
            if (counter)
                __sync_fetch_and_add(counter, 1);
            return XDP_PASS;
        }

        // Port matches, check IP blocklist
        __u8 *blocked = bpf_map_lookup_elem(&blocked_ipv6, &ip6->saddr);
        if (blocked) {
            // IP is blocked, drop the packet
            bpf_printk("XDP: BLOCKING IPv6 %pI6c on port %d", &ip6->saddr, dest_port);
            __u32 key = STAT_DROPPED;
            __u64 *counter = bpf_map_lookup_elem(&stats, &key);
            if (counter)
                __sync_fetch_and_add(counter, 1);
            return XDP_DROP;
        }
    }

    // Not blocked, allow the packet
    __u32 key = STAT_ALLOWED;
    __u64 *counter = bpf_map_lookup_elem(&stats, &key);
    if (counter)
        __sync_fetch_and_add(counter, 1);

    return XDP_PASS;
}

char __license[] SEC("license") = "GPL";