package extensions

import (
	"cmp"
	"errors"
	"slices"

	"foundation"
)

var (
	ErrIndexOutOfBounds = errors.New("index out of bounds")
)

// ============================================================
// SLICE STRUCTURAL OPERATIONS
// ============================================================

/*
InsertAt returns a new slice with value inserted at index.

Semantics:
  - Valid when: 0 <= index <= len(s)
  - index == len(s) appends
  - original slice is never mutated

This is a convenience helper.
In hot paths prefer manual append/copy logic to avoid allocations.
*/
func InsertAt[T any](s []T, index int, value T) ([]T, error) {
	if !foundation.IsNumberValid(index, foundation.BoundsConfiguration{
		Min:          0,
		Max:          len(s),
		MinInclusive: true,
		MaxInclusive: true,
	}) {
		return nil, ErrIndexOutOfBounds
	}

	result := make([]T, len(s)+1)

	copy(result[:index], s[:index])
	result[index] = value
	copy(result[index+1:], s[index:])

	return result, nil
}

/*
InsertSliceAt returns a new slice with values inserted at index.

Equivalent to splicing.

Example:

	InsertSliceAt([1 2 5], 2, [3 4]) → [1 2 3 4 5]
*/
func InsertSliceAt[T any](s []T, index int, values []T) ([]T, error) {
	if len(values) == 0 {
		return slices.Clone(s), nil
	}

	if !foundation.IsNumberValid(index, foundation.BoundsConfiguration{
		Min:          0,
		Max:          len(s),
		MinInclusive: true,
		MaxInclusive: true,
	}) {
		return nil, ErrIndexOutOfBounds
	}

	result := make([]T, len(s)+len(values))

	copy(result[:index], s[:index])
	copy(result[index:], values)
	copy(result[index+len(values):], s[index:])

	return result, nil
}

/*
RemoveAt returns a new slice with the element at index removed.

Valid when:

	0 <= index < len(s)

The original slice is never mutated.
*/
func RemoveAt[T any](s []T, index int) ([]T, error) {
	if !foundation.IsSliceIDXValid(index, s) {
		return nil, ErrIndexOutOfBounds
	}

	result := make([]T, len(s)-1)

	copy(result[:index], s[:index])
	copy(result[index:], s[index+1:])

	return result, nil
}

/*
RemoveWhere returns a new slice with all elements removed for which pred returns true.

The relative order of remaining elements is preserved.

This performs a single pass over the slice and allocates exactly once.

The original slice is never mutated.

Example:

	RemoveWhere([]int{1, 2, 3, 4}, func(v int) bool {
		return v%2 == 0
	})
	// → [1 3]
*/
func RemoveWhere[T any](s []T, pred func(T) bool) ([]T, int) {
	keepCount := 0
	for _, v := range s {
		if !pred(v) {
			keepCount++
		}
	}

	deletedCount := len(s) - keepCount

	if keepCount == len(s) {
		return slices.Clone(s), 0
	}

	if keepCount == 0 {
		return []T{}, deletedCount
	}

	result := make([]T, 0, keepCount)
	for _, v := range s {
		if !pred(v) {
			result = append(result, v)
		}
	}

	return result, deletedCount
}

/*
RemoveWhereInPlace compacts s in place and returns the kept prefix.

This mutates the original slice and performs zero allocations.
*/
func RemoveWhereInPlace[T any](s []T, pred func(T) bool) ([]T, int) {
	newLen := 0
	for i := range s {
		if !pred(s[i]) {
			s[newLen] = s[i]
			newLen++
		}
	}

	removedCount := len(s) - newLen

	clear(s[newLen:])

	return s[:newLen], removedCount
}

/*
ReplaceAt returns a new slice with the element at index replaced by value.

Valid when:

	0 <= index < len(s)

The original slice is never mutated.
*/
func ReplaceAt[T any](s []T, index int, value T) ([]T, error) {
	if !foundation.IsSliceIDXValid(index, s) {
		return nil, ErrIndexOutOfBounds
	}

	result := slices.Clone(s)
	result[index] = value

	return result, nil
}

// ============================================================
// ACCESS & SEARCH UTILITIES
// ============================================================

/*
GetAt returns the element at index and whether the access was valid.

No panic is triggered for out-of-range indices.
*/
func GetAt[T any](s []T, index int) (T, bool) {
	if !foundation.IsSliceIDXValid(index, s) {
		var zero T
		return zero, false
	}

	return s[index], true
}

/*
IndexOfFunc returns the index of the first element satisfying pred.

Returns -1 if no element matches.
*/
func IndexOfFunc[T any](s []T, pred func(T) bool) int {
	for i, v := range s {
		if pred(v) {
			return i
		}
	}
	return -1
}

/*
Equal reports whether a and b have the same length and equal elements
at every index.

Two nil slices are equal.
A nil slice and an empty (non-nil) slice are not equal.
*/
func Equal[T comparable](a []T, b []T) bool {
	if len(a) != len(b) {
		return false
	}

	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}

	return true
}

/*
EqualFunc reports whether a and b have the same length and all element pairs
compare equal under eq.

Two nil slices are equal.
A nil slice and an empty (non-nil) slice are not equal.
*/
func EqualFunc[T any](a []T, b []T, eq func(a, b T) bool) bool {
	if len(a) != len(b) {
		return false
	}

	for i := range a {
		if !eq(a[i], b[i]) {
			return false
		}
	}

	return true
}

// ============================================================
// SORTING HELPERS
// ============================================================

/*
SortedCopyShallow returns a new slice containing the elements of s
sorted using cmpFunc.

The original slice is never mutated.

If s is nil, nil is returned.

This is a convenience helper intended for clarity and immutability.
For performance-critical code, prefer sorting in place.
*/
func SortedCopyShallow[T any](s []T, cmpFunc func(a, b T) int) []T {
	if s == nil {
		return nil
	}

	result := slices.Clone(s)
	slices.SortFunc(result, cmpFunc)

	return result
}

/*
OrderedCMPGenerator returns a comparison function for ordered types.

If descending is false:

	a < b → negative
	a == b → zero
	a > b → positive

If descending is true, the order is reversed.

Useful for passing into slices.SortFunc without rewriting comparison logic.
*/
func OrderedCMPGenerator[T cmp.Ordered](descending bool) func(a, b T) int {
	if descending {
		return func(a, b T) int {
			return cmp.Compare(b, a)
		}
	}

	return func(a, b T) int {
		return cmp.Compare(a, b)
	}
}
