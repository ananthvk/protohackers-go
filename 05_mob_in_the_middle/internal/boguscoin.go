package internal

import (
	"unicode"
)

func isAlphaNumeric(c rune) bool {
	return ('a' <= c && c <= 'z') || ('A' <= c && c <= 'Z') || ('0' <= c && c <= '9')
}

func ReplaceBoguscoin(input string, target string) string {
	// Two pointer solution to handle replacement, since splitting by
	// space does not work (when newline / other whitespace characters are present)
	left := 0
	runes := []rune(input)
	targetRunes := []rune(target)
	result := make([]rune, 0, len(runes))
	for right, ch := range runes {
		if !isAlphaNumeric(ch) {
			length := right - left
			if length >= 26 && length <= 35 && runes[left] == '7' && unicode.IsSpace(ch) {
				// Found a boguscoin
				result = append(result, targetRunes...)
			} else {
				result = append(result, runes[left:right]...)
			}
			// Add the non alphanumeric character at right
			result = append(result, runes[right])
			left = right + 1
		}
	}
	// Check if there is a final substring
	if left < len(runes) {
		length := len(runes) - left
		if length >= 26 && length <= 35 && runes[left] == '7' {
			// Found a boguscoin
			result = append(result, targetRunes...)
		} else {
			result = append(result, runes[left:]...)
		}
	}
	return string(result)
}
