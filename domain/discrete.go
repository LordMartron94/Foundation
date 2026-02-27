package domain

import (
	"math"
	"unicode"
)

/*
DiscreteDomain defines a discrete, totally ordered value space with successor and previous steps.

It is used by pattern and lexer compilation (e.g. autarch/pattern, lexarch) to compare
observations, sort and merge ranges, and build symbol alphabets. Callers pass a single
DiscreteDomain instead of separate comparison and successor functions.

Fields:
  - Min: smallest value in the domain.
  - Max: largest value in the domain.
  - NextFn: returns the next value after current; second return is false at Max (no successor).
  - PreviousFn: returns the previous value before current; second return is false at Min (no predecessor).
  - OrderingCmp: three-way comparison; must return <0, 0, or >0 like cmp.Compare(a, b).
*/
type DiscreteDomain[TDomainElement any] struct {
	Min         TDomainElement
	Max         TDomainElement
	NextFn      func(current TDomainElement) (next TDomainElement, exist bool)
	PreviousFn  func(current TDomainElement) (previous TDomainElement, exist bool)
	OrderingCmp func(a, b TDomainElement) int
}

/*
IsMax returns true if val equals the domain maximum.
*/
func (d *DiscreteDomain[TDomainElement]) IsMax(val TDomainElement) bool {
	return d.OrderingCmp(val, d.Max) == 0
}

/*
IsMin returns true if val equals the domain minimum.
*/
func (d *DiscreteDomain[TDomainElement]) IsMin(val TDomainElement) bool {
	return d.OrderingCmp(val, d.Min) == 0
}

/*
LessThan returns true if a is strictly before b in the domain ordering.
*/
func (d *DiscreteDomain[TDomainElement]) LessThan(a, b TDomainElement) bool {
	return d.OrderingCmp(a, b) < 0
}

/*
GreaterThan returns true if a is strictly after b in the domain ordering.
*/
func (d *DiscreteDomain[TDomainElement]) GreaterThan(a, b TDomainElement) bool {
	return d.OrderingCmp(a, b) > 0
}

/*
AreAdjacent returns true if a and b are consecutive in the domain (NextFn(a)==b or NextFn(b)==a).
*/
func (d *DiscreteDomain[TDomainElement]) AreAdjacent(a, b TDomainElement) bool {
	if n, ok := d.NextFn(a); ok && d.OrderingCmp(n, b) == 0 {
		return true
	}
	if n, ok := d.NextFn(b); ok && d.OrderingCmp(n, a) == 0 {
		return true
	}
	return false
}

/*
MustNext returns the successor of v. Panics if v is the domain maximum (no next value).
*/
func (d *DiscreteDomain[TDomainElement]) MustNext(v TDomainElement) TDomainElement {
	next, ok := d.NextFn(v)
	if !ok {
		panic("domain overflow: attempted to get Next of Max")
	}
	return next
}

/*
MustPrevious returns the predecessor of v. Panics if v is the domain minimum (no previous value).
*/
func (d *DiscreteDomain[TDomainElement]) MustPrevious(v TDomainElement) TDomainElement {
	next, ok := d.PreviousFn(v)
	if !ok {
		panic("domain underflow: attempted to get Previous of Min")
	}
	return next
}

/*
DiscreteDomainRuneCreate returns the canonical discrete domain for Unicode code points (rune).

Min is 0, Max is unicode.MaxRune. Surrogate halves (U+D800–U+DFFF) are skipped in NextFn/PreviousFn
so that iteration follows valid Unicode scalar values. Use for rune-based lexers and pattern compilation.

Time complexity: O(1). Space complexity: O(1).
*/
func DiscreteDomainRuneCreate() *DiscreteDomain[rune] {
	const (
		surrogateMin = 0xD800
		surrogateMax = 0xDFFF
	)

	return &DiscreteDomain[rune]{
		Min: 0,
		Max: unicode.MaxRune,
		NextFn: func(current rune) (rune, bool) {
			if current >= unicode.MaxRune {
				return current, false
			}
			next := current + 1
			if next >= surrogateMin && next <= surrogateMax {
				return surrogateMax + 1, true
			}
			return next, true
		},
		PreviousFn: func(current rune) (rune, bool) {
			if current <= 0 {
				return current, false
			}
			prev := current - 1
			if prev >= surrogateMin && prev <= surrogateMax {
				return surrogateMin - 1, true
			}
			return prev, true
		},
		OrderingCmp: func(a, b rune) int {
			if a < b {
				return -1
			}
			if a > b {
				return 1
			}
			return 0
		},
	}
}

/*
DiscreteDomainByteCreate returns the discrete domain for bytes (0–255).

Use for byte-based pattern or lexer observation spaces.

Time complexity: O(1). Space complexity: O(1).
*/
func DiscreteDomainByteCreate() *DiscreteDomain[byte] {
	return &DiscreteDomain[byte]{
		Min: 0,
		Max: math.MaxUint8,
		NextFn: func(c byte) (byte, bool) {
			if c >= math.MaxUint8 { // this is logically impossible, but we keep it for strictness.
				return c, false
			}
			return c + 1, true
		},
		PreviousFn: func(c byte) (byte, bool) {
			if c <= 0 {
				return c, false
			}
			return c - 1, true
		},
		OrderingCmp: func(a, b byte) int {
			return int(a) - int(b)
		},
	}
}

/*
DiscreteDomainIntCreate returns the discrete domain for int (math.MinInt to math.MaxInt).

Use for integer observation spaces in pattern or lexer compilation.

Time complexity: O(1). Space complexity: O(1).
*/
func DiscreteDomainIntCreate() *DiscreteDomain[int] {
	return &DiscreteDomain[int]{
		Min: math.MinInt,
		Max: math.MaxInt,
		NextFn: func(c int) (int, bool) {
			if c == math.MaxInt {
				return c, false
			}
			return c + 1, true
		},
		PreviousFn: func(c int) (int, bool) {
			if c == math.MinInt {
				return c, false
			}
			return c - 1, true
		},
		OrderingCmp: func(a, b int) int {
			if a < b {
				return -1
			}
			if a > b {
				return 1
			}
			return 0
		},
	}
}
