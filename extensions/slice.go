package extensions

import "slices"

/* SortedCopyShallow returns a new slice containing the elements of s sorted by cmpFunc. */
func SortedCopyShallow[T any](s []T, cmpFunc func(a, b T) int) []T {
	if s == nil {
		return nil
	}

	result := slices.Clone(s)
	slices.SortFunc(result, cmpFunc)

	return result
}
