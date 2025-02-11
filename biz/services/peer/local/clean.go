package local

import (
	"sync"
	"sync/atomic"
	"time"

	"github.com/PBH-BTN/trunker/biz/config"
	"github.com/PBH-BTN/trunker/biz/services/peer/common"
	"github.com/bytedance/gopkg/util/gopool"
)

func (m *Manager) Clean() int64 {
	count := atomic.Int64{}
	wg := sync.WaitGroup{}
	m.infoHashMap.Range(func(key string, value *InfoHashRoot) bool {
		wg.Add(1)
		gopool.Go(func() {
			defer wg.Done()
			cleaned := m.cleanUp(value)
			count.Add(cleaned)
		})
		return true
	})
	wg.Wait()
	return count.Load()
}
func (m *Manager) cleanUp(root *InfoHashRoot) int64 {
	if root.lastClean.Add(time.Duration(config.AppConfig.Tracker.TTL) * time.Second).After(time.Now()) {
		return 0
	}
	toClean := make([]string, 0)
	root.peerMap.Range(func(key string, value *common.Peer) bool {
		if time.Now().Add(time.Duration(-1*config.AppConfig.Tracker.TTL) * time.Second).After(value.LastSeen) {
			toClean = append(toClean, key)
		}
		return true
	})
	for _, key := range toClean {
		root.peerMap.Delete(key)
	}
	root.lastClean = time.Now()
	if root.peerMap.Len() == 0 {
		m.infoHashMap.Delete(root.infoHash)
	}
	return int64(len(toClean))
}
