package types

type Job[T, R any] struct {
	Data     T
	Response chan R
}
