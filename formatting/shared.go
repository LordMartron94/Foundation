package formatting

/*
FormatSliceOptions configures the behavior of format slice functions.

All fields are optional. Zero values produce sensible defaults.

This struct is intentionally explicit instead of using many variadic arguments —
clarity and predictability beat cleverness in systems code.
*/
type FormatSliceOptions[TElement any] struct {
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
	Escape func(TElement) string

	// Optional per-item formatter.
	// Allows mapping, annotation, coloring, etc.
	FormatItem func(index int, value TElement) string

	// If true, prefixes each element with its index.
	// Example: "0: foo, 1: bar"
	IncludeIndex bool

	// Separator used between index and value.
	// Default when empty: ": "
	IndexSeparator string

	// SkipIf allows conditional exclusion of elements.
	// Return true to omit the element.
	SkipIf func(index int, value TElement) bool

	// Prealloc sets an explicit initial capacity for the builder.
	// Use when output size is approximately known.
	Prealloc int
}
