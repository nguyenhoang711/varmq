package types

type Item[T any] struct {
	Value    T
	Priority int
	Index    int
}
