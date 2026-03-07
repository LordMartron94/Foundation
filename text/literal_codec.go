package text

import (
	"fmt"
	"strings"
	"unicode"
)

/*
HexEscapeStyle dictates how non-printable Unicode characters are encoded.
Different contexts (Go source code vs. PCRE regex) require different representations.
*/
type HexEscapeStyle uint8

const (
	/* HexStyleGo formats non-printables as \uXXXX or \UXXXXXXXX. */
	HexStyleGo HexEscapeStyle = iota

	/* HexStyleBraced formats non-printables as \x{XXXX} (common in PCRE/Rust regex). */
	HexStyleBraced
)

/*
EscapeConfig defines the exact behavior of an Escaper instance.
It allows granular control over which boundary characters are escaped.
*/
type EscapeConfig struct {
	EscapeSingleQuote bool
	EscapeDoubleQuote bool
	EscapeBacktick    bool
	// SpecialChars is a string of exact runes to backslash-escape (e.g., "{}[]()+-*?^$.|\\")
	SpecialChars string
	HexStyle     HexEscapeStyle
}

/*
Escaper provides a high-performance, reusable text encoding engine.
It pre-computes an internal fast-lookup table for special characters,
making it safe and efficient for continuous use across a compiler pipeline.
*/
type Escaper struct {
	cfg       EscapeConfig
	isSpecial [128]bool // O(1) fast-path lookup for ASCII special characters
}

/*
NewEscaper initializes an Escaper with the provided configuration.
*/
func NewEscaper(cfg EscapeConfig) *Escaper {
	e := &Escaper{cfg: cfg}
	for _, r := range cfg.SpecialChars {
		if r < 128 {
			e.isSpecial[r] = true
		}
	}
	return e
}

// ----------------------------------------------------------------- PRESETS

var (
	/* GoStringEscaper safely escapes contents for inside a double-quoted "..." string. */
	GoStringEscaper = NewEscaper(EscapeConfig{
		EscapeDoubleQuote: true,
		HexStyle:          HexStyleGo,
	})

	/* GoCharEscaper safely escapes contents for inside a single-quoted '...' literal. */
	GoCharEscaper = NewEscaper(EscapeConfig{
		EscapeSingleQuote: true,
		HexStyle:          HexStyleGo,
	})

	/* RegexLiteralEscaper escapes characters that hold semantic meaning in standard RegEx. */
	RegexLiteralEscaper = NewEscaper(EscapeConfig{
		SpecialChars: `\.+*?()|[]{}^$"'`,
		HexStyle:     HexStyleBraced,
	})

	/* RegexClassEscaper escapes characters that hold semantic meaning inside a RegEx class [...]. */
	RegexClassEscaper = NewEscaper(EscapeConfig{
		SpecialChars: `[\]^"'-`,
		HexStyle:     HexStyleBraced,
	})
)

// ----------------------------------------------------------------- ESCAPING

/*
Escape safely encodes an entire string based on the Escaper's configuration.
*/
func (e *Escaper) Escape(s string) string {
	var out strings.Builder
	out.Grow(len(s) * 2)

	for _, r := range s {
		out.WriteString(e.EscapeRune(r))
	}
	return out.String()
}

/*
EscapeRune formats a single rune into its safe text representation.
*/
func (e *Escaper) EscapeRune(r rune) string {
	// 1. Handle Standard Control Escapes
	switch r {
	case '\n':
		return `\n`
	case '\r':
		return `\r`
	case '\t':
		return `\t`
	case '\\':
		return `\\`
	case '\'':
		if e.cfg.EscapeSingleQuote {
			return `\'`
		}
	case '"':
		if e.cfg.EscapeDoubleQuote {
			return `\"`
		}
	case '`':
		if e.cfg.EscapeBacktick {
			return "\\`"
		}
	}

	// 2. Handle Configured Special Characters
	if r < 128 && e.isSpecial[r] {
		return `\` + string(r)
	}

	// 3. Handle Non-Printables (Hex Encoding)
	if !IsPrintable(r) {
		return e.formatHex(r)
	}

	// 4. Default: Return the raw character
	return string(r)
}

func (e *Escaper) formatHex(r rune) string {
	if e.cfg.HexStyle == HexStyleBraced {
		return fmt.Sprintf(`\x{%X}`, r)
	}

	// HexStyleGo
	if r <= 0xFFFF {
		return fmt.Sprintf(`\u%04X`, r)
	}
	return fmt.Sprintf(`\U%08X`, r)
}

/*
IsPrintable determines if a rune is safe to print directly in source code.
It considers standard ASCII printables and valid Unicode printables,
expressly rejecting control characters, zero-width spaces, and surrogates.
*/
func IsPrintable(r rune) bool {
	if r >= 0x20 && r != 0x7F && r <= 0x7E {
		return true
	}
	return unicode.IsPrint(r)
}

// ----------------------------------------------------------------- UNESCAPING

/*
Unescape decodes escape sequences in the inner content of a quoted literal.
Supports \', \", \\, \n, \r, \t, \xXX, \uXXXX, \UXXXXXXXX.
*/
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

/*
decodeHexSequence safely parses a fixed-length hexadecimal string into a rune.
Returns false if the string is too short or contains invalid hex characters.
*/
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

	// Guard against invalid Unicode code points (Surrogates)
	if v > unicode.MaxRune || (0xD800 <= v && v <= 0xDFFF) {
		return unicode.ReplacementChar, true
	}

	return v, true
}
