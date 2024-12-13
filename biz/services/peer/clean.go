package peer

import (
	"arena"
	"time"

	"github.com/PBH-BTN/trunker/biz/config"
	"github.com/zhangyunhao116/skipmap"
)

func CleanUp(m *skipmap.OrderedMap[string, *Peer]) {
	a := arena.NewArena()
	toClean := arena.MakeSlice[string](a, 0, m.Len()/2)
	m.Range(func(key string, value *Peer) bool {
		if time.Now().Add(time.Duration(-1*config.AppConfig.Tracker.TTL) * time.Second).After(value.LastSeen) {
			toClean = append(toClean, key)
		}
		return true
	})
	for _, key := range toClean {
		m.Delete(key)
	}
	a.Free()
}
