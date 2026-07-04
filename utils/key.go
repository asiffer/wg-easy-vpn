package utils

import "fmt"

func checkRune(c rune) bool {
	upper := (c >= 65) && (c <= 90)
	lower := (c >= 97) && (c <= 122)
	number := (c >= 48) && (c <= 57)
	underscore := (c == '_')
	return upper || lower || number || underscore
}

func checkKey(key string) error {
	for _, r := range key {
		if !checkRune(r) {
			return fmt.Errorf("Key %s contains invalid characters ('%c')",
				key, r)
		}
	}
	return nil
}
