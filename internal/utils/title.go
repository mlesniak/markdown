package utils

import "strings"

// AutoCaptialize replaces the beginning of each word in a string with its uppercase pendant.
func AutoCaptialize(title string) string {
	parts := strings.Split(title, " ")
	capitalized := []string{}
	for _, part := range parts {
		t := strings.ToTitle(string(part[0]))
		if len(part) > 1 {
			t = t + part[1:]
		}

		capitalized = append(capitalized, t)
	}
	return strings.Join(capitalized, " ")
}
