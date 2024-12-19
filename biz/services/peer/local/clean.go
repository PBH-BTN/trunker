package local

import (
	"time"

	"github.com/PBH-BTN/trunker/biz/config"
	"github.com/PBH-BTN/trunker/biz/services/peer/common"
)

func (m *Manager) Clean() {
	m.infoHashMap.Range(func(key string, value *InfoHashRoot) bool {
		m.cleanUp(value)
		return true
	})
}
func (m *Manager) cleanUp(root *InfoHashRoot) {
	if root.lastClean.Add(time.Duration(config.AppConfig.Tracker.TTL) * time.Second).After(time.Now()) {
		return
	}
	toClean := make([]string, 0)
	root.peerMap.Range(func(key string, value *common.Peer) bool {
		if time.Now().Add(time.Duration(-1*config.AppConfig.Tracker.TTL) * time.Second).After(value.LastSeen) {
			toClean = append(toClean, key)
		}
		return true
	})
	for _, key := range toClean {
		_, ok := root.peerMap.LoadAndDelete(key)
		if ok {
			m.peerCount.Add(-1)
		}
	}
	root.lastClean = time.Now()
	if root.peerMap.Len() == 0 {
		m.infoHashMap.Delete(root.infoHash)
	}
}
