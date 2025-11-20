package mux_local

import (
	"crypto/rand"
	"fmt"
	"net"
	"os"
	"testing"
	"time"

	"github.com/PBH-BTN/trunker/biz/config"
	"github.com/PBH-BTN/trunker/biz/model"
	"github.com/PBH-BTN/trunker/biz/services/peer/common"
	"github.com/PBH-BTN/trunker/utils/conv"
	json "github.com/bytedance/sonic"
	"github.com/cilium/fake"
	"github.com/stretchr/testify/assert"
	"github.com/zhangyunhao116/fastrand"
)

func TestMuxLocalManager_StoreToPersist(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "persist_test")
	if err != nil {
		t.Fatal(err)
	}
	configStr := fmt.Sprintf(`{"tracker": {"memory": {"persistFile": "%s/persist.dat","enablePersist":true},"ttl":600}}`, tempDir)
	err = json.Unmarshal(conv.UnsafeStringToBytes(configStr), &config.AppConfig)
	if err != nil {
		t.Fatal(err)
	}
	m := NewMuxLocalManager(10)
	// generate 100000 peers
	now := time.Now()
	for i := 0; i < 100; i++ {
		infoHash := rand.Text()[:20]
		for j := 0; j < 1000; j++ {
			peer := &common.Peer{
				IP:         net.ParseIP(fake.IP(fake.WithIPv4())),
				IPv4:       net.ParseIP(fake.IP(fake.WithIPv4())),
				IPv6:       net.ParseIP(fake.IP(fake.WithIPv6())),
				ClientIP:   net.ParseIP(fake.IP(fake.WithIPv4())),
				LastSeen:   now,
				ID:         rand.Text()[:20],
				UserAgent:  rand.Text(),
				Port:       fastrand.Intn(65534),
				Left:       uint64(fastrand.Uint32()),
				Uploaded:   uint64(fastrand.Uint32()),
				Downloaded: uint64(fastrand.Uint32()),
				Type:       model.PeerTypeBittorrent,
				Event:      common.PeerEvent_Started,
				Source:     model.SourceHTTP,
			}
			// fuzz test
			if fastrand.Uint64()%2 == 0 {
				peer.IPv6 = nil
			}
			if fastrand.Uint64()%3 == 1 {
				peer.IPv4 = nil
			}
			m.pickWorker(conv.UnsafeStringToBytes(infoHash)).DirectStore(infoHash, peer)
		}
	}
	// test store to persist
	m.StoreToPersist()

	// load from persist
	m2 := NewMuxLocalManager(10)
	m2.LoadFromPersist()
	statistic1, err := m.GetStatistic(t.Context())
	if err != nil {
		return
	}
	statistic2, err := m2.GetStatistic(t.Context())
	if err != nil {
		return
	}
	for i, s2 := range statistic2.Shards {
		assert.Equal(t, s2.TotalPeers, statistic1.Shards[i].TotalPeers)
		assert.Equal(t, s2.TotalTorrents, statistic1.Shards[i].TotalTorrents)
	}

}
