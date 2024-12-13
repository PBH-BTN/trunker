package utils

type ordered interface {
	~int | ~int8 | ~int16 | ~int32 | ~int64 | // sign
		~uint | ~uint8 | ~uint16 | ~uint32 | ~uint64 | ~uintptr | // unsign
		~float32 | ~float64 | // float
		~string
}

type rangeAble[keyT ordered, valueT any] interface {
	Range(f func(key keyT, value valueT) bool)
	Len() int
}

func SkipMapToSlice[K ordered, V any](m rangeAble[K, V]) []V {
	res := make([]V, 0, m.Len())
	m.Range(func(_ K, value V) bool {
		res = append(res, value)
		return true
	})
	return res
}
