package utils

import (
	"cmp"
)

// AsPtr returns a pointer to the given value v.
// Useful for converting a value to a pointer, especially in generic code.
//
// Example:
//
//	p := AsPtr(42) // p is of type *int, pointing to 42
func AsPtr[V any](v V) *V {
	return &v
}

// AsValue returns the value pointed to by v.
// If v is nil, it returns the zero value of type V.
// Useful for safely dereferencing pointers in generic code.
//
// Example:
//
//	var p *int
//	val := AsValue(p) // val is 0 (zero value for int)
func AsValue[V any](v *V) V {
	if v == nil {
		return *new(V)
	}
	return *v
}

// AsValueOrDefault returns the value pointed to by v.
// If v is nil, it returns the provided default value.
// Useful for safely dereferencing pointers in generic code.
//
// Example:
//
//	var p *int
//	val := AsValueOrDefault(p, 42) // val is 42 (default value)
func AsValueOrDefault[V any](v *V, defaultValue V) V {
	if v == nil {
		return defaultValue
	}
	return *v
}

func Equals[T comparable](a, b *T) bool {
	if a == nil && b == nil {
		return true
	} else if a != nil && b != nil {
		return *a == *b
	}
	return false
}

func ContainsKey[T comparable, V any](m map[T]V, key T) bool {
	_, ok := m[key]
	return ok
}

func Max[T cmp.Ordered](args ...T) T {
	if len(args) == 0 {
		return *new(T) // zero value of T
	}

	if isNan(args[0]) {
		return args[0]
	}

	max := args[0]
	for _, arg := range args[1:] {

		if isNan(arg) {
			return arg
		}

		if arg > max {
			max = arg
		}
	}
	return max
}

func isNan[T cmp.Ordered](arg T) bool {
	return arg != arg
}
