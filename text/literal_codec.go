package text

import (
	"fmt"
	"strings"
	"unicode"
)

// Unescape decodes escape sequences in the inner content of a quoted literal.
// Supports \', \", \\, \n, \r, \t, \xXX, \uXXXX, \UXXXXXXXX.
func Unescape(inner string) string {
	if !strings.ContainsRune(inner, '\\') {
		return inner
	}

	var out strings.Builder
	out.Grow(len(inner))

	i := 0
	for i < len(inner) {
		if inner[i] != '\\' {
			out.WriteByte(inner[i])
			i++
			continue
		}

		i++ // Move past the backslash
		if i >= len(inner) {
			out.WriteByte('\\') // Handle trailing backslash gracefully
			break
		}

		next := inner[i]
		switch next {
		case 'x':
			if r, ok := decodeHexSequence(inner[i+1:], 2); ok {
				out.WriteRune(r)
				i += 3 // 'x' + 2 digits
				continue
			}
		case 'u':
			if r, ok := decodeHexSequence(inner[i+1:], 4); ok {
				out.WriteRune(r)
				i += 5 // 'u' + 4 digits
				continue
			}
		case 'U':
			if r, ok := decodeHexSequence(inner[i+1:], 8); ok {
				out.WriteRune(r)
				i += 9 // 'U' + 8 digits
				continue
			}
		case '\'', '"', '\\':
			out.WriteByte(next)
		case 'n':
			out.WriteByte('\n')
		case 'r':
			out.WriteByte('\r')
		case 't':
			out.WriteByte('\t')
		default:
			// Unknown escape: keep the character but drop the backslash
			out.WriteByte(next)
		}
		i++
	}
	return out.String()
}

// Escape encodes a string for use inside quoted literals.
// It encodes non-printable characters as \uXXXX or \UXXXXXXXX.
func Escape(s string) string {
	var out strings.Builder
	out.Grow(len(s) * 2)

	for _, r := range s {
		switch r {
		case '\\':
			out.WriteString(`\\`)
		case '\'':
			out.WriteString(`\'`)
		case '"':
			out.WriteString(`\"`)
		case '\n':
			out.WriteString(`\n`)
		case '\r':
			out.WriteString(`\r`)
		case '\t':
			out.WriteString(`\t`)
		default:
			if !isPrintRune(r) {
				if r <= 0xFFFF {
					fmt.Fprintf(&out, `\u%04X`, r)
				} else {
					fmt.Fprintf(&out, `\U%08X`, r)
				}
			} else {
				out.WriteRune(r)
			}
		}
	}
	return out.String()
}

func EscapeRuneForRegexLiteral(r rune) string {
	switch r {
	case '\n':
		return `\n`
	case '\r':
		return `\r`
	case '\t':
		return `\t`
	case '\\', '.', '+', '*', '?', '(', ')', '|', '[', ']', '{', '}', '^', '$', '"', '\'':
		return `\` + string(r)
	}
	if !isPrintRune(r) {
		return fmt.Sprintf(`\x{%X}`, r)
	}
	return string(r)
}

func EscapeRuneForRegexClass(r rune) string {
	switch r {
	case '[', '\\', '-', ']', '^', '"', '\'':
		return `\` + string(r)
	case '\n':
		return `\n`
	case '\r':
		return `\r`
	case '\t':
		return `\t`
	}
	if !isPrintRune(r) {
		return fmt.Sprintf(`\x{%04X}`, r)
	}
	return string(r)
}

func isPrintRune(r rune) bool {
	if r >= 0x20 && r != 0x7F && r <= 0x7E {
		return true
	}
	return unicode.IsPrint(r)
}

func decodeHexSequence(s string, n int) (rune, bool) {
	if len(s) < n {
		return 0, false
	}

	var v rune
	for i := 0; i < n; i++ {
		c := s[i]
		v <<= 4
		switch {
		case '0' <= c && c <= '9':
			v |= rune(c - '0')
		case 'a' <= c && c <= 'f':
			v |= rune(c - 'a' + 10)
		case 'A' <= c && c <= 'F':
			v |= rune(c - 'A' + 10)
		default:
			return 0, false
		}
	}

	// Guard against invalid Unicode code points
	if v > unicode.MaxRune || (0xD800 <= v && v <= 0xDFFF) {
		return unicode.ReplacementChar, true
	}

	return v, true
}
