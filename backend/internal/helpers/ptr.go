package helpers

// StringPtr creates a pointer to a string — use for optional fields only
func StringPtr(s string) *string {
	return &s
}
