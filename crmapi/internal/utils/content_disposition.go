package utils

import (
	"net/url"
	"regexp"
	"strings"
)

var (
	contentDispositionFilenameUTF8 = regexp.MustCompile(`filename\*=UTF-8''([^;]+)`)
	contentDispositionFilenameQ    = regexp.MustCompile(`filename="([^"]+)"`)
	contentDispositionFilenameBare = regexp.MustCompile(`filename=([^;]+)`)
)

// ParseContentDispositionFilename extracts filename from Content-Disposition header.
// Returns empty string if filename cannot be determined.
func ParseContentDispositionFilename(value string) string {
	if strings.TrimSpace(value) == "" {
		return ""
	}

	if m := contentDispositionFilenameUTF8.FindStringSubmatch(value); len(m) == 2 {
		decoded, err := url.QueryUnescape(m[1])
		if err == nil {
			return decoded
		}
		return m[1]
	}

	if m := contentDispositionFilenameQ.FindStringSubmatch(value); len(m) == 2 {
		return m[1]
	}

	if m := contentDispositionFilenameBare.FindStringSubmatch(value); len(m) == 2 {
		return strings.TrimSpace(m[1])
	}

	return ""
}
