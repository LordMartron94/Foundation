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
	cfg FormatSliceOptions[string],
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
