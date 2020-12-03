package encode

import (
	"encoding/base64"
	"fmt"
	"strings"
	"unicode"
)

// UTF8 encodes s into a base64 encoded ASCII string that is understandable by mail clients.
func UTF8(s string) string {
	return fmt.Sprintf("=?utf-8?B?%s?=", base64.StdEncoding.EncodeToString([]byte(s)))
}

// ToASCII returns s with all non-ASCII characters replaced by unicode.MaxASCII.
func ToASCII(s string) string {
	return strings.Map(func(r rune) rune {
		if r > unicode.MaxASCII {
			return unicode.MaxASCII
		}
		return r
	}, s)
}
