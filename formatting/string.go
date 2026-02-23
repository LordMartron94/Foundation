package formatting

import (
	"strconv"
	"strings"
)

/*
FormatStringSlice formats a slice of strings into a single string using a high-performance
strings.Builder pipeline with flexible, composable formatting options.

The function is designed for:

  - zero unnecessary allocations
  - predictable output
  - extensible formatting behavior
  - high-throughput debug/log/serialization use

It supports:

  - custom separators (", ", " | ", "\n", etc.)
  - optional prefix and suffix (e.g. "[", "]")
  - per-element formatting hooks
  - optional quoting
  - optional escaping
  - optional index inclusion
  - conditional element skipping
  - capacity pre-allocation for performance

────────────────────────────────────────────────────────────
Default behavior (no options):

	["a", "b", "c"] → "a, b, c"

────────────────────────────────────────────────────────────
Example:

	FormatStringSlice(
	    []string{"foo", "bar"},
	    FormatStringSliceOptions{
	        Prefix: "[",
	        Suffix: "]",
	        Quote:  true,
	    },
	)

Result:

	["foo", "bar"]

────────────────────────────────────────────────────────────
Performance characteristics:

  - O(n) time
  - single growing buffer
  - no intermediate slices
  - minimal branching in hot loop

This function is appropriate for:

  - debug renderers
  - serialization helpers
  - logging
  - DSL emitters
  - diagnostics tooling
*/
func FormatStringSlice(
	items []string,
	cfg FormatStringSliceOptions,
) string {
	if cfg.Separator == "" {
		cfg.Separator = ", "
	}

	var b strings.Builder

	// ------------------------------------------------------------
	// Capacity hint (best-effort, avoids repeated growth)
	// ------------------------------------------------------------

	if cfg.Prealloc > 0 {
		b.Grow(cfg.Prealloc)
	} else {
		// heuristic: average 8 chars per entry + separators
		b.Grow(len(items) * (8 + len(cfg.Separator)))
	}

	// ------------------------------------------------------------
	// Prefix
	// ------------------------------------------------------------

	if cfg.Prefix != "" {
		b.WriteString(cfg.Prefix)
	}

	first := true

	for i, s := range items {
		if cfg.SkipIf != nil && cfg.SkipIf(i, s) {
			continue
		}

		if !first {
			b.WriteString(cfg.Separator)
		}
		first = false

		// --------------------------------------------------------
		// Index prefix
		// --------------------------------------------------------

		if cfg.IncludeIndex {
			b.WriteString(strconv.Itoa(i))
			b.WriteString(cfg.IndexSeparator)
		}

		// --------------------------------------------------------
		// Element formatting
		// --------------------------------------------------------

		if cfg.FormatItem != nil {
			s = cfg.FormatItem(i, s)
		}

		if cfg.Escape != nil {
			s = cfg.Escape(s)
		}

		if cfg.Quote {
			b.WriteByte('"')
			b.WriteString(s)
			b.WriteByte('"')
		} else {
			b.WriteString(s)
		}
	}

	// ------------------------------------------------------------
	// Suffix
	// ------------------------------------------------------------

	if cfg.Suffix != "" {
		b.WriteString(cfg.Suffix)
	}

	return b.String()
}

/*
FormatStringSliceOptions configures the behavior of FormatStringSlice.

All fields are optional. Zero values produce sensible defaults.

This struct is intentionally explicit instead of using many variadic arguments —
clarity and predictability beat cleverness in systems code.
*/
type FormatStringSliceOptions struct {
	// Separator inserted between elements.
	// Default: ", "
	Separator string

	// Optional prefix written before first element.
	// Example: "[" for JSON-like output
	Prefix string

	// Optional suffix written after last element.
	// Example: "]" for JSON-like output
	Suffix string

	// If true, wraps each element in double quotes.
	Quote bool

	// Escape function applied before quoting (if provided).
	// Example: strings.ReplaceAll(s, `"`, `\"`)
	Escape func(string) string

	// Optional per-item formatter.
	// Allows mapping, annotation, coloring, etc.
	FormatItem func(index int, value string) string

	// If true, prefixes each element with its index.
	// Example: "0: foo, 1: bar"
	IncludeIndex bool

	// Separator used between index and value.
	// Default when empty: ": "
	IndexSeparator string

	// SkipIf allows conditional exclusion of elements.
	// Return true to omit the element.
	SkipIf func(index int, value string) bool

	// Prealloc sets an explicit initial capacity for the builder.
	// Use when output size is approximately known.
	Prealloc int
}
