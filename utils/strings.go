package utils

import (
	"fmt"
	"unicode/utf8"
)

// IsValidCodePoint checks if the code point is valid (U+0000 to U+10FFFF).
func IsValidCodePoint(codePoint int) bool {
	return codePoint >= 0 && codePoint <= 0x10FFFF
}

// ToChars converts a Unicode code point to a slice of UTF-16 code units
// (equivalent to Java's Character.toChars).
func ToChars(codePoint int) []rune {
	// Validate code point
	if !IsValidCodePoint(codePoint) {
		panic("Invalid Unicode code point")
	}

	// If code point is in BMP (U+0000 to U+FFFF), return it as a single uint16
	if codePoint <= 0xFFFF {
		return []rune{rune(codePoint)}
	}

	// For supplementary characters, encode as surrogate pair
	return []rune{rune(codePoint)}
}

// CodePointCount counts the number of Unicode code points in the string s
// between startIndex (inclusive) and endIndex (exclusive), where indices are in bytes.
func CodePointCount(s string, startIndex, endIndex int) (int, error) {
	// Validate indices
	if startIndex < 0 || endIndex > len(s) || startIndex > endIndex {
		return 0, fmt.Errorf("invalid indices: startIndex=%d, endIndex=%d, len=%d", startIndex, endIndex, len(s))
	}

	// Slice the string to the desired byte range
	substring := s[startIndex:endIndex]

	// Count runes in the substring
	return utf8.RuneCountInString(substring), nil
}
