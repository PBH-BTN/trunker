package utils

import "golang.org/x/exp/constraints"

func Positive[T constraints.Integer](x T) T {
	if x < 0 {
		return 0
	}
	return x
}
