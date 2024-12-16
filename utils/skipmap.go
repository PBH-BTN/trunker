package utils

import "golang.org/x/exp/constraints"

type rangeAble[K constraints.Ordered, V any] interface {
	Range(f func(key K, value V) bool)
	Len() int
}

func SkipMapToSlice[K constraints.Ordered, V any](m rangeAble[K, V]) []V {
	res := make([]V, 0, m.Len())
	m.Range(func(_ K, value V) bool {
		res = append(res, value)
		return true
	})
	return res
}
