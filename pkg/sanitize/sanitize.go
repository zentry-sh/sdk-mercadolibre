package sanitize

import (
	"strings"
	"unicode"
)

func String(s string) string {
	return strings.Map(func(r rune) rune {
		if r == 0 {
			return -1
		}
		return r
	}, strings.TrimSpace(s))
}

func ID(s string) string {
	s = String(s)
	var b strings.Builder
	b.Grow(len(s))
	for _, r := range s {
		if unicode.IsLetter(r) || unicode.IsDigit(r) || r == '-' || r == '_' {
			b.WriteRune(r)
		}
	}
	return b.String()
}

func Email(s string) string {
	return strings.ToLower(String(s))
}

func CountryCode(s string) string {
	s = strings.TrimSpace(s)
	if len(s) != 2 {
		return s
	}
	return strings.ToUpper(s)
}

func CurrencyCode(s string) string {
	s = strings.TrimSpace(s)
	if len(s) != 3 {
		return s
	}
	return strings.ToUpper(s)
}
