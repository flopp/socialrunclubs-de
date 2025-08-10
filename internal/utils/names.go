package utils

import (
	"strings"
	"unicode"

	"golang.org/x/text/runes"
	"golang.org/x/text/transform"
	"golang.org/x/text/unicode/norm"
)

// specialCharReplacer is used to replace special characters in names.
var specialCharReplacer = strings.NewReplacer(
	"ä", "ae",
	"ö", "oe",
	"ü", "ue",
	"ß", "ss",
	" ", "-",
	".", "-",
	"'", "-",
	"\"", "-",
	"(", "-",
	")", "-",
)

func SanitizeName(s string) string {
	// lowercase
	sanitized := strings.ToLower(s)

	// replace special characters
	sanitized = specialCharReplacer.Replace(sanitized)

	// remove all other special characters
	result, _, err := transform.String(transform.Chain(norm.NFD, runes.Remove(runes.In(unicode.Mn))), sanitized)
	if err != nil {
		result = sanitized
	}

	// remove all non-alphanumeric characters & replace with '-'; avoid leading, trailing and consecutive '-'
	var builder strings.Builder
	builder.Grow(len(result)) // Pre-allocate capacity
	needSep := false
	for _, char := range result {
		if ('a' <= char && char <= 'z') || ('0' <= char && char <= '9') {
			if needSep && builder.Len() > 0 {
				builder.WriteByte('-')
			}
			needSep = false
			builder.WriteByte(byte(char))
		} else {
			needSep = true
		}
	}

	return builder.String()
}
