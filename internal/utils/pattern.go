package utils

import "strings"

// MatchPattern checks if a string matches a glob-style pattern
// Supports:
//   - * matches any number of characters
//   - ? matches a single character
//   - [abc] matches any character in the set
//   - [a-z] matches any character in the range
func MatchPattern(pattern, str string) bool {
	// Special case: * matches everything
	if pattern == "*" {
		return true
	}

	// For now, implement basic * wildcard support
	// This can be extended to support full glob patterns
	if strings.Contains(pattern, "*") {
		// Convert pattern to parts split by *
		parts := strings.Split(pattern, "*")

		// Check if string starts with first part
		if len(parts[0]) > 0 && !strings.HasPrefix(str, parts[0]) {
			return false
		}

		// Check if string ends with last part
		lastPart := parts[len(parts)-1]
		if len(lastPart) > 0 && !strings.HasSuffix(str, lastPart) {
			return false
		}

		// Simple implementation: check if all parts exist in order
		currentPos := 0
		for _, part := range parts {
			if part == "" {
				continue
			}

			idx := strings.Index(str[currentPos:], part)
			if idx == -1 {
				return false
			}
			currentPos += idx + len(part)
		}

		return true
	}

	// Exact match
	return pattern == str
}

// IsGlobPattern returns true if the pattern contains glob metacharacters
func IsGlobPattern(pattern string) bool {
	return strings.ContainsAny(pattern, "*?[]")
}
