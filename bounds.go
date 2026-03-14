package foundation

/*
BoundsConfiguration defines a numeric interval with explicit inclusivity.

It represents the interval:

	(Min, Max)
	[Min, Max)
	(Min, Max]
	[Min, Max]

depending on MinInclusive and MaxInclusive.

This abstraction exists for correctness, clarity, and reuse.
In hot paths, prefer writing direct comparisons for maximum performance.
*/
type BoundsConfiguration struct {
	Min, Max int

	/* If true, Min is allowed (>= Min), otherwise strictly greater (> Min) */
	MinInclusive bool

	/* If true, Max is allowed (<= Max), otherwise strictly less (< Max) */
	MaxInclusive bool
}

/*
IsNumberValid checks whether a number lies within the configured bounds.

Semantics:

	MinInclusive = true  → number >= Min
	MinInclusive = false → number >  Min

	MaxInclusive = true  → number <= Max
	MaxInclusive = false → number <  Max

This function prioritizes clarity and correctness.
For tight loops or critical performance paths, inline the comparisons manually.
*/
func IsNumberValid(number int, bounds BoundsConfiguration) bool {

	/* ---- Lower bound ---- */

	if bounds.MinInclusive {
		if number < bounds.Min {
			return false
		}
	} else {
		if number <= bounds.Min {
			return false
		}
	}

	/* ---- Upper bound ---- */

	if bounds.MaxInclusive {
		if number > bounds.Max {
			return false
		}
	} else {
		if number >= bounds.Max {
			return false
		}
	}

	return true
}

/* IsBetweenInclusive checks whether number ∈ [min, max]. */
func IsBetweenInclusive(number, min, max int) bool {
	return IsNumberValid(number, BoundsConfiguration{
		Min:          min,
		Max:          max,
		MinInclusive: true,
		MaxInclusive: true,
	})
}

/* IsBetweenExclusive checks whether number ∈ (min, max). */
func IsBetweenExclusive(number, min, max int) bool {
	return IsNumberValid(number, BoundsConfiguration{
		Min:          min,
		Max:          max,
		MinInclusive: false,
		MaxInclusive: false,
	})
}

/* IsBetweenLeftInclusive checks whether number ∈ [min, max). */
func IsBetweenLeftInclusive(number, min, max int) bool {
	return IsNumberValid(number, BoundsConfiguration{
		Min:          min,
		Max:          max,
		MinInclusive: true,
		MaxInclusive: false,
	})
}

/* IsBetweenRightInclusive checks whether number ∈ (min, max]. */
func IsBetweenRightInclusive(number, min, max int) bool {
	return IsNumberValid(number, BoundsConfiguration{
		Min:          min,
		Max:          max,
		MinInclusive: false,
		MaxInclusive: true,
	})
}

/*
IsSliceIDXValid checks whether number is a valid index into the slice.

Semantics:

	Valid when: 0 <= number < len(slice)

This is purely a convenience wrapper around IsNumberValid.
For performance-critical code, prefer:

	if i >= 0 && i < len(s) { ... }
*/
func IsSliceIDXValid[TSlice any](number int, s []TSlice) bool {
	return IsNumberValid(number, BoundsConfiguration{
		Min:          0,
		Max:          len(s),
		MinInclusive: true,
		MaxInclusive: false,
	})
}
