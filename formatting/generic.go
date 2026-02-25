package formatting

import (
	"strconv"
	"strings"
)

/*
FormatSlice formats a slice of arbitrary elements into a single string using a
high-performance strings.Builder pipeline with flexible, composable formatting.

Unlike FormatStringSlice, this function is fully generic and makes no assumptions
about the underlying element type — all rendering flows through the provided
FormatItem hook.

Design goals:

  - zero reflection
  - zero intermediate allocations
  - predictable hot loop
  - extensible formatting
  - suitable for diagnostics, debug printers, and serialization

────────────────────────────────────────────────────────────
Required:

	cfg.FormatItem must be provided.

This avoids any implicit fmt-based slow paths and keeps performance explicit.

────────────────────────────────────────────────────────────
Default behavior (with simple formatter):

	FormatSlice(
	    []int{1,2,3},
	    FormatSliceOptions[int]{
	        FormatItem: func(_ int, v int) string {
	            return strconv.Itoa(v)
	        },
	    },
	)

→ "1, 2, 3"
*/
func FormatSlice[TElement any](
	items []TElement,
	cfg FormatSliceOptions[TElement],
) string {
	if cfg.FormatItem == nil {
		panic("FormatSlice: FormatItem is required for generic formatting")
	}

	if cfg.Separator == "" {
		cfg.Separator = ", "
	}

	if cfg.IndexSeparator == "" {
		cfg.IndexSeparator = ": "
	}

	var b strings.Builder

	// ------------------------------------------------------------
	// Capacity hint
	// ------------------------------------------------------------

	if cfg.Prealloc > 0 {
		b.Grow(cfg.Prealloc)
	} else {
		// conservative heuristic: 12 chars per element + separators
		b.Grow(len(items) * (12 + len(cfg.Separator)))
	}

	// ------------------------------------------------------------
	// Prefix
	// ------------------------------------------------------------

	if cfg.Prefix != "" {
		b.WriteString(cfg.Prefix)
	}

	first := true

	for i, v := range items {
		if cfg.SkipIf != nil && cfg.SkipIf(i, v) {
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

		s := cfg.FormatItem(i, v)

		if cfg.Escape != nil {
			s = cfg.Escape(v)
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
