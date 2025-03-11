package utils

func If[T any](cond bool, t, f T) T {
	if cond {
		return t
	}
	return f
}

func IfLazyL[T any](cond bool, t func() T, f T) T {
	if cond {
		return t()
	}
	return f
}

func IfLazyR[T any](cond bool, t T, f func() T) T {
	if cond {
		return t
	}
	return f()
}
