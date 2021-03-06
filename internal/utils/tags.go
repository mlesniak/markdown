package utils

import (
	"regexp"
	"strings"
)

// GetTags is defined here to prevent cyclic import.
func GetTags(data []byte) []string {
	markdown := string(data)
	regex := regexp.MustCompile(` *(#\w+)`)
	matches := regex.FindAllString(markdown, -1)

	tags := []string{}

	for _, tag := range matches {
		if tag != "" {
			// Triming is easier than using the matcher's group.
			tag = strings.Trim(tag, " \n\r\t")
			tags = append(tags, tag)
		}
	}

	return tags
}
