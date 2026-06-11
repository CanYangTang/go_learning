package stringutil

import "strings"

func IsBlank(s string) bool {
	return strings.TrimSpace(s) == ""
}

func NormalizeSpace(s string) string {
	field := strings.Fields(s)
	return strings.Join(field, " ")
}

func JoinNonEmpty(sep string, parts ...string) string {
	var noEmptyParts []string
	for _, part := range parts {
		normalizedPart := NormalizeSpace(part)
		if IsBlank(normalizedPart) {
			continue
		}
		noEmptyParts = append(noEmptyParts, normalizedPart)
	}
	return strings.Join(noEmptyParts, sep)
}
