package text

// LevenshteinDistance computes the levenshtein distance between two strings.
// See: https://en.wikipedia.org/wiki/Levenshtein_distance
func LevenshteinDistance(s1, s2 string) int {
	m, n := len(s1), len(s2)
	if m < n {
		s1, s2 = s2, s1
		m, n = n, m
	}

	row := make([]int, n+1)
	for i := 0; i <= n; i++ {
		row[i] = i
	}

	for i := 1; i <= m; i++ {
		prevResult := row[0]
		row[0] = i
		for j := 1; j <= n; j++ {
			temp := row[j]
			var cost int
			if s1[i-1] != s2[j-1] {
				cost = 1
			}

			// row[j] is deletion, row[j-1] is insertion, prevResult is substitution
			row[j] = min(
				row[j]+1,
				row[j-1]+1,
				prevResult+cost,
			)
			prevResult = temp
		}
	}

	return row[n]
}
