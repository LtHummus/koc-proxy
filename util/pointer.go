package util

func P[T any](x T) *T {
	return &x
}
