package utils

import "fmt"

var (
	ErrorIndexOutOfBounds error = fmt.Errorf("string buffer index out of bounds")
	ErrorIncorrectRange   error = fmt.Errorf("range start is lesser than end")
)

func SliceAddAtIndex[T any](dst []T, index int, item T) []T {
	if index < 0 || index > len(dst) {
		panic("index out of bounds")
	}

	right := append([]T{item}, dst[index:]...)
	return append(dst[:index], right...)
}
