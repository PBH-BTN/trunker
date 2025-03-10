package mapx

import "github.com/zhangyunhao116/skipmap"

// NewSkipMap before go1.24, the skipmap has better performance
func NewSkipMap[V any]() SyncStringMap[V] {
	return skipmap.NewString[V]()
}
