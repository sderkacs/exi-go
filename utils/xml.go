package utils

import (
	"fmt"
	"strings"
)

const (
	XMLWhiteSpaceSpace byte = ' '
	XMLWhiteSpaceNL    byte = '\n'
	XMLWhiteSpaceCR    byte = '\r'
	XMLWhiteSpaceTab   byte = '\t'

	Colon string = ":"
)

func IsWhiteSpace(c byte) bool {
	switch c {
	case XMLWhiteSpaceSpace, XMLWhiteSpaceNL, XMLWhiteSpaceCR, XMLWhiteSpaceTab:
		return true
	default:
		return false
	}
}

func GetLeadingWhiteSpaces(ch []byte, start, length int) (int, error) {
	end := start + length
	if end >= len(ch) {
		return -1, ErrorIndexOutOfBounds
	}

	leadingWS := 0

	for start < end && IsWhiteSpace(ch[start]) {
		start++
		leadingWS++
	}

	return leadingWS, nil
}

func GetTrailingWhiteSpaces(ch []byte, start, length int) (int, error) {
	pos := start + length - 1
	if pos >= len(ch) {
		return -1, ErrorIndexOutOfBounds
	}
	trailingWS := 0

	for pos >= start && IsWhiteSpace(ch[pos]) {
		pos--
		trailingWS++
	}

	return trailingWS, nil
}

func IsWhiteSpaceOnly(s string) bool {
	if len(s) == 0 {
		return false
	}

	chars := []byte(s)
	if !IsWhiteSpace(chars[0]) {
		return false
	}

	end := len(chars)
	start := 1

	for start < end {
		if !IsWhiteSpace(chars[start]) {
			return false
		}
		start++
	}

	return true
}

func IsWhiteSpaceOnlyInRange(ch []byte, start, length int) (bool, error) {
	end := start + length
	if start >= len(ch) || end >= len(ch) {
		return false, ErrorIndexOutOfBounds
	}
	if !IsWhiteSpace(ch[start]) {
		return false, nil
	}

	start++
	for start < end {
		if !IsWhiteSpace(ch[start]) {
			return false, nil
		}
		start++
	}

	return true, nil
}

func GetQualifiedName(localName string, prefix *string) string {
	p := ""
	if prefix != nil {
		p = *prefix
	}

	if p == "" {
		return localName
	} else {
		return fmt.Sprintf("%s:%s", p, localName)
	}
}

func GetPrefixPart(qname string) string {
	index := strings.Index(qname, ":")
	if index >= 0 {
		return qname[:index]
	} else {
		return ""
	}
}

func GetLocalPart(qname string) string {
	index := strings.Index(qname, ":")
	if index < 0 {
		return qname
	} else {
		return qname[index+1:]
	}
}
