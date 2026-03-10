package text

import (
	"regexp"
	"strings"
)

// Rule defines a single, focused sanitization operation.
type Rule interface {
	Apply(input string) string
}

// Sanitizer processes text through a configurable pipeline of predefined rules.
type Sanitizer struct {
	rules []Rule
}

// NewSanitizer constructs a thread-safe sanitization pipeline.
func NewSanitizer(rules ...Rule) *Sanitizer {
	return &Sanitizer{
		rules: rules,
	}
}

// Sanitize executes all configured rules in sequence.
func (s *Sanitizer) Sanitize(input string) string {
	result := input
	for _, rule := range s.rules {
		result = rule.Apply(result)
	}
	return result
}

// regexRule handles pattern-based replacements.
type regexRule struct {
	re          *regexp.Regexp
	replacement string
}

// NewRegexRule creates a pre-compiled regex rule to avoid allocation overhead on every call.
func NewRegexRule(pattern, replacement string) Rule {
	return &regexRule{
		re:          regexp.MustCompile(pattern),
		replacement: replacement,
	}
}

// Apply executes the regex replacement.
func (r *regexRule) Apply(input string) string {
	return r.re.ReplaceAllString(input, r.replacement)
}

// templateRule handles exact string replacements efficiently.
type templateRule struct {
	replacer *strings.Replacer
}

// NewTemplateRule constructs a rule from a map of template variables to their replacements.
func NewTemplateRule(mapping map[string]string) Rule {
	return &templateRule{
		replacer: buildReplacer(mapping),
	}
}

// buildReplacer is a helper to isolate the strings.Replacer construction.
func buildReplacer(mapping map[string]string) *strings.Replacer {
	args := make([]string, 0, len(mapping)*2)
	for k, v := range mapping {
		args = append(args, k, v)
	}
	return strings.NewReplacer(args...)
}

// Apply executes the string replacement template.
func (t *templateRule) Apply(input string) string {
	return t.replacer.Replace(input)
}

// NewIdentifierSanitizer constructs a highly restricted pipeline suitable for
// internal system identifiers. It enforces alphanumeric characters and underscores,
// explicitly avoiding runtime allocations for rule compilation.
func NewIdentifierSanitizer() *Sanitizer {
	return NewSanitizer(
		// 1. Convert any whitespace block into a single underscore
		NewRegexRule(`\s+`, "_"),

		// 2. Strip absolutely everything that isn't a letter, number, or underscore
		NewRegexRule(`[^a-zA-Z0-9_]`, ""),

		// 3. Collapse consecutive underscores to prevent messy outputs (e.g., "my___var")
		NewRegexRule(`_+`, "_"),

		// 4. Clean up the edges by removing leading/trailing underscores
		NewTrimRule("_"),
	)
}

// trimRule handles removal of specific leading and trailing characters.
type trimRule struct {
	cutset string
}

// NewTrimRule creates a rule that strips the specified characters from both ends of the string.
func NewTrimRule(cutset string) Rule {
	return &trimRule{
		cutset: cutset,
	}
}

// Apply executes the string trimming.
func (t *trimRule) Apply(input string) string {
	return strings.Trim(input, t.cutset)
}
