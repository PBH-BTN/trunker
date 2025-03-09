//go:build !go1.24

package mapx

import "github.com/zhangyunhao116/skipmap"

// before go1.24, the skipmap has better performance
func New[V any]() SyncStringMap[V] {
	return skipmap.NewString[V]()
}
