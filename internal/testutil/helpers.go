package testutil

// BoolPtr returns a pointer to the given bool value
func BoolPtr(b bool) *bool {
	return &b
}

// StringPtr returns a pointer to the given string value
func StringPtr(s string) *string {
	return &s
}
