package mapx

import (
	"github.com/bytedance/gg/collection/skipmap"
)

// NewSkipMap before go1.24, the skipmap has better performance
func NewSkipMap[V any]() SyncStringMap[V] {
	return skipmap.New[string, V]()
}
