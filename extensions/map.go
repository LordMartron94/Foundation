package extensions

import (
	"maps"
	"slices"
)

// ============================================================
// MAP ENTRY MODEL
// ============================================================

/*
KeyValuePair represents a single map entry.

Used for transforming maps into ordered, slice-based structures
for sorting, iteration, and functional-style processing.
*/
type KeyValuePair[K comparable, V any] struct {
	Key   K
	Value V
}

// ============================================================
// MAP → SORTED SLICE
// ============================================================

/*
MapSortFunc returns a slice of KeyValuePair entries sorted using cmpFn.

The original map is never mutated.

cmpFn must return:
  - negative when pairA < pairB
  - zero when pairA == pairB
  - positive when pairA > pairB

If m is nil, nil is returned.

This is a convenience helper.
For performance-critical paths, consider collecting and sorting manually.
*/
func MapSortFunc[K comparable, V any](
	m map[K]V,
	cmpFn func(pairA, pairB KeyValuePair[K, V]) int,
) []KeyValuePair[K, V] {
	if m == nil {
		return nil
	}

	entries := make([]KeyValuePair[K, V], 0, len(m))
	for k, v := range m {
		entries = append(entries, KeyValuePair[K, V]{Key: k, Value: v})
	}

	slices.SortFunc(entries, cmpFn)

	return entries
}

// ============================================================
// MAP COLLECTION HELPERS
// ============================================================

/*
MapKeys returns a slice containing all keys of m.

The order is unspecified.

If m is nil, nil is returned.
*/
func MapKeys[K comparable, V any](m map[K]V) []K {
	if m == nil {
		return nil
	}

	keys := make([]K, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}

	return keys
}

/*
MapValues returns a slice containing all values of m.

The order corresponds to MapKeys and is unspecified.

If m is nil, nil is returned.
*/
func MapValues[K comparable, V any](m map[K]V) []V {
	if m == nil {
		return nil
	}

	values := make([]V, 0, len(m))
	for _, v := range m {
		values = append(values, v)
	}

	return values
}

// ============================================================
// MAP TRANSFORMATION UTILITIES
// ============================================================

/*
MapCloneShallow returns a shallow copy of m.

Keys and values are copied by assignment.
Reference types remain shared.

If m is nil, nil is returned.
*/
func MapCloneShallow[K comparable, V any](m map[K]V) map[K]V {
	if m == nil {
		return nil
	}

	clone := make(map[K]V, len(m))
	maps.Copy(clone, m)

	return clone
}

/*
MapMergeShallow returns a new map containing all entries of base
overridden by entries of overlay.

If a key exists in both, overlay wins.

Neither input map is mutated.
*/
func MapMergeShallow[K comparable, V any](
	base map[K]V,
	overlay map[K]V,
) map[K]V {
	if base == nil && overlay == nil {
		return nil
	}

	result := make(map[K]V, len(base)+len(overlay))
	maps.Copy(result, base)
	maps.Copy(result, overlay)

	return result
}
