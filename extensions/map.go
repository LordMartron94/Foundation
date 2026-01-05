package extensions

type KeyValuePair[K comparable, V any] struct {
	Key   K
	Value V
}

/*
	ExtensionsMapSortFunc returns a sorted instance of a map based on the provided comparison function.

The comparison func must return:
<br>- -1 when KeyValuePair1 < KeyValuePair2
<br>- 0 when KeyValuePair1 == KeyValuePair2
<br>- +1 when KeyValuePair1 > KeyValuePair2
*/
func ExtensionsMapSortFunc[K comparable, V any](m map[K]V, cmpFn func(pairA, pairB KeyValuePair[K, V]) int) []KeyValuePair[K, V] {
	if m == nil {
		return nil
	}

	entries := make([]KeyValuePair[K, V], 0, len(m))
	for k, v := range m {
		entries = append(entries, KeyValuePair[K, V]{Key: k, Value: v})
	}

	sortEntries(entries, cmpFn)

	return entries
}

func sortEntries[K comparable, V any](entries []KeyValuePair[K, V], cmp func(pairA, pairB KeyValuePair[K, V]) int) {
	for i := 0; i < len(entries); i++ {
		for j := i + 1; j < len(entries); j++ {
			if cmp(entries[i], entries[j]) > 0 {
				entries[i], entries[j] = entries[j], entries[i]
			}
		}
	}
}
