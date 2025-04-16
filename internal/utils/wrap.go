package utils

// Wrap truncates a string from the right side, keeping the leftmost characters up to the specified limit.
// If no break words are provided, it defaults to using "..." as the truncation indicator.
// If the string is shorter than or equal to the limit, the original string is returned unchanged.
func Wrap(str string, limit int, breakWords ...string) string {
	if len(breakWords) == 0 {
		breakWords = []string{"..."}
	}

	if len(str) > limit {
		return str[:limit] + "..."
	}
	return str
}

// RightWrap truncates a string from the left side, keeping the rightmost characters up to the specified limit.
// If no break words are provided, it defaults to using "..." as the truncation indicator.
// If the string is shorter than or equal to the limit, the original string is returned unchanged.
func RightWrap(str string, limit int, breakWords ...string) string {
	if len(breakWords) == 0 {
		breakWords = []string{"..."}
	}

	if len(str) > limit {
		start := len(str) - limit
		return "..." + str[start:]
	}
	return str
}
