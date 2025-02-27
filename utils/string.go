package utils

import "strings"

// CapitalizeWords MAKES THE FIRST LETTER OF EACH WORD IN A STRING UPPERCASE.
func CapitalizeWords(s string) string {
	words := strings.Fields(s)

	for i, word := range words {
		if len(word) > 0 {
			words[i] = strings.ToUpper(word[:1]) + strings.ToLower(word[1:])
		}
	}

	return strings.Join(words, " ")
}
