package peer

import (
	"arena"
	"time"

	"github.com/PBH-BTN/trunker/biz/config"
)

func CleanUp(root *InfoHashRoot) {
	a := arena.NewArena()
	toClean := arena.MakeSlice[string](a, 0, root.peerMap.Len()/2)
	root.peerMap.Range(func(key string, value *Peer) bool {
		if time.Now().Add(time.Duration(-1*config.AppConfig.Tracker.TTL) * time.Second).After(value.LastSeen) {
			toClean = append(toClean, key)
		}
		return true
	})
	for _, key := range toClean {
		root.peerMap.Delete(key)
	}
	a.Free()
	root.lastClean = time.Now()
}
