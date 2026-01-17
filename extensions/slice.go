package extensions

import (
	"cmp"
	"slices"
)

/* SortedCopyShallow returns a new slice containing the elements of s sorted by cmpFunc. */
func SortedCopyShallow[T any](s []T, cmpFunc func(a, b T) int) []T {
	if s == nil {
		return nil
	}

	result := slices.Clone(s)
	slices.SortFunc(result, cmpFunc)

	return result
}

/* OrderedCMPGenerator creates a comparison func for type T */
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
