package utils

import "fmt"

// FormatCount returns a string with the count and the singular or plural form of a word.
func FormatCount(count int, singular, plural string) string {
	if count == 1 {
		return fmt.Sprintf("%d %s", count, singular)
	}

	return fmt.Sprintf("%d %s", count, plural)
}
