package text

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"unicode"
)

// TextSplitPathPrefix splits text by separator and returns the prefix (all parts except last)
// and suffix (last part). If separator is not found, prefix is empty and suffix is the full text.
//
// Example:
//   TextSplitPathPrefix("1/2/3", "/") returns ("1/2", "3")
//   TextSplitPathPrefix("test", "/") returns ("", "test")
func TextSplitPathPrefix(text, separator string) (prefix, suffix string) {
	parts := strings.Split(text, separator)
	if len(parts) <= 1 {
		return "", text
	}
	prefix = strings.Join(parts[:len(parts)-1], separator)
	suffix = parts[len(parts)-1]
	return prefix, suffix
}

// TextExtractUntil extracts text from the beginning until the first occurrence of delimiter.
// Returns the text before delimiter, text after delimiter, and whether delimiter was found.
//
// Example:
//   TextExtractUntil("hello|world", "|") returns ("hello", "world", true)
//   TextExtractUntil("hello", "|") returns ("hello", "", false)
func TextExtractUntil(text, delimiter string) (before, after string, found bool) {
	idx := strings.Index(text, delimiter)
	if idx < 0 {
		return text, "", false
	}
	return text[:idx], text[idx+len(delimiter):], true
}

// TextStripSuffixes strips multiple suffixes from text in order.
// Stops at the first suffix that is found and stripped.
//
// Example:
//   TextStripSuffixes("title#location|override", "#", "|") returns "title"
//   TextStripSuffixes("title|override", "#", "|") returns "title"
func TextStripSuffixes(text string, suffixes ...string) string {
	result := text
	for _, suffix := range suffixes {
		if idx := strings.Index(result, suffix); idx >= 0 {
			result = result[:idx]
		}
	}
	return result
}

// TextNormalizeEmptyLines replaces empty lines (lines containing only whitespace)
// with the specified marker to preserve them during parsing.
//
// Example:
//   TextNormalizeEmptyLines("line1\n\nline2", "EMPTY") returns "line1\nEMPTY\nline2"
func TextNormalizeEmptyLines(content, marker string) string {
	out := strings.Builder{}
	lines := strings.Split(content, "\n")

	for i := 0; i < len(lines); i++ {
		line := lines[i]

		if strings.TrimSpace(line) == "" {
			out.WriteString(marker)
			out.WriteByte('\n')
			continue
		}

		out.WriteString(line)
		if i < len(lines)-1 {
			out.WriteByte('\n')
		}
	}
	return out.String()
}

// TextProcessLines processes each line of content using the provided processor function.
// The processor receives the line content and its index (0-based).
//
// Example:
//   TextProcessLines("a\nb\nc", func(line string, idx int) string {
//       return strings.ToUpper(line)
//   }) returns "A\nB\nC"
func TextProcessLines(content string, processor func(line string, index int) string) string {
	lines := strings.Split(content, "\n")
	processed := make([]string, len(lines))
	for i, line := range lines {
		processed[i] = processor(line, i)
	}
	return strings.Join(processed, "\n")
}

// TextSanitizeFilename converts text into a valid filename by replacing
// any character not in the valid character set with an underscore.
//
// The validChars parameter should be a regex character class pattern (e.g., "a-zA-Z0-9 ._-").
//
// Example:
//   TextSanitizeFilename("1/1a", "a-zA-Z0-9 ._-") returns "1_1a"
//   TextSanitizeFilename("abc@123", "a-zA-Z0-9 ._-") returns "abc_123"
func TextSanitizeFilename(text, validChars string) string {
	pattern := fmt.Sprintf(`[^%s]`, regexp.QuoteMeta(validChars))
	invalidCharRegex := regexp.MustCompile(pattern)
	return invalidCharRegex.ReplaceAllString(text, "_")
}

// TextIncrementTrailingDigits increments trailing digits in text.
// If no trailing digits are found, appends "1".
//
// Example:
//   TextIncrementTrailingDigits("abc") returns ("abc1", nil)
//   TextIncrementTrailingDigits("abc1") returns ("abc2", nil)
//   TextIncrementTrailingDigits("abc99") returns ("abc100", nil)
func TextIncrementTrailingDigits(text string) (string, error) {
	if len(text) == 0 {
		return "1", nil
	}

	runes := []rune(text)

	// Find the start of the trailing numeric sequence
	firstDigitIdx := -1
	for i := len(runes) - 1; i >= 0; i-- {
		if unicode.IsDigit(runes[i]) {
			firstDigitIdx = i
		} else {
			break
		}
	}

	// If no trailing digits found, append "1"
	if firstDigitIdx == -1 {
		return text + "1", nil
	}

	// Extract the numeric sequence
	lastDigitIdx := len(runes) - 1
	numberStr := string(runes[firstDigitIdx : lastDigitIdx+1])
	number, err := strconv.Atoi(numberStr)
	if err != nil {
		return "", fmt.Errorf("failed to parse trailing digits: %w", err)
	}

	// Reconstruct with incremented number
	prefix := string(runes[:firstDigitIdx])
	return fmt.Sprintf("%s%d", prefix, number+1), nil
}

// TextIncrementLetterSequence increments a sequence of letters like a base-26 counter.
//
// Examples:
//   TextIncrementLetterSequence("a") returns "b"
//   TextIncrementLetterSequence("z") returns "za"
//   TextIncrementLetterSequence("zz") returns "zza"
func TextIncrementLetterSequence(letters string) string {
	if len(letters) == 0 {
		return "a"
	}

	// Convert to lowercase for processing
	runes := []rune(letters)
	result := make([]rune, len(runes))
	for i, r := range runes {
		result[i] = unicode.ToLower(r)
	}

	// Special case: if the last character is 'z', append 'a' to the entire sequence
	if len(result) > 0 && result[len(result)-1] == 'z' {
		return string(result) + "a"
	}

	// Increment from right to left with carry
	carry := true
	for i := len(result) - 1; i >= 0 && carry; i-- {
		if result[i] == 'z' {
			result[i] = 'a'
			carry = true
		} else {
			result[i]++
			carry = false
		}
	}

	// If we still have a carry, prepend 'a'
	if carry {
		return "a" + string(result)
	}

	return string(result)
}

// TextJoinPathPrefix joins prefix and suffix with separator.
// If prefix is empty, returns suffix only.
//
// Example:
//   TextJoinPathPrefix("1/2", "3", "/") returns "1/2/3"
//   TextJoinPathPrefix("", "3", "/") returns "3"
func TextJoinPathPrefix(prefix, suffix, separator string) string {
	if prefix == "" {
		return suffix
	}
	return prefix + separator + suffix
}

