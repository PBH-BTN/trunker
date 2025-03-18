package local

import (
	"sync/atomic"
	"time"

	"github.com/PBH-BTN/trunker/biz/config"
	"github.com/PBH-BTN/trunker/biz/model"
	"github.com/PBH-BTN/trunker/biz/services/peer/common"
)

func (m *Manager) Clean() int64 {
	count := atomic.Int64{}
	m.infoHashMap.Range(func(key string, value *InfoHashRoot) bool {
		count.Add(m.cleanUp(value))
		return true
	})
	return count.Load()
}
func (m *Manager) cleanUp(root *InfoHashRoot) int64 {
	if root.lastClean.Add(time.Duration(config.AppConfig.Tracker.TTL) * time.Second).After(time.Now()) {
		return 0
	}
	toClean := make([]string, 0)
	expireTime := time.Now().Add(time.Duration(-1*config.AppConfig.Tracker.TTL) * time.Second)
	root.Range(func(key string, value *common.Peer) bool {
		if expireTime.After(value.LastSeen) {
			toClean = append(toClean, key)
			if value.Type == model.PeerTypeWebtorrent && value.Conn != nil {
				_ = value.Conn.Close()
				value.Conn = nil
			}
		}
		return true
	})
	for _, key := range toClean {
		root.LoadAndDelete(key)
	}
	root.lastClean = time.Now()
	if root.Len() == 0 {
		m.infoHashMap.Delete(root.infoHash)
	}
	return int64(len(toClean))
}
