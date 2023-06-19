package helpers

import "strings"

// Format a key
func FormatKey(parts ...string) string {
	return strings.Join(parts, ":")
}
