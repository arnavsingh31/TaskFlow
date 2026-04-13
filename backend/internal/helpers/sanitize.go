package helpers

import (
	"regexp"
	"strings"
)

var htmlTagPattern = regexp.MustCompile(`<[^>]*>`)

// ContainsHTML checks if a string contains HTML tags
func ContainsHTML(s string) bool {
	return htmlTagPattern.MatchString(s)
}

// ContainsHTMLPtr checks a *string for HTML tags
func ContainsHTMLPtr(s *string) bool {
	if s == nil {
		return false
	}
	return ContainsHTML(*s)
}

// TrimString trims whitespace from a string
func TrimString(s string) string {
	return strings.TrimSpace(s)
}

// TrimPtr trims whitespace from a *string if non-nil
func TrimPtr(s *string) *string {
	if s == nil {
		return nil
	}
	trimmed := TrimString(*s)
	return &trimmed
}
