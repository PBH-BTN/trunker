package mapx

type SyncStringMap[V any] interface {
	Delete(key string) bool
	Load(key string) (value V, ok bool)
	LoadAndDelete(key string) (value V, loaded bool)
	LoadOrStore(key string, value V) (actual V, loaded bool)
	LoadOrStoreLazy(key string, lazy func() V) (actual V, loaded bool)
	Range(f func(key string, value V) bool)
	Store(key string, value V)
	Len() int
}
