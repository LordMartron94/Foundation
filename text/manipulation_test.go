package text

import (
	"strings"
	"testing"
)

func TestTextSplitPathPrefix(t *testing.T) {
	tests := []struct {
		text           string
		separator      string
		expectedPrefix string
		expectedSuffix string
	}{
		{"1/2/3", "/", "1/2", "3"},
		{"test", "/", "", "test"},
		{"a/b", "/", "a", "b"},
		{"", "/", "", ""},
		{"path/to/file", "/", "path/to", "file"},
		{"no-separator", "/", "", "no-separator"},
	}

	for _, tt := range tests {
		prefix, suffix := TextSplitPathPrefix(tt.text, tt.separator)
		if prefix != tt.expectedPrefix || suffix != tt.expectedSuffix {
			t.Errorf("TextSplitPathPrefix(%q, %q) = (%q, %q), expected (%q, %q)",
				tt.text, tt.separator, prefix, suffix, tt.expectedPrefix, tt.expectedSuffix)
		}
	}
}

func TestTextExtractUntil(t *testing.T) {
	tests := []struct {
		text           string
		delimiter      string
		expectedBefore string
		expectedAfter  string
		expectedFound  bool
	}{
		{"hello|world", "|", "hello", "world", true},
		{"hello", "|", "hello", "", false},
		{"a#b", "#", "a", "b", true},
		{"test", "|", "test", "", false},
		{"", "|", "", "", false},
	}

	for _, tt := range tests {
		before, after, found := TextExtractUntil(tt.text, tt.delimiter)
		if before != tt.expectedBefore || after != tt.expectedAfter || found != tt.expectedFound {
			t.Errorf("TextExtractUntil(%q, %q) = (%q, %q, %v), expected (%q, %q, %v)",
				tt.text, tt.delimiter, before, after, found, tt.expectedBefore, tt.expectedAfter, tt.expectedFound)
		}
	}
}

func TestTextStripSuffixes(t *testing.T) {
	tests := []struct {
		text     string
		suffixes []string
		expected string
	}{
		{"title#location|override", []string{"#", "|"}, "title"},
		{"title|override", []string{"#", "|"}, "title"},
		{"title#location", []string{"#", "|"}, "title"},
		{"title", []string{"#", "|"}, "title"},
		{"test", []string{}, "test"},
	}

	for _, tt := range tests {
		result := TextStripSuffixes(tt.text, tt.suffixes...)
		if result != tt.expected {
			t.Errorf("TextStripSuffixes(%q, %v) = %q, expected %q",
				tt.text, tt.suffixes, result, tt.expected)
		}
	}
}

func TestTextNormalizeEmptyLines(t *testing.T) {
	tests := []struct {
		content  string
		marker   string
		expected string
	}{
		{"line1\n\nline2", "EMPTY", "line1\nEMPTY\nline2"},
		{"a\nb\nc", "EMPTY", "a\nb\nc"},
		{"line1\n  \nline2", "EMPTY", "line1\nEMPTY\nline2"},
		{"", "EMPTY", ""},
		{"\n\n", "EMPTY", "EMPTY\nEMPTY"},
	}

	for _, tt := range tests {
		result := TextNormalizeEmptyLines(tt.content, tt.marker)
		if result != tt.expected {
			t.Errorf("TextNormalizeEmptyLines(%q, %q) = %q, expected %q",
				tt.content, tt.marker, result, tt.expected)
		}
	}
}

func TestTextProcessLines(t *testing.T) {
	result := TextProcessLines("a\nb\nc", func(line string, idx int) string {
		return strings.ToUpper(line)
	})
	expected := "A\nB\nC"
	if result != expected {
		t.Errorf("TextProcessLines = %q, expected %q", result, expected)
	}
}

func TestTextSanitizeFilename(t *testing.T) {
	tests := []struct {
		text     string
		valid    string
		expected string
	}{
		{"1/1a", "a-zA-Z0-9 ._-", "1_1a"},
		{"abc@123", "a-zA-Z0-9 ._-", "abc_123"},
		{"test-id", "a-zA-Z0-9 ._-", "test-id"},
		{"file.name", "a-zA-Z0-9 ._-", "file.name"},
	}

	for _, tt := range tests {
		result := TextSanitizeFilename(tt.text, tt.valid)
		if result != tt.expected {
			t.Errorf("TextSanitizeFilename(%q, %q) = %q, expected %q",
				tt.text, tt.valid, result, tt.expected)
		}
	}
}

func TestTextIncrementTrailingDigits(t *testing.T) {
	tests := []struct {
		input    string
		expected string
		hasError bool
	}{
		{"abc", "abc1", false},
		{"abc1", "abc2", false},
		{"abc99", "abc100", false},
		{"", "1", false},
		{"1", "2", false},
		{"99", "100", false},
	}

	for _, tt := range tests {
		result, err := TextIncrementTrailingDigits(tt.input)
		if tt.hasError {
			if err == nil {
				t.Errorf("TextIncrementTrailingDigits(%q) expected error, got nil", tt.input)
			}
		} else {
			if err != nil {
				t.Errorf("TextIncrementTrailingDigits(%q) unexpected error: %v", tt.input, err)
			}
			if result != tt.expected {
				t.Errorf("TextIncrementTrailingDigits(%q) = %q, expected %q",
					tt.input, result, tt.expected)
			}
		}
	}
}

func TestTextIncrementLetterSequence(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"a", "b"},
		{"z", "za"},
		{"zz", "zza"},
		{"", "a"},
		{"A", "b"},
		{"Z", "za"},
		{"aa", "ab"},
		{"az", "ba"},
	}

	for _, tt := range tests {
		result := TextIncrementLetterSequence(tt.input)
		if result != tt.expected {
			t.Errorf("TextIncrementLetterSequence(%q) = %q, expected %q",
				tt.input, result, tt.expected)
		}
	}
}

func TestTextJoinPathPrefix(t *testing.T) {
	tests := []struct {
		prefix   string
		suffix   string
		sep      string
		expected string
	}{
		{"1/2", "3", "/", "1/2/3"},
		{"", "3", "/", "3"},
		{"a", "b", "/", "a/b"},
		{"path/to", "file", "/", "path/to/file"},
	}

	for _, tt := range tests {
		result := TextJoinPathPrefix(tt.prefix, tt.suffix, tt.sep)
		if result != tt.expected {
			t.Errorf("TextJoinPathPrefix(%q, %q, %q) = %q, expected %q",
				tt.prefix, tt.suffix, tt.sep, result, tt.expected)
		}
	}
}

